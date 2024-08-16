#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status.
set -o pipefail  # Return exit status of the last command in the pipeline that failed.

# Handle pyproject.toml for Linux build
cp pyproject_linux_cpu.toml pyproject.toml
python -m build
rm pyproject.toml

cp pyproject_linux_cuda.toml pyproject.toml
python -m build
rm pyproject.toml

# Install the package
# pip install dist/nexa*.whl --force-reinstall

# Pause before exiting
read -p "Press [Enter] key to continue..."