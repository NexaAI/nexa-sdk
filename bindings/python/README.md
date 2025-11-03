# NexaAI Python SDK

This directory contains the NexaAI Python SDK and comprehensive examples for various AI inference tasks.

## Quick Start

The easiest way to get started with NexaAI is through our interactive Jupyter notebooks. Choose the appropriate notebook for your platform:

### 📓 Interactive Notebooks

| Platform | Notebook | Description |
|----------|----------|-------------|
| **macOS** | [`notebook/macos.ipynb`](notebook/macos.ipynb) | Complete examples for macOS with Apple Silicon optimization |
| **Windows (x64)** | [`notebook/windows(x64).ipynb`](notebook/windows(x64).ipynb) | Examples for Windows x64 systems |
| **Windows (ARM64)** | [`notebook/winodws(arm64).ipynb`](notebook/winodws(arm64).ipynb) | NPU-optimized examples for Snapdragon X Elite |

Each notebook includes:
- **LLM (Large Language Model)**: Text generation and conversation
- **VLM (Vision Language Model)**: Multimodal understanding and generation  
- **Embedder**: Text vectorization and similarity computation
- **Reranker**: Document reranking
- **ASR (Automatic Speech Recognition)**: Speech-to-text transcription
- **CV (Computer Vision)**: OCR/text recognition

## Prerequisites

| Platform | Required Python | 
|----------|----------------|
| **Windows (ARM64, Snapdragon X Elite)** | **3.11 – 3.13 (arm64)** |
| **macOS / Windows (x64)** | **3.10 (x64)** |

## Command Line Examples

If you prefer command-line usage, here are the basic examples:

### LLM
```bash
python llm.py
```

### Multi-Modal
```bash
python vlm.py
```

### Reranker
```bash
python rerank.py
```

### Embedder
```bash
python embedder.py
```

### Computer Vision
```bash
python cv_ocr.py
```

## Common Arguments

- `--model`: Path to the model file
- `--device`: Device to run on (cpu, gpu, etc.)
- `--max-tokens`: Maximum tokens to generate (for LLM/VLM)
- `--batch-size`: Batch size for processing
- `--system`: System message for chat models
- `--plugin-id`: Plugin ID to use (default: cpu_gpu)

## Plugin ID Options

The `--plugin-id` parameter supports different backends:

- `cpu_gpu`: Default, supports both CPU and GPU
- `metal`: Apple Silicon optimized (for supported models)
- `npu`: Qualcomm NPU optimized (for supported models)
- `nexaml`: NexaML optimized (for supported models)

### Supported Models by Backend

| Backend | Supported Models |
|---------|------------------|
| `cpu_gpu` | GGUF models (default backend) |
| `metal` | Models with MLX format (e.g., Qwen3-VL-4B-MLX-4bit, gpt-oss-20b-MLX-4bit) |
| `npu` | **LLM:** Granite-4-Micro-NPU, phi4-mini-npu-turbo, Qwen3-4B-Instruct-2507-npu, Qwen3-4B-Thinking-2507-npu, Llama3.2-3B-NPU-Turbo, jan-v1-4B-npu, qwen3-4B-npu, phi3.5-mini-npu<br>**VLM:** Qwen3-VL-4B-Instruct-NPU, OmniNeural-4B, LFM2-1.2B-npu<br>**Embedder:** embeddinggemma-300m-npu<br>**ASR:** parakeet-tdt-0.6b-v3-npu<br>**CV:** convnext-tiny-npu, paddleocr-npu, yolov12-npu<br>**Reranker:** jina-v2-rerank-npu |
| `nexaml` | **VLM:** Qwen3-VL-4B-Instruct-GGUF:Q4_0, Qwen3-VL-4B-Thinking-GGUF:Q4_0 |

## Getting Started

1. **Choose your platform** and open the corresponding notebook from the [`notebook/`](notebook/) directory
2. **Follow the setup instructions** in the notebook for your specific platform
3. **Run the examples** step by step to explore different AI capabilities
4. **Customize the examples** for your specific use cases

For detailed setup instructions and platform-specific requirements, please refer to the individual notebooks.
