import logging
import logging.config
from typing import Optional


def setup_logging(
    level: str = "INFO",
    to_file: bool = False,
    use_colorlog: bool = True,
    prefix: Optional[str] = None,
    disable_existing_loggers: bool = False,
):
    # base logging config
    logging_config = {
        "version": 1,
        "handlers": {
            "console": {"formatter": "simple", "stream": "ext://sys.stdout", "class": "logging.StreamHandler"},
        },
        "root": {"level": level, "handlers": ["console"]},
        "disable_existing_loggers": disable_existing_loggers,
    }

    # formatters
    logging_config["formatters"] = {
        "simple": {"format": "[%(asctime)s][%(name)s][%(levelname)s] - %(message)s"},
    }

    # add file handler
    if to_file:
        logging_config["handlers"]["file"] = {
            "formatter": "simple",
            "filename": "benchmark.log",
            "class": "logging.FileHandler",
        }
        logging_config["root"]["handlers"].append("file")

    # use colorlog
    if use_colorlog:
        logging_config["formatters"]["colorlog"] = {
            "()": "colorlog.ColoredFormatter",
            "format": "[%(cyan)s%(asctime)s%(reset)s][%(blue)s%(name)s%(reset)s][%(log_color)s%(levelname)s%(reset)s] - %(message)s",
            "log_colors": {"DEBUG": "purple", "INFO": "green", "WARNING": "yellow", "CRITICAL": "red", "ERROR": "red"},
        }
        for handler in logging_config["handlers"]:
            logging_config["handlers"][handler]["formatter"] = "colorlog"

    # format prefix
    if prefix is not None:
        for formatter in logging_config["formatters"]:
            logging_config["formatters"][formatter]["format"] = (
                f"[{prefix}]" + logging_config["formatters"][formatter]["format"]
            )

    logging.config.dictConfig(logging_config)
