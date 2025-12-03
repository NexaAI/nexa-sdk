## About

This demo showcases a function tool implementation using NexaAI Python binding. 

## Setup

### Prerequisites

| Platform | Required Python | 
|----------|----------------|
| **Windows (ARM64, Snapdragon X Elite)** | **3.11 – 3.13 (arm64)** |
| **macOS / Windows (x64)** | **3.10 (x64)** |

For python install, you can check our interactive Jupyter notebooks. Choose the appropriate notebook for your platform:
- [macOS Installation Guide](../../../bindings/python/notebook/macos.ipynb)
- [windows(x64) Installation Guide](../../../bindings/python/notebook/windows(x64).ipynb)
- [windows(ARM64) Installation Guide](../../../bindings/python/notebook/winodws(arm64).ipynb)

### Install Dependencies

```bash
# Navigate to the example directory
cd Python-Binding-Example

# Create a Python virtual environment
python -m venv .venv

# Activate the virtual environment
.\.venv\Scripts\activate # windows

source .venv/bin/activate # macOS


# Install all required dependencies
pip install -r requirements.txt
```

## Running the Example

```bash
# Run the CLI version which provides an interactive terminal interface
# This version allows direct interaction with the agent through command line
python agent_nexa.py

# Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py

```

## Usage Examples

- **Web Search**: "What's the latest AI news?"
- **File Operations**: "Save this conversation to notes.txt"
