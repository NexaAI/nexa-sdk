# NexaAI Python Examples

This directory contains examples for using the NexaAI Python SDK.

## Prerequisites

Before installing **NexaAI SDK**, please ensure you have the correct **Python version** and **architecture**.

### 1. Check your Python environment

```sh
python -c "import sys,platform;print(f'Python version: {sys.version} | Architecture: {platform.machine()}')"
```

### 2. Python version requirements

| Platform                                | Required Python         | Notes                                                                                                                    |
| --------------------------------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| **Windows (ARM64, Snapdragon X Elite)** | **3.11 – 3.13 (arm64)** | Please install **official ARM64 Python** from [python-3.11.1-arm64.exe](https://www.python.org/ftp/python/3.11.1/python-3.11.1-arm64.exe). |
| **Linux / macOS / Windows (x64)**       | **3.10 (x64)**          | You can use **conda** to install and manage this version.                                                                |

**Create a Python 3.10 environment (recommended for x64):**

```sh
conda create -n nexaai python=3.10
conda activate nexaai
```

### 3. Install NexaAI SDK

* **Windows / Linux**

  ```bash
  pip install nexaai
  ```
* **macOS (Apple Silicon)**

  ```bash
  pip install 'nexaai[mlx]'
  ```

### Authentication Setup

Before running any examples, you need to set up your NexaAI authentication token.

### Set Token in Code

Replace `"YOUR_NEXA_TOKEN_HERE"` with your actual NexaAI token from [https://sdk.nexa.ai/](https://sdk.nexa.ai/):

- Linux/macOS:
  ```bash
  export NEXA_TOKEN="YOUR_NEXA_TOKEN_HERE"
  ```
- Windows:
  ```powershell
  $env:NEXA_TOKEN="YOUR_NEXA_TOKEN_HERE"
  ```

## Running Examples

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

### CV

#### OCR

```bash
python cv_ocr.py
```

## Running Examples (Windows ARM64, Snapdragon X Elite)

### LLM

```bash
python llm.py --model NexaAI/Llama3.2-3B-NPU-Turbo --plugin-id npu --device npu --max-tokens 100 --system "You are a helpful assistant."
```

### Multi-Modal

```bash
python vlm.py --model NexaAI/OmniNeural-4B --plugin-id npu --device npu --max-tokens 100 --system "You are a helpful assistant."
```

### Reranker

```bash
python rerank.py --model NexaAI/jina-v2-rerank-npu --plugin-id npu --query "Where is on-device AI?" --documents "On-device AI is a type of AI that is processed on the device itself, rather than in the cloud." "edge computing" "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality." "The capital of France is Paris."
```

### Embedder

```bash
python embedder.py --model NexaAI/embeddinggemma-300m-npu --plugin-id npu --texts "On-device AI is a type of AI that is processed on the device itself, rather than in the cloud." "edge computing" "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality." "The capital of France is Paris." --query "what is on device AI" --batch-size 2
```

### CV

#### OCR

```bash
python cv_ocr.py --model NexaAI/paddleocr-npu --image path/to/image.png
```

### ASR

```bash
python asr.py --model NexaAI/parakeet-npu --audio path/to/audio.wav
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
