import requests


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
        raise AttributeError(f"Function '{func_name}' not found or is not callable.")

    # Call the function with arguments
    return func(*args, **kwargs)


def add_integer(a: int, b: int):
    return a + b


# Fetch real-time weather data for a city using wttr.in, a simple console-based weather API.
def get_weather(city: str) -> str:
    url = f"https://wttr.in/{city}?format=%t"
    response = requests.get(url)

    if response.status_code == 200:
        return response.text.strip()
    else:
        return "City not found"


system_prompt = (
    "You are an AI assistant that generates structured function calling responses. "
    "Identify the correct function from the available tools and return a JSON object "
    "containing the function name and all required parameters. Ensure the parameters "
    "are accurately derived from the user's input and formatted correctly."
)
