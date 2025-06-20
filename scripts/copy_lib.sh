#!/bin/bash

# Create directories
mkdir -p build/lib
mkdir -p build/include

# Detect operating system
OS=$(uname)

if [ "$OS" = "Darwin" ]; then
    # macOS - copy .dylib files
    cp ../nexa-sdk-binding/third-party/llama.cpp/build/bin/*.dylib ./build/lib/ 2>/dev/null || true
    cp ../nexa-sdk-binding/backends/llama-cpp/interface/build/*.dylib ./build/lib/ 2>/dev/null || true
    echo "Copied .dylib files to build/lib/"
else
    # Linux and other Unix systems - copy .so files
    cp ../nexa-sdk-binding/third-party/llama.cpp/build/bin/*.so ./build/lib/ 2>/dev/null || true
    cp ../nexa-sdk-binding/backends/llama-cpp/interface/build/*.so ./build/lib/ 2>/dev/null || true
    echo "Copied .so files to build/lib/"
fi

# Copy header files (same for all platforms)
cp ../nexa-sdk-binding/common/include/llm.h ./build/include/
echo "Copied header files to build/include/"