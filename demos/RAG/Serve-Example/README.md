## About

This demo showcases a RAG implementation using Nexa Sdk.

## Setup

### Prerequisites

- Nexa SDK Installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk?tab=readme-ov-file#step-1-download-nexa-cli-with-one-click))
- Python 3.11 – 3.13

## Download Models

### Snapdragon NPU (Windows ARM64)
These models are optimized for Qualcomm NPU:

```bash
nexa pull NexaAI/embeddinggemma-300m-npu

nexa pull NexaAI/jina-v2-rerank-npu

nexa pull NexaAI/Llama3.2-3B-NPU-Turbo
```

### macOS & Windows x64

```bash
nexa pull NexaAI/Qwen3-4B-GGUF

nexa pull jinaai/jina-embeddings-v4-text-retrieval-GGUF

nexa pull jinaai/jina-reranker-v3-GGUF
```
💡 These models are fully compatible with macOS and Windows x64.
No NPU is required; they run on CPU/GPU.


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
python rag_nexa.py --data ../docs

# Option 2: Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py

```