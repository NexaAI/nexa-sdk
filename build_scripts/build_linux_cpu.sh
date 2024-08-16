#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Return exit status of the last command in the pipeline that failed.

ROOT_DIR=$(pwd)
CONFIG_FLAG=Release
LIBS_DEST_DIR=${ROOT_DIR}/nexa/gguf/_libs/cpu
BUILD_DIR=build_cpu

# Create the destination directory if it doesn't exist
mkdir -p ${LIBS_DEST_DIR}

# install dependencies from requirements.txt
# pip install -r requirements.txt

# Number of processors for parallel build
NPROC=$(nproc)

# Function to build and copy libraries
build_and_copy_libs() {
    local FOLDER=$1
    local BUILD_FLAGS=$2
    local LIBRARY_PATHS=("${@:3}")

    cd ${FOLDER}
    rm -rf ${BUILD_DIR}
    cmake -B ${BUILD_DIR} ${BUILD_FLAGS}
    cmake --build ${BUILD_DIR} -j ${NPROC} --config ${CONFIG_FLAG}

    # Copy the built libraries to the appropriate directory
    for LIB_PATH in "${LIBRARY_PATHS[@]}"; do
        cp ${LIB_PATH} ${LIBS_DEST_DIR}
    done
}

# Build llama.cpp libraries
LLAMA_FOLDER=${ROOT_DIR}/dependency/llama.cpp
LLAMA_LIB_PATHS=(
    "${LLAMA_FOLDER}/${BUILD_DIR}/src/libllama.so"
    "${LLAMA_FOLDER}/${BUILD_DIR}/ggml/src/libggml.so"
    "${LLAMA_FOLDER}/${BUILD_DIR}/examples/llava/libllava_shared.so"
)
build_and_copy_libs "${LLAMA_FOLDER}" "-DGGML_CUDA=OFF -DGGML_METAL=OFF" "${LLAMA_LIB_PATHS[@]}"

# Build stable-diffusion.cpp libraries
SD_FOLDER=${ROOT_DIR}/dependency/stable-diffusion.cpp
SD_LIB_PATHS=(
    "${SD_FOLDER}/${BUILD_DIR}/bin/libstable-diffusion.so"
)
build_and_copy_libs "${SD_FOLDER}" "-DGGML_OPENBLAS=ON -DSD_CUBLAS=OFF -DSD_BUILD_SHARED_LIBS=ON" "${SD_LIB_PATHS[@]}"