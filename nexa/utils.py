import itertools
import platform
import sys
import threading
import time
from functools import partial, wraps
from importlib.metadata import PackageNotFoundError, distribution
from typing import Dict, List
import json
import logging
import streamlit as st
from nexa.constants import (
    EXIT_COMMANDS,
    EXIT_REMINDER,
    NEXA_MODEL_LIST_PATH,
)


def get_available_models() -> Dict[str, dict]:
    """Get list of available computer vision (cv) models from the model list JSON file."""
    # check whether the model list file exists:
    if not NEXA_MODEL_LIST_PATH.exists():
        st.error("Model list file not found")
        return {}  # empty dict

    try:
        # read model list from the JSON file:
        with open(NEXA_MODEL_LIST_PATH, "r") as f:
            available_models = json.load(f)
            return available_models

    except json.JSONDecodeError as e:
        logging.error(f"Invalid JSON in model list file: {e}")
        return {}
    except Exception as e:
        logging.error(f"Error loading available models: {e}")
        return {}


def filter_available_models(
    models: Dict[str, dict],
    specified_run_type: str,
    model_map: Dict[str, str]
) -> List[str]:
    """Filter available models by run type and apply model mapping."""
    if not models:
        return []

    filtered_models = set()  # to avoid duplicates

    for model_name, model_info in models.items():
        # skip if run_type doesn't match:
        if model_info.get('run_type') != specified_run_type or 'projector' in model_name:
            continue

        if model_name in model_map.values():
            # find short form from mapping:
            for short_name, full_name in model_map.items():
                if full_name == model_name:
                    filtered_models.add(short_name)
                    break
        else:
            filtered_models.add(model_name)

    return sorted(list(filtered_models))


def get_model_options(
    specified_run_type: str,
    model_map: Dict[str, str]
) -> List[str]:
    """Get list of model options including special options."""
    available_models = get_available_models()
    models_list = filter_available_models(available_models, specified_run_type, model_map)
    # add special options at the end of the dropdown menu:
    models_list.extend(["Use Model From Nexa Model Hub 🔍", "Local Model 📁"])
    return models_list


def update_model_options(
    specified_run_type: str,
    model_map: Dict[str, str]
) -> None:
    """Update the model options in session state and force a refresh."""
    try:
        fresh_options = get_model_options(specified_run_type, model_map)
        st.session_state.model_options = fresh_options  # update session state with new options

        if hasattr(st.session_state, 'current_model_path') and st.session_state.current_model_path:
            if st.session_state.current_model_path in fresh_options:
                st.session_state.current_model_index = fresh_options.index(st.session_state.current_model_path)
            else:
                # if current model not in list, reset to Model Hub option:
                hub_index = fresh_options.index("Use Model From Nexa Model Hub 🔍")
                st.session_state.current_model_index = hub_index

    except Exception as e:
        logging.error(f"Error updating model options: {e}")


def is_package_installed(package_name: str) -> bool:
    """Check if a given backend package is installed."""
    try:
        _ = distribution(f"{package_name}")
        return True
    except PackageNotFoundError:
        return False


def is_nexa_cuda_installed() -> bool:
    """Check if the Nexa CUDA package is installed."""
    if is_package_installed("nexaai-cuda"):
        print("Nexa CUDA package is installed.")
        return True
    return False


def is_nexa_metal_installed() -> bool:
    """Check if the Nexa Metal package is installed."""
    if is_package_installed("nexaai-metal"):
        print("Nexa Metal package is installed.")
        return True
    return False


def is_metal_available():
    arch = platform.machine().lower()
    return sys.platform == "darwin" and (
        "arm" in arch or "aarch" in arch
    )  # ARM architecture for Apple Silicon


def is_x86() -> bool:
    """Check if the architecture is x86."""
    return platform.machine().startswith("x86")


def is_arm64() -> bool:
    """Check if the architecture is ARM64."""
    return platform.machine().startswith("arm")

# For prompt input based on the platform
if sys.platform == "win32":
    import msvcrt
else:
    from prompt_toolkit import prompt, HTML
    from prompt_toolkit.styles import Style

    _style = Style.from_dict(
        {
            "prompt": "ansiblue",
        }
    )

    _prompt = partial(prompt, ">>> ", style=_style)

