#!/usr/bin/env python3

"""
NexaAI Rerank Example - Document Reranking

This example demonstrates how to use the NexaAI SDK to rerank documents based on a query.
It includes basic model initialization, document reranking, and score analysis.
"""

import os
from nexaai.rerank import Reranker, RerankConfig


def main():
    model_path = os.path.expanduser("~/.cache/nexa.ai/nexa_sdk/models/NexaAI/jina-v2-rerank-mlx/jina-reranker-v2-base-multilingual-f16.safetensors")
    
    # For now, this modality is only supported in MLX.
    reranker: Reranker = Reranker.from_(name_or_path=model_path, plugin_id="mlx")
    documents = [
        "On-device AI is a type of AI that is processed on the device itself, rather than in the cloud.",
        "edge computing",
        "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality.",
        "The capital of France is Paris."
    ]

    query = "Where is on-device AI?"

    scores = reranker.rerank(query=query, documents=documents, config=RerankConfig(batch_size=len(documents)))

    print(f"Query: {query}")
    print(f"Documents: {len(documents)} documents")
    print("-" * 50)
    for i, score in enumerate(scores):
        print(f"[{score:.4f}] : {documents[i]}")


if __name__ == "__main__":
    main()
