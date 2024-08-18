import ctypes
import logging
import os
import sys
from pathlib import Path
from typing import List, Optional
import pathlib
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

# Load the library
def _load_shared_library(lib_base_name: str):
    # Construct the paths to the possible shared library names
    _base_path = pathlib.Path(os.path.abspath(os.path.dirname(__file__))) / "lib"
    print(f"Base path for libraries: {_base_path}")
    # Searching for the library in the current directory under the name "libllama" (default name
    # for llamacpp) and "llama" (default name for this repo)
    _lib_paths: List[pathlib.Path] = []
    # Determine the file extension based on the platform
    if sys.platform.startswith("linux") or sys.platform.startswith("freebsd"):
        _lib_paths += [
            _base_path / f"lib{lib_base_name}.so",
        ]
    elif sys.platform == "darwin":
        _lib_paths += [
            _base_path / f"lib{lib_base_name}.so",
            _base_path / f"lib{lib_base_name}.dylib",
        ]
    elif sys.platform == "win32":
        _lib_paths += [
            _base_path / f"{lib_base_name}.dll",
            _base_path / f"lib{lib_base_name}.dll",
        ]
    else:
        raise RuntimeError("Unsupported platform")
    print(f"Possible shared library paths: {_lib_paths}")

    if "LLAMA_CPP_LIB" in os.environ:
        lib_base_name = os.environ["LLAMA_CPP_LIB"]
        _lib = pathlib.Path(lib_base_name)
        _base_path = _lib.parent.resolve()
        _lib_paths = [_lib.resolve()]

    cdll_args = dict()  # type: ignore

    # Add the library directory to the DLL search path on Windows (if needed)
    if sys.platform == "win32":
        os.add_dll_directory(str(_base_path))
        os.environ["PATH"] = str(_base_path) + os.pathsep + os.environ["PATH"]

    if sys.platform == "win32" and sys.version_info >= (3, 8):
        os.add_dll_directory(str(_base_path))
        if "CUDA_PATH" in os.environ:
            os.add_dll_directory(os.path.join(os.environ["CUDA_PATH"], "bin"))
            os.add_dll_directory(os.path.join(os.environ["CUDA_PATH"], "lib"))
        if "HIP_PATH" in os.environ:
            os.add_dll_directory(os.path.join(os.environ["HIP_PATH"], "bin"))
            os.add_dll_directory(os.path.join(os.environ["HIP_PATH"], "lib"))
        cdll_args["winmode"] = ctypes.RTLD_GLOBAL

    # Try to load the shared library, handling potential errors
    for _lib_path in _lib_paths:
        print(f"Trying to load shared library from: {_lib_path}")
        if _lib_path.exists():
            try:
                loaded_lib = ctypes.CDLL(str(_lib_path), **cdll_args)  # type: ignore
                print(f"Successfully loaded shared library: {_lib_path}")
                return loaded_lib
            except Exception as e:
                print(f"Error loading shared library '{_lib_path}': {e}")
                raise RuntimeError(f"Failed to load shared library '{_lib_path}': {e}")

    raise FileNotFoundError(
        f"Shared library with base name '{lib_base_name}' not found in paths: {_lib_paths}"
    )



# Main logic to load the library
def load_library(_lib_base_name: str) -> ctypes.CDLL:
    return _load_shared_library(_lib_base_name)


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
