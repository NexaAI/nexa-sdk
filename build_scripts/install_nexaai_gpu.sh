#!/bin/bash

# Function to install GPU packages for Windows
install_windows() {
    echo "Detected Windows"
    pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/cu124 --extra-index-url https://pypi.org/simple
    if [ $? -ne 0 ]; then
        echo "GPU installation failed, falling back to CPU version..."
        pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
    fi
    
    CMAKE_ARGS="-DSD_CUBLAS=on" pip install stable-diffusion-cpp-python
    if [ $? -ne 0 ]; then
        echo "GPU installation failed, falling back to CPU version..."
        pip install stable-diffusion-cpp-python --index-url https://nexaai.github.io/stable-diffusion-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
    fi
    
    pip install nexaai
}

# Function to install GPU packages for macOS
install_macos() {
    # Detect architecture
    arch=$(uname -m)
    if [[ "$arch" == "x86_64" ]]; then
        echo "Detected macOS X86"
        pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
        pip install stable-diffusion-cpp-python --index-url https://nexaai.github.io/stable-diffusion-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
    elif [[ "$arch" == "arm64" ]]; then
        echo "Detected macOS ARM with Metal support"
        pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/metal --extra-index-url https://pypi.org/simple
        if [ $? -ne 0 ]; then
            echo "GPU installation failed, falling back to CPU version..."
            pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
        fi
        
        pip install stable-diffusion-cpp-python --index-url https://nexaai.github.io/stable-diffusion-cpp-python/whl/metal --extra-index-url https://pypi.org/simple
        if [ $? -ne 0 ]; then
            echo "GPU installation failed, falling back to CPU version..."
            pip install stable-diffusion-cpp-python --index-url https://nexaai.github.io/stable-diffusion-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
        fi
    else
        echo "Unsupported macOS architecture: $arch"
        exit 1
    fi
    pip install nexaai
}

# Function to install GPU packages for Linux
install_linux() {
    echo "Detected Linux"
    pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/cu124 --extra-index-url https://pypi.org/simple
    if [ $? -ne 0 ]; then
        echo "GPU installation failed, falling back to CPU version..."
        pip install llama-cpp-python==0.2.86 --index-url https://abetlen.github.io/llama-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
    fi
    
    pip install stable-diffusion-cpp-python --index-url https://nexaai.github.io/stable-diffusion-cpp-python/whl/cu124 --extra-index-url https://pypi.org/simple
    if [ $? -ne 0 ]; then
        echo "GPU installation failed, falling back to CPU version..."
        pip install stable-diffusion-cpp-python --index-url https://nexaai.github.io/stable-diffusion-cpp-python/whl/cpu --extra-index-url https://pypi.org/simple
    fi
    
    pip install nexaai
}

# Main script to detect the operating system
OS=$(uname -s)

case "$OS" in
    MINGW*|MSYS*|CYGWIN*)
        install_windows
        ;;
    Darwin)
        install_macos
        ;;
    Linux)
        install_linux
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac
