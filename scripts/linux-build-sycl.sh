#!/bin/bash

source /opt/intel/oneapi/setvars.sh   

CMAKE_GENERATOR="Ninja" CMAKE_ARGS="-DGGML_SYCL=ON -DCMAKE_C_COMPILER=cl -DCMAKE_CXX_COMPILER=icx" pip install -e .
# CMAKE_GENERATOR="Ninja" CMAKE_ARGS="-DGGML_SYCL=ON -DCMAKE_C_COMPILER=cl -DCMAKE_CXX_COMPILER=icx" pip install nexaai --prefer-binary --index-url https://github.nexa.ai/whl/sycl --extra-index-url https://pypi.org/simple --no-cache-dir
