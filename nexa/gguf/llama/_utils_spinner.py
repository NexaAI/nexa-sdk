# For similar spinner animation implementation, refer to: nexa/utils.py

import sys
import threading
import time
import os
import itertools
from contextlib import contextmanager

def get_spinner_style(style="default"):
    spinners = {
        "default": ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"]
    }
    return spinners.get(style, spinners["default"])

def _get_output_stream():
    """Get the appropriate output stream based on platform."""
    if sys.platform == "win32":
        return open('CONOUT$', 'wb')
    else:
        try:
            return os.open('/dev/tty', os.O_WRONLY)
        except (FileNotFoundError, OSError):
            return os.open('/dev/stdout', os.O_WRONLY)

def show_spinner(stop_event, style="default", message=""):
    spinner = itertools.cycle(get_spinner_style(style))
    fd = _get_output_stream()
    is_windows = sys.platform == "win32"
    
    try:
        while not stop_event.is_set():
            display = f"\r{message} {next(spinner)}" if message else f"\r{next(spinner)} "
            
            if is_windows:
                fd.write(display.encode())
                fd.flush()
            else:
                os.write(fd, display.encode())
            time.sleep(0.1)
            
        # Clear the spinner
        clear_msg = b"\r" + b" " * (len(message) + 2) + b"\r"
        if is_windows:
            fd.write(clear_msg)
            fd.flush()
        else:
            os.write(fd, clear_msg)
            
    finally:
        if is_windows:
            fd.close()
        else:
            os.close(fd)

def start_spinner(style="default", message=""):
    stop_event = threading.Event()
    spinner_thread = threading.Thread(
        target=show_spinner, 
        args=(stop_event, style, message),
        daemon=True
    )
    spinner_thread.start()
    return stop_event, spinner_thread

def stop_spinner(stop_event, spinner_thread):
    if stop_event and not stop_event.is_set():
        stop_event.set()
    if spinner_thread and spinner_thread.is_alive():
        spinner_thread.join()

@contextmanager
def spinning_cursor(message="", style="default"):
    """Context manager for spinner animation."""
    stop_event, thread = start_spinner(style, message)
    try:
        yield
    finally:
        stop_spinner(stop_event, thread)