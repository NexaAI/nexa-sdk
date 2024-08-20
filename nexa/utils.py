import itertools
import os
import sys
import threading
import time
from functools import partial, wraps
from importlib.metadata import PackageNotFoundError, distribution
import platform
from contextlib import redirect_stdout, redirect_stderr

from prompt_toolkit import HTML, prompt
from prompt_toolkit.styles import Style

from nexa.constants import EXIT_COMMANDS, EXIT_REMINDER


def is_package_installed(package_name: str) -> bool:
    """Check if a given backend package is installed."""
    try:
        _ = distribution(f"{package_name}")
        return True
    except PackageNotFoundError:
        return False


def is_nexa_gpu_installed() -> bool:
    """Check if the Nexa GPU package is installed."""
    return is_package_installed("nexaai-gpu")


def is_metal_available():
    arch = platform.machine().lower()
    return sys.platform == "darwin" and ('arm' in arch or 'aarch' in arch) # ARM architecture for Apple Silicon


def is_x86() -> bool:
    """Check if the architecture is x86."""
    return platform.machine().startswith("x86")

def is_arm64() -> bool:
    """Check if the architecture is ARM64."""
    return platform.machine().startswith("arm")


class suppress_stdout_stderr:
    """Context manager to suppress stdout and stderr."""
    def __enter__(self):
        self.null_file = open(os.devnull, "w")
        self.old_stdout = sys.stdout
        self.old_stderr = sys.stderr
        self.stdout_redirect = redirect_stdout(self.null_file)
        self.stderr_redirect = redirect_stderr(self.null_file)
        self.stdout_redirect.__enter__()
        self.stderr_redirect.__enter__()
        return self

    def __exit__(self, *args):
        self.stdout_redirect.__exit__(*args)
        self.stderr_redirect.__exit__(*args)
        sys.stdout = self.old_stdout
        sys.stderr = self.old_stderr
        self.null_file.close()


_style = Style.from_dict({
    "prompt": "ansiblue",
})

_prompt = partial(prompt, ">>> ", style=_style)


def light_text(placeholder):
    """Apply light text style to the placeholder."""
    return HTML(f'<style color="#777777">{placeholder} (type "/exit" to quit)</style>')


def nexa_prompt(placeholder: str = "Send a message ...") -> str:
    """Display a prompt to the user and handle input."""
    try:
        user_input = _prompt(placeholder=light_text(placeholder)).strip()

        # Clear the placeholder if the user pressed Enter without typing
        if user_input == placeholder:
            user_input = ""

        if user_input.lower() in EXIT_COMMANDS:
            print("Exiting...")
            exit(0)
        return user_input
    except KeyboardInterrupt:
        print(EXIT_REMINDER)
        return
    except EOFError:
        print("Exiting...")

    exit(0)


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
    def __init__(self):
        frames = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
        self.spinner = itertools.cycle(frames)
        self.stop_spinning = threading.Event()

    def _spin(self):
        while not self.stop_spinning.is_set():
            sys.stdout.write(f"\r{next(self.spinner)} ")
            sys.stdout.flush()
            time.sleep(0.1)
            if self.stop_spinning.is_set():
                break

    def __enter__(self):
        self.thread = threading.Thread(target=self._spin)
        self.thread.start()
        return self

    def __exit__(self, exc_type, exc_value, traceback):
        self.stop_spinning.set()
        self.thread.join()
        sys.stdout.write("\r")
        sys.stdout.flush()

    def __call__(self, func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            with self:
                return func(*args, **kwargs)

        return wrapper
