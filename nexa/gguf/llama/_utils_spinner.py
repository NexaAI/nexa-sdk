import sys
import threading
import time
import os

def get_spinner_style(style="default"):
    spinners = {
        "default": '⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏'
    }
    return spinners.get(style, spinners["default"])

def spinning_cursor(style="default"):
    while True:
        for cursor in get_spinner_style(style):
            yield cursor

def show_spinner(stop_event, style="default", message=""):
    spinner = spinning_cursor(style)
    
    fd = os.open('/dev/tty', os.O_WRONLY)
    
    while not stop_event.is_set():
        display = f"\r{message} {next(spinner)}" if message else f"\r{next(spinner)}"
        os.write(fd, display.encode())
        time.sleep(0.1)
    
    os.write(fd, b"\r" + b" " * (len(message) + 2))  
    os.write(fd, b"\r")
    os.close(fd)

def start_spinner(style="default", message=""):
    stop_event = threading.Event()
    spinner_thread = threading.Thread(target=show_spinner, args=(stop_event, style, message))
    spinner_thread.daemon = True
    spinner_thread.start()
    return stop_event, spinner_thread

def stop_spinner(stop_event, spinner_thread):
    stop_event.set()
    spinner_thread.join()