import importlib.metadata
import importlib.util

_pynvml_available = importlib.util.find_spec("pynvml") is not None
_codecarbon_available = importlib.util.find_spec("codecarbon") is not None
_amdsmi_available = importlib.util.find_spec("amdsmi") is not None
_psutil_available = importlib.util.find_spec("psutil") is not None
_pyrsmi_available = importlib.util.find_spec("pyrsmi") is not None
_nexa_sdk_available = importlib.util.find_spec("nexaai") is not None


def is_pyrsmi_available():
    return _pyrsmi_available

def is_psutil_available():
    return _psutil_available

def is_pynvml_available():
    return _pynvml_available

def is_amdsmi_available():
    return _amdsmi_available

def is_codecarbon_available():
    return _codecarbon_available

def nexa_sdk_version():
    if _nexa_sdk_available:
        return importlib.metadata.version("nexaai")

