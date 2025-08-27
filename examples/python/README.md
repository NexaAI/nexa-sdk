# NexaAI Python Examples

This directory contains examples for using the NexaAI Python SDK.

## Prerequisites

- Python 3.10
- Install the latest NexaAI Python SDK from [PyPI](https://pypi.org/project/nexaai/#history).

    For example:
    ```bash
    pip install nexaai==1.0.4rc14
    ```

## Running Examples

### LLM

```bash
nexa pull Qwen/Qwen3-0.6B-GGUF

python llm.py
```
### Multi-Modal

```bash
nexa pull mlx-community/gemma-3-4b-it-8bit

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