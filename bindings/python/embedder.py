#!/usr/bin/env python3

"""
NexaAI Embedding Example - Text Embedding Generation

This example demonstrates how to use the NexaAI SDK to generate text embeddings.
It includes basic model initialization, single and batch embedding generation, and embedding analysis.
"""

import argparse
import os
import numpy as np

from nexaai.embedder import Embedder, EmbeddingConfig

def main():
    parser = argparse.ArgumentParser(description="NexaAI Embedding Example")
    parser.add_argument("--model", default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/jina-v2-fp16-mlx/model.safetensors",
                       help="Path to the embedding model")
    parser.add_argument("--texts", nargs="+", 
                       default=["On-device AI is a type of AI that is processed on the device itself, rather than in the cloud.",
                               "Nexa AI allows you to run state-of-the-art AI models locally on CPU, GPU, or NPU — from instant use cases to production deployments.",
                               "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality.",
                               "The capital of France is Paris."],
                       help="Texts to embed")
    parser.add_argument("--query", default="what is on device AI",
                       help="Query text for similarity analysis")
    parser.add_argument("--batch-size", type=int, help="Batch size for processing")
    parser.add_argument("--plugin-id", default="cpu_gpu", help="Plugin ID to use")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    embedder = Embedder.from_(name_or_path=model_path, plugin_id=args.plugin_id)
    print('Embedder loaded successfully!')

    dim = embedder.get_embedding_dim()
    print(f"Dimension: {dim}")

    batch_size = args.batch_size or len(args.texts)
    embeddings = embedder.generate(
        texts=args.texts, config=EmbeddingConfig(batch_size=batch_size))
    
    print("\n" + "="*80)
    print("GENERATED EMBEDDINGS")
    print("="*80)
    
    for i, (text, embedding) in enumerate(zip(args.texts, embeddings)):
        print(f"\nText {i+1}:")
        print(f"  Content: {text}")
        print(f"  Embedding shape: {len(embedding)} dimensions")
        print(f"  First 10 elements: {embedding[:10]}")
        print("-" * 70)

    print(f"\n" + "="*80)
    print("QUERY PROCESSING")
    print("="*80)
    print(f"Query: '{args.query}'")
    
    query_embedding = embedder.generate(
        texts=[args.query], config=EmbeddingConfig(batch_size=1))[0]
    print(f"Query embedding shape: {len(query_embedding)} dimensions")
    
    print(f"\n" + "="*80)
    print("SIMILARITY ANALYSIS (Inner Product)")
    print("="*80)
    
    for i, (text, embedding) in enumerate(zip(args.texts, embeddings)):
        query_vec = np.array(query_embedding)
        text_vec = np.array(embedding)
        inner_product = np.dot(query_vec, text_vec)
        
        print(f"\nText {i+1}:")
        print(f"  Content: {text}")
        print(f"  Inner product with query: {inner_product:.6f}")
        print("-" * 70)


if __name__ == "__main__":
    main()
