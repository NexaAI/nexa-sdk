"""Internal module use at your own risk

This module provides a minimal interface for working with ggml tensors from llama-cpp-python
"""
import os
import pathlib

from nexa.gguf.lib_utils import load_library

libggml = load_library("ggml")

