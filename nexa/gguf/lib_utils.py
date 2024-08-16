import ctypes
import logging
import os
import sys
from pathlib import Path
from typing import List, Optional

from nexa.utils import (
    is_nexa_cuda_installed,
    is_nexa_metal_installed,
    try_add_cuda_lib_path,
)


# Determine the appropriate library folder
def _determine_lib_folder(lib_folder: Optional[str]) -> str:
    if lib_folder:
        valid_lib_folders = ["cpu", "cuda", "metal"]
        assert (
            lib_folder in valid_lib_folders
        ), f"lib_folder must be one of {valid_lib_folders}"
        return lib_folder

    if is_nexa_cuda_installed():
        return "cuda"
    elif is_nexa_metal_installed():
        return "metal"
    return "cpu"


# Load the shared library
def _load_shared_library(
    lib_base_name: str, lib_folder: Optional[str] = None
) -> ctypes.CDLL:
    assert (
        sys.platform in ["darwin", "win32"]
        or sys.platform.startswith("linux")
        or sys.platform.startswith("freebsd")
    ), "Unsupported platform"

    lib_folder = _determine_lib_folder(lib_folder)
    _base_path = Path(os.path.dirname(__file__)).resolve() / "_libs" / lib_folder
    cdll_args = {}

    if sys.platform == "win32":
        cdll_args["winmode"] = ctypes.RTLD_GLOBAL
        _add_windows_dll_directories(_base_path)

    _lib_paths: List[Path] = []

    if "LLAMA_CPP_LIB" in os.environ:
        lib_base_name = os.environ["LLAMA_CPP_LIB"]
        _lib = Path(lib_base_name)
        _base_path = _lib.parent.resolve()
        _lib_paths += [_lib.resolve()]

    prefix = "" if sys.platform == "win32" else "lib"

    ext = {
        # "linux": "so",
        # "freebsd": "so",
        "darwin": "dylib",
        "win32": "dll",
    }.get(sys.platform, "so")

    _lib_paths += [_base_path / f"{prefix}{lib_base_name}.{ext}"]

    # Manually load platform-specific ggml library
    ctypes.CDLL(str(_base_path / f"{prefix}ggml.{ext}"), **cdll_args)

    # Try to load the shared library, handling potential errors
    for _lib_path in _lib_paths:
        if _lib_path.exists():
            try:
                return ctypes.CDLL(str(_lib_path), **cdll_args)
            except Exception as e:
                raise RuntimeError(f"Failed to load shared library '{_lib_path}': {e}")

    raise FileNotFoundError(
        f"Shared library with base name '{lib_base_name}' not found"
    )


# Main logic to load the library
def load_library(_lib_base_name: str) -> ctypes.CDLL:
    try:
        return _load_shared_library(_lib_base_name)
    except (FileNotFoundError, RuntimeError) as e:
        logging.warning(
            f"Failed to load shared library '{_lib_base_name}' with error:"
            f"\n{e}"
            "\nFalling back to CPU version."
        )
        return _load_shared_library(_lib_base_name, "cpu")


def _add_windows_dll_directories(base_path: Path) -> None:
    try_add_cuda_lib_path()
    os.add_dll_directory(str(base_path))
    os.environ["PATH"] = str(base_path) + os.pathsep + os.environ["PATH"]

    if sys.version_info >= (3, 8):
        if "CUDA_PATH" in os.environ:
            os.add_dll_directory(os.path.join(os.environ["CUDA_PATH"], "bin"))
            os.add_dll_directory(os.path.join(os.environ["CUDA_PATH"], "lib"))
        if "HIP_PATH" in os.environ:
            os.add_dll_directory(os.path.join(os.environ["HIP_PATH"], "bin"))
            os.add_dll_directory(os.path.join(os.environ["HIP_PATH"], "lib"))
