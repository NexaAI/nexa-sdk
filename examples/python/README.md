# NexaAI Python Examples

This directory contains examples for using the NexaAI Python SDK.

## Prerequisites

- Python 3.10
- Install the NexaAI Python SDK
```bash
pip install nexaai==1.0.4rc10
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
nexa pull nexaml/jina-v2-rerank-mlx

python reranker.py
```

### Embedder

```bash
nexa pull nexaml/jina-v2-fp16-mlx

python embedder.py
```

### CV

#### OCR

```bash
nexa pull nexaml/paddle-ocr-mlx

python cv_ocr.py
```