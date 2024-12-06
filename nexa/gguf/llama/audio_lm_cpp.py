import ctypes
import os
import sys
from pathlib import Path


# Load the library
def _load_shared_library(lib_base_name: str, base_path: Path = None):
    # Determine the file extension based on the platform
    if sys.platform.startswith("linux"):
        lib_ext = ".so"
    elif sys.platform == "darwin":
        lib_ext = ".dylib"
    elif sys.platform == "win32":
        lib_ext = ".dll"
    else:
        raise RuntimeError("Unsupported platform")
    # Construct the paths to the possible shared library names
    if base_path is None:
        _base_path = Path(__file__).parent.parent.resolve()
    else:
        _base_path = base_path
    _lib_paths = [
        _base_path / f"lib{lib_base_name}{lib_ext}",
        _base_path / f"{lib_base_name}{lib_ext}",
    ]
    # Add the library directory to the DLL search path on Windows (if needed)
    if sys.platform == "win32" and sys.version_info >= (3, 8):
        os.add_dll_directory(str(_base_path))
    # Try to load the shared library, handling potential errors
    for _lib_path in _lib_paths:
        if _lib_path.exists():
            try:
                return ctypes.CDLL(str(_lib_path))
            except Exception as e:
                print(f"Failed to load shared library '{_lib_path}': {e}")
    raise FileNotFoundError(
        f"Shared library with base name '{lib_base_name}' not found"
    )

def _get_lib(is_qwen: bool = True):
    # Specify the base name of the shared library to load
    _lib_base_name = "nexa-qwen2-audio-lib_shared" if is_qwen else "nexa-omni-audio-lib_shared"
    base_path = (
        Path(__file__).parent.parent.parent.parent.resolve()
        / "nexa"
        / "gguf"
        / "lib"
    )
    return _load_shared_library(_lib_base_name, base_path)

# Initialize both libraries
_lib_omni = _get_lib(is_qwen=False)
_lib_qwen = _get_lib(is_qwen=True)

#   conda config --add channels conda-forge
#   conda update libstdcxx-ng
# struct omni_context_params
# {
#     char *model;
#     char *mmproj;
#     char *file;
#     char *prompt;
#     int32_t n_gpu_layers;
# };
class omni_context_params(ctypes.Structure):
    _fields_ = [
        ("model", ctypes.c_char_p),
        ("mmproj", ctypes.c_char_p),
        ("file", ctypes.c_char_p),
        ("prompt", ctypes.c_char_p),
        ("n_gpu_layers", ctypes.c_int32),
    ]

omni_context_params_p = ctypes.POINTER(omni_context_params)
omni_context_p = ctypes.c_void_p

# OMNI_AUDIO_API omni_context_params omni_context_default_params();
def context_default_params(is_qwen: bool = True) -> omni_context_params:
    _lib = _lib_qwen if is_qwen else _lib_omni
    return _lib.omni_context_default_params()

# OMNI_AUDIO_API struct omni_context *omni_init_context(omni_context_params &params);
def init_context(params: omni_context_params_p, is_qwen: bool = True) -> omni_context_p:  # type: ignore
    _lib = _lib_qwen if is_qwen else _lib_omni
    return _lib.omni_init_context(params)

# OMNI_AUDIO_API void omni_process_full(
#     struct omni_context *ctx_omni,
#     omni_context_params &params
# );
def process_full(ctx: omni_context_p, params: omni_context_params_p, is_qwen: bool = True):  # type: ignore
    _lib = _lib_qwen if is_qwen else _lib_omni
    return _lib.omni_process_full(ctx, params)

# OMNI_AUDIO_API void omni_free(struct omni_context *ctx_omni);
def free(ctx: omni_context_p, is_qwen: bool = True):
    _lib = _lib_qwen if is_qwen else _lib_omni
    return _lib.omni_free(ctx)

for lib in [_lib_omni, _lib_qwen]:
    # Configure context_default_params
    lib.omni_context_default_params.argtypes = []
    lib.omni_context_default_params.restype = omni_context_params

    # Configure init_context
    lib.omni_init_context.argtypes = [omni_context_params_p]
    lib.omni_init_context.restype = omni_context_p

    # Configure process_full
    lib.omni_process_full.argtypes = [omni_context_p, omni_context_params_p]
    lib.omni_process_full.restype = ctypes.c_char_p

    # Configure free
    lib.omni_free.argtypes = [omni_context_p]
    lib.omni_free.restype = None
