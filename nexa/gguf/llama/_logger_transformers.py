import sys
import ctypes
import logging

import nexa.gguf.llama as llama_cpp

# enum ggml_log_level {
#     GGML_LOG_LEVEL_NONE  = 0,
#     GGML_LOG_LEVEL_INFO  = 1,
#     GGML_LOG_LEVEL_WARN  = 2,
#     GGML_LOG_LEVEL_ERROR = 3,
#     GGML_LOG_LEVEL_DEBUG = 4,
#     GGML_LOG_LEVEL_CONT  = 5, // continue previous log
# };
GGML_LOG_LEVEL_TO_LOGGING_LEVEL = {
    0: logging.CRITICAL,
    1: logging.INFO,
    2: logging.WARNING,
    3: logging.ERROR,
    4: logging.DEBUG,
    5: logging.DEBUG,
}
# Mapping ggml log levels to Python logging levels
GGML_LOG_LEVEL_TO_LOGGING_LEVEL = {
    2: logging.ERROR,
    3: logging.WARNING,
    4: logging.INFO,
    5: logging.DEBUG,
}

# Initialize the logger for llama-cpp-python
logger = logging.getLogger("nexa-transformers")

# Utility function to set verbosity
def set_verbose(verbose: bool):
    logger.setLevel(logging.DEBUG if verbose else logging.ERROR)

# Example usage
if __name__ == "__main__":
    # Set the verbosity based on a condition or user input
    set_verbose(False)
    # Rest of your application code here
