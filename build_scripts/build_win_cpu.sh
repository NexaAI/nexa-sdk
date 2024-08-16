#!/bin/bash

# To run this file, open PowerShell and run the following command:
# .\build_win.sh

ROOT_DIR=$(pwd)
NPROC=$(nproc)
CONFIG_FLAG=Release
DEVICE=cpu
LIBS_DEST_DIR=${ROOT_DIR}/nexa/gguf/_libs/${DEVICE}
BUILD_FOLDER=build_${DEVICE}

# Function to build a project and copy DLLs
build_and_copy_libs() {
    local PROJECT_NAME=$1
    local BUILD_FLAGS=$2
    local DLL_FILES=("${@:3}")

    PROJECT_FOLDER=${ROOT_DIR}/dependency/${PROJECT_NAME}.cpp

    echo "Building ${PROJECT_NAME}"
    cd ${PROJECT_FOLDER}

    # Build libs
    rm -rf ${BUILD_FOLDER}
    cmake . -B ${BUILD_FOLDER} ${BUILD_FLAGS}
    cmake --build ${BUILD_FOLDER} --config ${CONFIG_FLAG} -j ${NPROC}

    # Copy DLL files
    for DLL_FILE in "${DLL_FILES[@]}"; do
        cp ${PROJECT_FOLDER}/${BUILD_FOLDER}/bin/${CONFIG_FLAG}/${DLL_FILE} ${LIBS_DEST_DIR}/
    done
}

# Build llama (and llava_shared)
build_and_copy_libs "llama" "-DBUILD_SHARED_LIBS=ON" "ggml.dll" "llama.dll" "llava_shared.dll"

# Build stable-diffusion
build_and_copy_libs "stable-diffusion" "-DSD_BUILD_SHARED_LIBS=ON" "stable-diffusion.dll"
