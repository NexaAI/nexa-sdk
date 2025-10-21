# Agent with Granite-4-Nano-NPU

## Overview

This demo showcases a lightweight on-device AI assistant powered by Nexa SDK's Granite-4-Nano-NPU model. The agent executes tasks through function calling, including web searches, file operations, and memory-based interactions using FAISS vector storage. Built with a Gradio interface for local NPU-accelerated agentic workflows.

## Prerequisites
- Python 3.13+
- Nexa SDK installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk))
- Nexa access token

## Quick Start

### 1. Environment Setup

```bash
# Create conda environment
conda create -n granite-agent python==3.13 -y
conda activate granite-agent

# Install dependencies
pip install -r requirements.txt
```

### 2. Start Nexa Server

```bash
# Set access token (Windows PowerShell)
$env:NEXA_TOKEN="your_token_here"

# Start server
nexa serve
```

### 3. Run the Agent

In a new terminal:

```bash
# Activate environment
conda activate granite-agent

# Start agent backend
python agent_nexa.py

# In another terminal, start Gradio UI
python gradio_ui.py
```

Open your browser at [http://127.0.0.1:7860](http://127.0.0.1:7860)

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