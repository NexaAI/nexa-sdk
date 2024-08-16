#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Return exit status of the last command in the pipeline that failed.

ROOT_DIR=$(pwd)
NPROC=$(sysctl -n hw.ncpu)
CONFIG_FLAG=Release
LIBS_DEST_DIR=${ROOT_DIR}/nexa/gguf/_libs/metal
BUILD_DIR=build_metal

# Create the destination directory if it doesn't exist
mkdir -p ${LIBS_DEST_DIR}

# Install dependencies from requirements.txt
# pip install -r requirements.txt

# Function to build and copy libraries
build_and_copy_libs() {
    local FOLDER=$1
    local BUILD_FLAGS=$2
    local LIBRARY_PATHS=("${@:3}")

    cd ${FOLDER}
    rm -rf ${BUILD_DIR}
    cmake -B ${BUILD_DIR} ${BUILD_FLAGS}
    cmake --build ${BUILD_DIR} --config ${CONFIG_FLAG} -j ${NPROC}

    # Copy the built libraries to the appropriate directory
    for LIB_PATH in "${LIBRARY_PATHS[@]}"; do
        cp ${LIB_PATH} ${LIBS_DEST_DIR}
    done
}

# Build llama.cpp libraries with Metal support
LLAMA_FOLDER=${ROOT_DIR}/dependency/llama.cpp
LLAMA_LIB_PATHS=(
    "${LLAMA_FOLDER}/${BUILD_DIR}/src/libllama.dylib"
    "${LLAMA_FOLDER}/${BUILD_DIR}/ggml/src/libggml.dylib"
    "${LLAMA_FOLDER}/${BUILD_DIR}/examples/llava/libllava_shared.dylib"
)
build_and_copy_libs "${LLAMA_FOLDER}" "-DBUILD_SHARED_LIBS=ON" "${LLAMA_LIB_PATHS[@]}"

# Build stable-diffusion.cpp libraries with Metal support
SD_FOLDER=${ROOT_DIR}/dependency/stable-diffusion.cpp
SD_LIB_PATHS=(
    "${SD_FOLDER}/${BUILD_DIR}/bin/libstable-diffusion.dylib"
)
build_and_copy_libs "${SD_FOLDER}" "-DSD_BUILD_SHARED_LIBS=ON" "${SD_LIB_PATHS[@]}"

# Handle pyproject.toml for Metal build
cd ${ROOT_DIR}
cp pyproject_macos_metal.toml pyproject.toml
python -m build
rm pyproject.toml

# Install the package
pip install dist/nexa*.whl --force-reinstall

# Pause before exiting
read -p "Press [Enter] key to continue..."