## About

This project demonstrates a simple agent implementation built on **Nexa SDK Serve**. 

## Setup

### Prerequisites

- Nexa SDK Installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#step-1-download-nexa-cli-with-one-click))
- Python 3.11 – 3.13


### Install Dependencies

```bash
# Navigate to the agent directory
cd Agent

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

In a new terminal window

```bash

# Option 2: Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py

```