# Agent with Granite-4-Nano-NPU

## Overview

This demo showcases a lightweight on-device AI assistant powered by Nexa SDK's Granite-4-Nano-NPU model. The agent executes tasks through function calling, including web searches. Built with a Gradio interface for local NPU-accelerated agentic workflows.

## Prerequisites
- Python 3.11 – 3.13 (ARM64 build)
- Nexa SDK installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk))
- Nexa access token

## Quick Start

### 1. Environment Setup

```bash
# Set access token (Windows PowerShell)
$env:NEXA_TOKEN="your_token_here"

# Download model
nexa pull NexaAI/granite-4-Nano-NPU

# Start Nexa server
nexa serve

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

### 2. Running the Example

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

## Additional Resources

- [Nexa SDK Repository](https://github.com/NexaAI/nexa-sdk)
- [Granite Models](https://huggingface.co/ibm-granite)
- [SerpAPI Documentation](https://serpapi.com/)
- [python-3.11.1-arm64.exe](https://www.python.org/ftp/python/3.11.1/python-3.11.1-arm64.exe)
- [python-3.13.8-arm64.exe](https://www.python.org/ftp/python/3.13.8/python-3.13.8-arm64.exe)