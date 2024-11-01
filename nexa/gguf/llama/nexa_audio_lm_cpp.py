import ctypes
import os
import sys
import pathlib

# Load the library
def _load_shared_library(lib_base_name: str):
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
    _base_path = pathlib.Path(__file__).parent.parent.resolve()
    _lib_paths = [
        _base_path / f"lib/lib{lib_base_name}{lib_ext}",
        _base_path / f"lib/{lib_base_name}{lib_ext}",
    ]

    if "NEXA_LLAMA_CPP_LIB" in os.environ:
        lib_base_name = os.environ["NEXA_LLAMA_CPP_LIB"]
        _lib = pathlib.Path(lib_base_name)
        _base_path = _lib.parent.resolve()
        _lib_paths = [_lib.resolve()]

    # Add the library directory to the DLL search path on Windows (if needed)
    if sys.platform == "win32" and sys.version_info >= (3, 8):
        os.add_dll_directory(str(_base_path))

    # Try to load the shared library, handling potential errors
    for _lib_path in _lib_paths:
        if _lib_path.exists():
            try:
                return ctypes.CDLL(str(_lib_path))
            except Exception as e:
                raise RuntimeError(f"Failed to load shared library '{_lib_path}': {e}")

    raise FileNotFoundError(
        f"Shared library with base name '{lib_base_name}' not found"
    )

# Load both libraries
_lib_base_name = "hf-omni-audio-cli_shared"
_lib_omni = _load_shared_library(_lib_base_name)
_lib_base_name = "hf-qwen2-audio_shared"
_lib_qwen2 = _load_shared_library(_lib_base_name)


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


def get_lib(is_qwen: bool):
    return _lib_qwen2 if is_qwen else _lib_omni


def context_default_params(is_qwen: bool = False) -> omni_context_params:
    lib = get_lib(is_qwen)
    return lib.omni_context_default_params()


def init_context(params: omni_context_params_p, is_qwen: bool = False) -> omni_context_p:  # type: ignore
    lib = get_lib(is_qwen)
    return lib.omni_init_context(params)


def process_full(ctx: omni_context_p, params: omni_context_params_p, is_qwen: bool = False):  # type: ignore
    lib = get_lib(is_qwen)
    return lib.omni_process_full(ctx, params)


def free_context(ctx: omni_context_p, is_qwen: bool = False):
    lib = get_lib(is_qwen)
    return lib.omni_free(ctx)


# Set up function signatures for both libraries
for lib in [_lib_omni, _lib_qwen2]:
    lib.omni_context_default_params.argtypes = []
    lib.omni_context_default_params.restype = omni_context_params
    
    lib.omni_init_context.argtypes = [omni_context_params_p]
    lib.omni_init_context.restype = omni_context_p
    
    lib.omni_process_full.argtypes = [omni_context_p, omni_context_params_p]
    lib.omni_process_full.restype = None
    
    lib.omni_free.argtypes = [omni_context_p]
    lib.omni_free.restype = None