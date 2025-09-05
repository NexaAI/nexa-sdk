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
    pip install nexaai[mlx]
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
