# World's First Fully NPU-Supported RAG Pipeline

## 1. About
This is the **world's first fully NPU-supported RAG pipeline** running entirely on Qualcomm Snapdragon NPU with state-of-the-art models.

**What makes it special:**
- 🔒 **100% Private** — All processing happens locally, nothing leaves your device
- ⚡ **10× More Power Efficient** — Runs on NPU instead of GPU
- 🌟 **State-of-the-art Models** — Best-in-class embedding, reranking, and generation
- 🔌 **Always-On** — Efficient enough to run as a background service

![The Stack](./architecture.png)

**The Stack:**
- **Embedding:** Gemma-300M (Google DeepMind) — Top multilingual embedding model
- **Rerank:** Jina Reranker v2 — SOTA cross-lingual reranking
- **Generation:** Granite 4.0-Micro (IBM Research) — Frontier reasoning in 3B parameters
- **Runtime:** NexaML with OpenAI-compatible APIs

Bring your own files (PDFs, Word docs, text) and ask questions—the system retrieves relevant context and generates answers entirely on your device.

---

## 2. Setup

### Prerequisites
- Windows ARM64 device with Snapdragon NPU
- Python 3.11 – 3.13 (ARM64 build)
- Nexa SDK installed ([Installation Guide](https://github.com/NexaAI/nexa-sdk))


### Install Models
Download the three NPU-optimized models:

```bash
nexa pull NexaAI/embeddinggemma-300m-npu

nexa pull NexaAI/jina-v2-rerank-npu

nexa pull NexaAI/Granite-4-Micro-NPU
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

```bash

# Start Nexa server
nexa serve

# Run the CLI version which provides an interactive terminal interface
# This version allows direct interaction with the agent through command line
python rag_nexa.py --data ..\docs

# Run the Gradio UI version
# This starts a local web server with a chat interface at http://localhost:7860
python gradio_ui.py

```

**How it works:**
- System creates `Downloads\nexa-rag-docs` folder for your documents
- Add files (.pdf, .txt, .docx) to this folder
- Run with `--rebuild` flag or type `:reload` after adding new files
- Ask questions and get answers based on your documents

**Features:**
- **Left panel:** Upload files and rebuild index
- **Right panel:** Chat interface with streaming responses
- Fully interactive local RAG experience


## Additional Resources

- [Nexa SDK Repository](https://github.com/NexaAI/nexa-sdk)
- [Granite Models](https://huggingface.co/ibm-granite)
- [python-3.11.1-arm64.exe](https://www.python.org/ftp/python/3.11.1/python-3.11.1-arm64.exe)
- [python-3.13.8-arm64.exe](https://www.python.org/ftp/python/3.13.8/python-3.13.8-arm64.exe)