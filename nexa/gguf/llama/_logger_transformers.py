import sys
import ctypes
import logging

from nexa.gguf.llama import llama_cpp

# Mapping ggml log levels to Python logging levels
GGML_LOG_LEVEL_TO_LOGGING_LEVEL = {
    2: logging.ERROR,
    3: logging.WARNING,
    4: logging.INFO,
    5: logging.DEBUG,
}

# Initialize the logger for llama-cpp-python
logger = logging.getLogger("nexa-transformers")

# # Define the log callback function
# @llama_cpp.llama_log_callback
# def llama_log_callback(
#     level: int,
#     text: bytes,
#     user_data: ctypes.c_void_p,
# ):
#     # Check if the logger is set to log the provided level
#     if logger.level <= GGML_LOG_LEVEL_TO_LOGGING_LEVEL[level]:
#         # Print the log message to stderr
#         print(text.decode("utf-8"), end="", flush=True, file=sys.stderr)

# # Set the log callback function for llama_cpp
# llama_cpp.llama_log_set(llama_log_callback, ctypes.c_void_p(0))

# Utility function to set verbosity
def set_verbose(verbose: bool):
    logger.setLevel(logging.DEBUG if verbose else logging.ERROR)

# Example usage
if __name__ == "__main__":
    # Set the verbosity based on a condition or user input
    set_verbose(False)
    # Rest of your application code here
