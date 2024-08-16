import os
from setuptools import setup, find_packages
import platform

# Metadata
NAME = "nexaai"
VERSION = "0.0.1.dev4"
DESCRIPTION = "Nexa AI SDK"
LONG_DESCRIPTION = open(os.path.join(os.path.dirname(__file__), "README.md")).read()
AUTHOR = "Nexa AI"
AUTHOR_EMAIL = "octopus@nexa4ai.com"
URL = "https://github.com/NexaAI/nexa-sdk"

# Package data
package_data = {
    "nexa": []
}

# Use the TARGET_PLATFORM environment variable to determine the target platform or use the current platform
target_platform = os.environ.get("TARGET_PLATFORM", platform.system())

# Read requirements from files
def read_requirements(filename):
    req_file_path = os.path.join(os.path.dirname(__file__), filename)
    with open(req_file_path, 'r') as file:
        return [line.strip() for line in file if line.strip() and not line.startswith('#')]

base_requirements = read_requirements('requirements.txt')
onnx_requirements = read_requirements('requirements-onnx.txt')

setup(
    name=NAME,
    version=VERSION,
    description=DESCRIPTION,
    long_description=LONG_DESCRIPTION,
    long_description_content_type="text/markdown",
    author=AUTHOR,
    author_email=AUTHOR_EMAIL,
    url=URL,
    packages=find_packages(),
    install_requires=base_requirements,
    extras_require={
        "onnx": onnx_requirements,
    },
    python_requires=">=3.7",
    classifiers=[
        "Programming Language :: Python :: 3 :: Only",
        "Programming Language :: Python :: 3.7",
        "Programming Language :: Python :: 3.8",
        "Programming Language :: Python :: 3.9",
        "Programming Language :: Python :: 3.10",
        "Programming Language :: Python :: 3.11",
        "Programming Language :: Python :: 3.12",
    ],
    package_data=package_data,
    include_package_data=True,
    entry_points={
        "console_scripts": [
            "nexa-cli = nexa.cli.entry:main",
        ]
    },
)
"""
# test

rm -rf dist build nexaai.egg-info
python -m build
pip install dist/*.whl --force-reinstall # to install gguf
pip install 'nexaai[onnx]' --find-links=dist # to install gguf and onnx

# upload

twine upload dist/*.whl
"""