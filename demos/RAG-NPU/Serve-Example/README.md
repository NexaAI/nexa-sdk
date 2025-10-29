## About

This demo showcases a RAG implementation using Nexa Sdk.

## Setup

### Prerequisites
- Windows ARM64 device with Snapdragon NPU
- Python 3.11 – 3.13 (ARM64 build)
- Nexa SDK installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk))


### Download Models
Download the three NPU-optimized models:

```bash
nexa pull NexaAI/embeddinggemma-300m-npu

nexa pull NexaAI/jina-v2-rerank-npu

nexa pull NexaAI/Llama3.2-3B-NPU-Turbo
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
python rag_nexa.py --data ..\docs

# Option 2: Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py

```