## About

This demo showcases a function tool implementation using Nexa Sdk. 

## Setup

### Prerequisites
- Windows ARM64 device with Snapdragon NPU
- Python 3.11 – 3.13 (ARM64 build) - We provide detailed installation guides in Jupyter Notebook format
[Windows ARM64 Installation Guide](https://github.com/NexaAI/nexa-sdk/blob/main/bindings/python/notebook/winodws(arm64).ipynb)
- Nexa SDK installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#step-1-download-nexa-cli-with-one-click))

### Download Models
Download the three NPU-optimized models:

```bash

# Download model
nexa pull NexaAI/granite-4-Nano-NPU
```

### Install Dependencies

```bash
# Navigate to the example directory
cd Serve-Example

# Create a Python virtual environment
python -m venv .venv

# Activate the virtual environment
.\.venv\Scripts\activate

# Install all required dependencies
pip install -r requirements.txt
```

Note: Make sure you're using Python 3.11-3.13 (ARM64) as specified in the prerequisites.

### Running the Example

First, open a new terminal window and start the Nexa server:
```bash
# Start Nexa server
nexa serve
```

In a new terminal window, you can run either the CLI or Gradio UI version:

```bash
# Option 1: Run the CLI version which provides an interactive terminal interface
# This version allows direct interaction with the agent through command line
python agent_nexa.py

# Option 2: Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py

```