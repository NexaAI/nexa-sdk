## About

This demo showcases a function tool implementation using Nexa Sdk. 

## Setup

### Prerequisites

- Nexa SDK Installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#step-1-download-nexa-cli-with-one-click))
- Python 3.11 – 3.13

### Download Models

If you use the `NexaAI/Granite-4.0-h-350M-NPU` model, you will need an access token, which can be obtained from ([sdk.nexa.ai](https://sdk.nexa.ai/)), and then set the token using the `nexa config` command below. 

```bash
# Set access token
nexa config set license '<access_token>'
```

```bash
# Download model — only required on Windows platform
nexa pull NexaAI/Granite-4.0-h-350M-NPU

# Download model (compatible with both Windows and macOS)
nexa pull NexaAI/granite-4.0-micro-GGUF

```

### Install Dependencies

```bash
# Navigate to the example directory
cd Serve-Example

# Create a Python virtual environment
python -m venv .venv

# Activate the virtual environment
.\.venv\Scripts\activate # windows

source .venv/bin/activate # macOS

# Install all required dependencies
pip install -r requirements.txt
```

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