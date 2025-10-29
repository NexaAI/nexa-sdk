# World's First Fully NPU-Supported RAG Pipeline

## About
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
- **Generation:** NexaAI/Llama3.2-3B-NPU-Turbo
- **Runtime:** NexaML with OpenAI-compatible APIs

Bring your own files (PDFs, Word docs, text) and ask questions—the system retrieves relevant context and generates answers entirely on your device.

## Examples
- [Python-Binding-Example](./Python-Binding-Example/README.md)
- [Serve-Example](./Serve-Example/README.md)

## Additional Resources

- [Nexa SDK Repository](https://github.com/NexaAI/nexa-sdk)
- [Notebook](https://github.com/NexaAI/nexa-sdk/tree/main/bindings/python/notebook)
- [python-3.11.1-arm64.exe](https://www.python.org/ftp/python/3.11.1/python-3.11.1-arm64.exe)
- [python-3.13.8-arm64.exe](https://www.python.org/ftp/python/3.13.8/python-3.13.8-arm64.exe)