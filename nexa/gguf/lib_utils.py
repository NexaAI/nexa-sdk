import ctypes
import logging
import os
import pathlib
import sys
from importlib.metadata import PackageNotFoundError, distribution
from importlib.util import find_spec
from pathlib import Path
from typing import List

from nexa.utils import (
    is_nexa_cuda_installed,
    is_nexa_metal_installed,
)


def is_gpu_available():
    current_dir = os.path.dirname(os.path.abspath(__file__))
    sentinel_file_exists = os.path.exists(os.path.join(current_dir, "lib", "empty_file.txt"))
    return sentinel_file_exists

# Load the library
def load_library(lib_base_name: str):
    # Construct the paths to the possible shared library names
    _base_path = pathlib.Path(os.path.abspath(os.path.dirname(__file__))) / "lib"
    logging.debug(f"Base path for libraries: {_base_path}")
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
            _base_path / f"lib{lib_base_name}.dylib",
            _base_path / f"lib{lib_base_name}.so",
        ]
    elif sys.platform == "win32":
        _lib_paths += [
            _base_path / f"{lib_base_name}.dll",
            _base_path / f"lib{lib_base_name}.dll",
        ]
        _add_windows_dll_directories(_base_path)
    else:
        raise RuntimeError("Unsupported platform")
    logging.debug(f"Possible shared library paths: {_lib_paths}")

    if "LLAMA_CPP_LIB" in os.environ:
        lib_base_name = os.environ["LLAMA_CPP_LIB"]
        _lib = pathlib.Path(lib_base_name)
        _base_path = _lib.parent.resolve()
        _lib_paths = [_lib.resolve()]

    cdll_args = dict()  # type: ignore

    # Try to load the shared library, handling potential errors
    for _lib_path in _lib_paths:
        logging.debug(f"Trying to load shared library from: {_lib_path}")
        if _lib_path.exists():
            try:
                loaded_lib = ctypes.CDLL(str(_lib_path), **cdll_args)  # type: ignore
                logging.debug(f"Successfully loaded shared library: {_lib_path}")
                return loaded_lib
            except Exception as e:
                raise RuntimeError(f"Failed to load shared library '{_lib_path}': {e}")

    raise FileNotFoundError(
        f"Shared library with base name '{lib_base_name}' not found in paths: {_lib_paths}"
    )


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


def try_add_cuda_lib_path():
    """Try to add the CUDA library paths to the system PATH."""
    required_submodules = ["cuda_runtime", "cublas"]
    cuda_versions = ["11", "12"]

    module_spec = find_spec("nvidia")
    if not module_spec:
        return

    nvidia_lib_root = Path(module_spec.submodule_search_locations[0])

    for submodule in required_submodules:
        for ver in cuda_versions:
            try:
                package_name = f"nvidia_{submodule}_cu{ver}"
                _ = distribution(package_name)

                lib_path = nvidia_lib_root / submodule / "bin"
                os.add_dll_directory(str(lib_path))
                os.environ["PATH"] = str(lib_path) + os.pathsep + os.environ["PATH"]
                logging.debug(f"Added {lib_path} to PATH")
            except PackageNotFoundError:
                logging.debug(f"{package_name} not found")
