## About

This demo showcases how to use NexaAI Python Binding to build a Vision-Language Model (VLM) application that supports structured JSON outputs, multi-image input, and custom system prompt control.

## Setup

### Prerequisites
| Platform | Required Python | 
|----------|----------------|
| **Windows (ARM64, Snapdragon X Elite)** | **3.11 – 3.13 (arm64)** |
| **macOS / Windows (x64)** | **3.10 (x64)** |

### Install Dependencies

```bash
# Navigate to the example directory
cd Python-Binding-Example

# Create a Python virtual environment
python -m venv .venv

# Activate the virtual environment (windows)
.\.venv\Scripts\activate

# Activate the virtual environment (macOS)
source .venv/bin/activate

# Install all required dependencies
pip install nexaai
pip install gradio

```

## Running the Example

```bash
# Run the CLI version which provides an interactive terminal interface
python vlm_serve.py

# Run the Gradio UI version
python gradio_ui.py

```