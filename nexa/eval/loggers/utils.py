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


def add_tokenizer_info(storage: Dict[str, Any], lm):
    if getattr(lm, "tokenizer", False):
        try:
            tokenizer_info = {
                "tokenizer_pad_token": [
                    lm.tokenizer.pad_token,
                    str(lm.tokenizer.pad_token_id),
                ],
                "tokenizer_eos_token": [
                    lm.tokenizer.eos_token,
                    str(lm.tokenizer.eos_token_id),
                ],
                "tokenizer_bos_token": [
                    lm.tokenizer.bos_token,
                    str(lm.tokenizer.bos_token_id),
                ],
                "eot_token_id": getattr(lm, "eot_token_id", None),
                "max_length": getattr(lm, "max_length", None),
            }
            storage.update(tokenizer_info)
        except Exception as err:
            logger.debug(
                f"Logging detailed tokenizer info failed with {err}, skipping..."
            )
        # seems gguf and textsynth do not have tokenizer
    else:
        logger.debug(
            "LM does not have a 'tokenizer' attribute, not logging tokenizer metadata to results."
        )
