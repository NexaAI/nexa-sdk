#!/usr/bin/env python3

"""
NexaAI Embedding Example - Text Embedding Generation

This example demonstrates how to use the NexaAI SDK to generate text embeddings.
It includes basic model initialization, single and batch embedding generation, and embedding analysis.

LICENSE NOTICE - DUAL LICENSING:
- NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
- GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)
For NPU commercial use, contact: dev@nexa.ai | See LICENSE-NPU

Copyright (c) 2025 Nexa AI
"""

import os
import numpy as np

from nexaai.embedder import Embedder, EmbeddingConfig


def main():
    model_path = os.path.expanduser(
        "~/.cache/nexa.ai/nexa_sdk/models/NexaAI/jina-v2-fp16-mlx/model.safetensors"
    )

    # For now, this modality is only supported in MLX.
    embedder: Embedder = Embedder.from_(name_or_path=model_path, plugin_id="mlx")
    print("Embedder loaded successfully!")

    dim = embedder.get_embedding_dim()
    print(f"Dimension: {dim}")

    texts = [
        "On-device AI is a type of AI that is processed on the device itself, rather than in the cloud.",
        "Nexa AI allows you to run state-of-the-art AI models locally on CPU, GPU, or NPU — from instant use cases to production deployments.",
        "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality.",
        "The capital of France is Paris.",
    ]
    embeddings = embedder.generate(
        texts=texts, config=EmbeddingConfig(batch_size=len(texts))
    )

    print("\n" + "=" * 80)
    print("GENERATED EMBEDDINGS")
    print("=" * 80)

    for i, (text, embedding) in enumerate(zip(texts, embeddings)):
        print(f"\nText {i+1}:")
        print(f"  Content: {text}")
        print(f"  Embedding shape: {len(embedding)} dimensions")
        print(f"  First 10 elements: {embedding[:10]}")
        print("-" * 70)

    # Generate embedding for query
    query = "what is on device AI"
    print(f"\n" + "=" * 80)
    print("QUERY PROCESSING")
    print("=" * 80)
    print(f"Query: '{query}'")

    query_embedding = embedder.generate(
        texts=[query], config=EmbeddingConfig(batch_size=1)
    )[0]
    print(f"Query embedding shape: {len(query_embedding)} dimensions")

    # Compute inner product between query and all texts
    print(f"\n" + "=" * 80)
    print("SIMILARITY ANALYSIS (Inner Product)")
    print("=" * 80)

    for i, (text, embedding) in enumerate(zip(texts, embeddings)):
        # Convert to numpy arrays for easier computation
        query_vec = np.array(query_embedding)
        text_vec = np.array(embedding)

        # Compute inner product (dot product)
        inner_product = np.dot(query_vec, text_vec)

        print(f"\nText {i+1}:")
        print(f"  Content: {text}")
        print(f"  Inner product with query: {inner_product:.6f}")
        print("-" * 70)


if __name__ == "__main__":
    main()
