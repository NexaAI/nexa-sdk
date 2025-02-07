import requests


def call_function(func, *args, **kwargs):
    """
    Calls the given function with the provided positional and keyword arguments.

    Parameters:
    func (callable): The function to be called.
    *args: Variable-length positional arguments to pass to the function.
    **kwargs: Variable-length keyword arguments to pass to the function.

    Returns:
    The return value of the called function.
    """
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
