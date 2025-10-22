# Agent with Granite-4-Nano-NPU

## Overview

This demo showcases a lightweight on-device AI assistant powered by Nexa SDK's Granite-4-Nano-NPU model. The agent executes tasks through function calling, including web searches. Built with a Gradio interface for local NPU-accelerated agentic workflows.

The demo provides two different implementation approaches:

1. **Serve-Example**: Uses the `nexa serve` mode, where the model is hosted by Nexa server and accessed via HTTP API.
2. **Python-Binding-Example**: Uses the Nexa SDK Python bindings

## Prerequisites
- Python 3.13+
- Nexa SDK installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk))
- Nexa access token

## Quick Start

### 1. Environment Setup

```bash
# Create conda environment
conda create -n granite-arm64 python==3.13 -y

conda activate granite-arm64

# Install dependencies
pip install -r requirements.txt
```

### 2. Choose an Implementation

You can run the agent using either the Serve mode or Python bindings:

#### Option 1: Serve Mode (Serve-Example)

```bash
# Set access token (Windows PowerShell)
$env:NEXA_TOKEN="your_token_here"

# Start Nexa server
nexa serve

# In a new terminal, start agent backend
python Serve-Example/agent_nexa.py

# In another terminal, start Gradio UI
python Serve-Example/gradio_ui.py
```

#### Option 2: Python Bindings (Python-Binding-Example)

```bash
# Start agent backend
python Python-Binding-Example/agent_nexa.py

# In another terminal, start Gradio UI
python Python-Binding-Example/gradio_ui.py
```

## Usage Examples

- **Web Search**: "What's the latest AI news?"
- **File Operations**: "Save this conversation to notes.txt"
- **Memory Queries**: Ask questions that reference previous conversations
- **General Chat**: Regular Q&A without function calls

## Key Features

- Function calling for autonomous web search and file management
- FAISS-based memory system for context retention
- Fully local execution with NPU acceleration

## Additional Resources

- [Nexa SDK Repository](https://github.com/NexaAI/nexa-sdk)
- [Granite Models](https://huggingface.co/ibm-granite)