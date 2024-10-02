import logging
from typing import Any, Dict
from nexa import __version__

logger = logging.getLogger(__name__)

def add_env_info(storage: Dict[str, Any]):
    nexa_sdk_version = __version__
    added_info = {
        "nexa_sdk_version": nexa_sdk_version
    }
    storage.update(added_info)