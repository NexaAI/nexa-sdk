## About

This demo showcases a RAG implementation using NexaAI Python bindings. 

## Setup

### Prerequisites
- Windows ARM64 device with Snapdragon NPU
- Python 3.11 – 3.13 (ARM64 build)

### Install Dependencies

```bash
# Navigate to the example directory
cd Python-Binding-Example

# Create a Python virtual environment
python -m venv .venv

# Activate the virtual environment
.\.venv\Scripts\activate

# Install all required dependencies
pip install -r requirements.txt
```

Note: Make sure you're using Python 3.11-3.13 (ARM64) as specified in the prerequisites. 

## 2. Running the Example

```bash
# Option 1: Run the CLI version which provides an interactive terminal interface
# This version allows direct interaction with the agent through command line
python rag_nexa.py --data ..\docs

# Option 2: Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py
```
