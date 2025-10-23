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
- Nexa SDK installed

### Install Models
Download the three NPU-optimized models:

```bash
nexa pull NexaAI/embeddinggemma-300m-npu
nexa pull NexaAI/jina-v2-rerank-npu
nexa pull NexaAI/Granite-4-Micro-NPU
```

### Start Nexa Server
Launch the server in a separate terminal:

```bash
nexa serve
```

### Install Dependencies
```bash
# Optional: Create conda environment
conda create -n rag-nexa python=3.10 -y
conda activate rag-nexa

# Install dependencies
pip install -r requirements.txt
```

---

## 3. Quick Start

### CLI Mode
```bash
# Serve Example
python Serve-Example/rag_nexa.py --rebuild

# Python binding Example
python Python-Binding-Example/rag_nexa.py --rebuild

```

**How it works:**
- System creates `Downloads\nexa-rag-docs` folder for your documents
- Add files (.pdf, .txt, .docx) to this folder
- Run with `--rebuild` flag or type `:reload` after adding new files
- Ask questions and get answers based on your documents

### Gradio UI Mode
```bash
# Serve Example
python Serve-Example/gradio_ui.py

# Python binding Example
python Python-Binding-Example/gradio_ui.py
```

Open [http://127.0.0.1:7860](http://127.0.0.1:7860) in your browser.

**Features:**
- **Left panel:** Upload files and rebuild index
- **Right panel:** Chat interface with streaming responses
- Fully interactive local RAG experience
