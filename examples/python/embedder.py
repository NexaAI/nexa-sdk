#!/usr/bin/env python3

"""
NexaAI Embedding Example - Text Embedding Generation

This example demonstrates how to use the NexaAI SDK to generate text embeddings.
It includes basic model initialization, single and batch embedding generation, and embedding analysis.
"""

import os

from nexaai.embedder import Embedder, EmbeddingConfig


def main():
    model_path = os.path.expanduser(
        "~/.cache/nexa.ai/nexa_sdk/models/nexaml/jina-v2-fp16-mlx/model.safetensors")

    embedder: Embedder = Embedder.from_(
        name_or_path=model_path, plugin_id="mlx")
    print('Embedder loaded successfully!')

    dim = embedder.get_embedding_dim()
    print(f"Dimension: {dim}")

    texts = [
        "The Eiffel Tower is in Paris and France is beautiful.",
        "Bananas are yellow and delicious fruits.",
        "Machine learning is a subset of artificial intelligence."
    ]
    embeddings = embedder.generate(
        texts=texts, config=EmbeddingConfig(batch_size=len(texts)))
    print(f"Embeddings: {embeddings}")


if __name__ == "__main__":
    main()
