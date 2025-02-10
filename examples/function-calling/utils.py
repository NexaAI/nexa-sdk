import os
import sys
import requests

outnull_file = open(os.devnull, "w")
errnull_file = open(os.devnull, "w")

STDOUT_FILENO = 1
STDERR_FILENO = 2


class suppress_stdout_stderr(object):
    # NOTE: these must be "saved" here to avoid exceptions when using
    #       this context manager inside of a __del__ method
    sys = sys
    os = os

    def __init__(self, disable: bool = True):
        self.disable = disable

    # Oddly enough this works better than the contextlib version
    def __enter__(self):
        if self.disable:
            return self

        self.old_stdout_fileno_undup = STDOUT_FILENO
        self.old_stderr_fileno_undup = STDERR_FILENO

        self.old_stdout_fileno = self.os.dup(self.old_stdout_fileno_undup)
        self.old_stderr_fileno = self.os.dup(self.old_stderr_fileno_undup)

        self.old_stdout = self.sys.stdout
        self.old_stderr = self.sys.stderr

        self.os.dup2(outnull_file.fileno(), self.old_stdout_fileno_undup)
        self.os.dup2(errnull_file.fileno(), self.old_stderr_fileno_undup)

        self.sys.stdout = outnull_file
        self.sys.stderr = errnull_file
        return self

    def __exit__(self, *_):
        if self.disable:
            return

        # Check if sys.stdout and sys.stderr have fileno method
        self.sys.stdout = self.old_stdout
        self.sys.stderr = self.old_stderr

        self.os.dup2(self.old_stdout_fileno, self.old_stdout_fileno_undup)
        self.os.dup2(self.old_stderr_fileno, self.old_stderr_fileno_undup)

        self.os.close(self.old_stdout_fileno)
        self.os.close(self.old_stderr_fileno)


def call_function(func_name: str, *args, **kwargs):
    """
    Calls a function by its name from the global scope.

    Parameters:
    func_name (str): The name of the function as a string.
    *args: Positional arguments to pass to the function.
    **kwargs: Keyword arguments to pass to the function.

    Returns:
    The return value of the called function.

    Raises:
    AttributeError: If the function is not found or is not callable.
    """
    # Retrieve the function object from global variables
    func = globals().get(func_name)

    # Check if function exists and is callable
    if not callable(func):
        raise AttributeError(
            f"Function '{func_name}' not found or is not callable.")

    # Call the function with arguments
    return func(*args, **kwargs)


def add_integer(a: int, b: int):
    return a + b


# Fetch real-time weather data for a city using wttr.in, a simple console-based weather API.
def get_weather(city: str) -> str:
    city = city.strip()
    url = f"https://wttr.in/{city}"
    response = requests.get(url, timeout=5)
    try:
        response = requests.get(url, timeout=5)
        response.raise_for_status()
        # Only keep ASCII graphical outputs
        weather_output = '\n'.join(response.text.split('\n')[:-1])
        return weather_output
    except requests.exceptions.RequestException as e:
        print(f"Error fetching weather data: {e}")
        return f'Weather of {city} not found.'


system_prompt = (
    "You are an AI assistant that generates structured function calling responses. "
    "Identify the correct function from the available tools and return a JSON object "
    "containing the function name and all required parameters. Ensure the parameters "
    "are accurately derived from the user's input and formatted correctly."
)
