# NexaAI Python Examples

This directory contains examples for using the NexaAI Python SDK.

## Prerequisites

- Python 3.10
  - if you are using conda, you can create a new environment via
    ```sh
    conda create -n nexaai python=3.10
    conda activate nexaai
    ```
- Install the latest NexaAI Python SDK from [PyPI](https://pypi.org/project/nexaai/#history).

  Install command by OS:

  - Windows and Linux:
    ```bash
    pip install nexaai
    ```
  - macOS:
    ```bash
    pip install 'nexaai[mlx]'
    ```

## Running Examples

### LLM

```bash
nexa pull Qwen/Qwen3-0.6B-GGUF

python llm.py
```

### Multi-Modal

```bash
nexa pull NexaAI/gemma-3n-E4B-it-4bit-MLX

python vlm.py
```

### Reranker

```bash
nexa pull NexaAI/jina-v2-rerank-mlx

python rerank.py
```

### Embedder

```bash
nexa pull NexaAI/jina-v2-fp16-mlx

python embedder.py
```

### CV

#### OCR

```bash
nexa pull NexaAI/paddle-ocr-mlx

python cv_ocr.py
```
## Running Examples (Windows ARM64, Snapdragon X Elite)

### LLM
```bash
nexa pull NexaAI/Llama3.2-3B-NPU-Turbo

python llm.py --model NexaAI/Llama3.2-3B-NPU-Turbo --plugin-id npu --device npu --max-tokens 100 --system "You are a helpful assistant."
```

### Multi-Modal

```bash
nexa pull NexaAI/OmniNeural-4B

python vlm.py --model NexaAI/OmniNeural-4B --plugin-id npu --device npu --max-tokens 100 --system "You are a helpful assistant."
```

### Reranker
```bash
nexa pull NexaAI/jina-v2-rerank-npu

python rerank.py --model NexaAI/jina-v2-rerank-npu --plugin-id npu --query "Where is on-device AI?" --documents "On-device AI is a type of AI that is processed on the device itself, rather than in the cloud." "edge computing" "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality." "The capital of France is Paris."
```

### Embedder
```bash
nexa pull NexaAI/embeddinggemma-300m-npu

python embedder.py --model NexaAI/embeddinggemma-300m-npu --plugin-id npu --texts "On-device AI is a type of AI that is processed on the device itself, rather than in the cloud." "edge computing" "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality." "The capital of France is Paris." --query "what is on device AI" --batch-size 2
```

### CV

#### OCR
```bash
nexa pull NexaAI/paddleocr-npu

python cv_ocr.py --det-model NexaAI/paddleocr-npu --rec-model NexaAI/paddleocr-npu --image path/to/image.png
```

### ASR
```bash
nexa pull NexaAI/parakeet-npu

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
- `mlx`: Apple Silicon optimized (for supported models)
- `llama_cpp`: For GGUF format models
- `onnx`: ONNX runtime backend