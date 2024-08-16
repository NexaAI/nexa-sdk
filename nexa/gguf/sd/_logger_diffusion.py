import sys
import ctypes
import logging

import nexa.gguf.sd.stable_diffusion_cpp as stable_diffusion_cpp

# enum sd_log_level_t {
#     SD_LOG_DEBUG = 0,
#     SD_LOG_INFO = 1,
#     SD_LOG_WARN = 2,
#     SD_LOG_ERROR = 3
# };
SD_LOG_LEVEL_TO_LOGGING_LEVEL = {
    0: logging.ERROR,
    1: logging.WARNING,
    2: logging.INFO,
    3: logging.DEBUG,
}

logger = logging.getLogger("nexa-stable-diffusion")


# @stable_diffusion_cpp.sd_log_callback
# def sd_log_callback(
#     level: int,
#     text: bytes,
#     data: ctypes.c_void_p,
# ):
#     if logger.level <= SD_LOG_LEVEL_TO_LOGGING_LEVEL[level]:
#         print(text.decode("utf-8"), end="", flush=True, file=sys.stderr)


# stable_diffusion_cpp.sd_set_log_callback(sd_log_callback, ctypes.c_void_p(0))


# def set_verbose(verbose: bool):
#     logger.setLevel(logging.DEBUG if verbose else logging.ERROR)


# def log_event(level: int, message: str):
#     if logger.level <= SD_LOG_LEVEL_TO_LOGGING_LEVEL[level]:
#         print(message)