def light_text(placeholder):
    """Apply light text style to the placeholder."""
    if sys.platform == "win32":
        return f"\033[90m{placeholder} (type \"/exit\" to quit)\033[0m"
    else:
        return HTML(f'<style color="#777777">{placeholder} (type "/exit" to quit)</style>')

def strip_ansi(text):
    """Remove ANSI escape codes from a string."""
    import re
    ansi_escape = re.compile(r'\x1B(?:[@-Z\\-_]|\[[0-?]*[ -/]*[@-~])')
    return ansi_escape.sub('', text)

def nexa_prompt(placeholder: str = "Send a message ...") -> str:
    """Display a prompt to the user and handle input."""
    if sys.platform == "win32":
        try:
            hint = light_text(placeholder)
            hint_length = len(strip_ansi(hint))

            # Print the prompt with placeholder
            print(f">>> {hint}", end='', flush=True)

            # Move cursor back to the start of the line
            print('\r', end='', flush=True)
            print(">>> ", end='', flush=True)

            user_input = ""
            while True:
                char = msvcrt.getch().decode()
                if char == '\r':  # Enter key
                    break
                elif char == '\x03':  # Ctrl+C
                    raise KeyboardInterrupt
                elif char == '\x04':  # Ctrl+D (EOF)
                    raise EOFError
                elif char in ('\x08', '\x7f'):  # Backspace
                    if user_input:
                        user_input = user_input[:-1]
                        print('\b \b', end='', flush=True)
                else:
                    user_input += char
                    print(char, end='', flush=True)

                if len(user_input) == 1:  # Clear hint after first character
                    print('\r' + ' ' * (hint_length + 4), end='', flush=True)
                    print(f'\r>>> {user_input}', end='', flush=True)

            print()  # New line after Enter

            if user_input.lower().strip() in EXIT_COMMANDS:
                print("Exiting...")
                sys.exit(0)
            return user_input.strip()
        except KeyboardInterrupt:
            print(EXIT_REMINDER)
            return
        except EOFError:
            print("Exiting...")
            sys.exit(0)
    else:
        try:
            user_input = _prompt(placeholder=light_text(placeholder)).strip()

            # Clear the placeholder if the user pressed Enter without typing
            if user_input == placeholder:
                user_input = ""

            if user_input.lower() in EXIT_COMMANDS:
                print("Exiting...")
                sys.exit(0)
            return user_input
        except KeyboardInterrupt:
            print(EXIT_REMINDER)
            return
        except EOFError:
            print("Exiting...")

    sys.exit(0)


class SpinningCursorAnimation:
    """
    import time

    # Example usage as a context manager

    with SpinningCursorAnimation():
        time.sleep(5)  # Simulate a long-running task

    # Example usage as a decorator
    class MyClass:

    def __init__(self) -> None:
        self._load_model()

    @SpinningCursorAnimation()
    def _load_model(self):
        time.sleep(5)  # Simulate loading process

    obj = MyClass()
    """

    def __init__(self, alternate_stream: bool = True):
        frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
        self.spinner = itertools.cycle(frames)
        self.stop_spinning = threading.Event()
        self._use_alternate_stream = alternate_stream
        self.stream = sys.stdout

    def _spin(self):
        while not self.stop_spinning.is_set():
            self.stream.write(f"\r{next(self.spinner)} ")
            self.stream.flush()
            time.sleep(0.1)
            if self.stop_spinning.is_set():
                break

    def __enter__(self):
        if self._use_alternate_stream:
            if sys.platform == "win32":  # Windows
                self.stream = open('CONOUT$', "w")
            else:
                try:
                    self.stream = open('/dev/tty', "w")
                except (FileNotFoundError, OSError):
                    self.stream = open('/dev/stdout', "w")
        self.thread = threading.Thread(target=self._spin)
        self.thread.start()
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.stop_spinning.set()
        self.thread.join()
        self.stream.write("\r")
        self.stream.flush()
        if self._use_alternate_stream:
            self.stream.close()

    def __call__(self, func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            with self:
                return func(*args, **kwargs)

        return wrapper
