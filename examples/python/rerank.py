#!/usr/bin/env python3

"""
NexaAI Rerank Example - Document Reranking

This example demonstrates how to use the NexaAI SDK to rerank documents based on a query.
It includes basic model initialization, document reranking, and score analysis.
"""

import argparse
import os
from nexaai.rerank import Reranker, RerankConfig


def main():
    parser = argparse.ArgumentParser(description="NexaAI Rerank Example")
    parser.add_argument("--model", 
                       default="~/.cache/nexa.ai/nexa_sdk/models/NexaAI/jina-v2-rerank-mlx/jina-reranker-v2-base-multilingual-f16.safetensors",
                       help="Path to the rerank model")
    parser.add_argument("--query", default="Where is on-device AI?",
                       help="Query text for reranking")
    parser.add_argument("--documents", nargs="+",
                       default=["On-device AI is a type of AI that is processed on the device itself, rather than in the cloud.",
                               "edge computing",
                               "A ragdoll is a breed of cat that is known for its long, flowing hair and gentle personality.",
                               "The capital of France is Paris."],
                       help="Documents to rerank")
    parser.add_argument("--batch-size", type=int, help="Batch size for processing")
    parser.add_argument("--plugin-id", default="cpu_gpu", help="Plugin ID to use")
    args = parser.parse_args()

    model_path = os.path.expanduser(args.model)
    reranker = Reranker.from_(name_or_path=model_path, plugin_id=args.plugin_id)
    
    batch_size = args.batch_size or len(args.documents)
    scores = reranker.rerank(query=args.query, documents=args.documents, 
                           config=RerankConfig(batch_size=batch_size))

    print(f"Query: {args.query}")
    print(f"Documents: {len(args.documents)} documents")
    print("-" * 50)
    for i, score in enumerate(scores):
        print(f"[{score:.4f}] : {args.documents[i]}")


if __name__ == "__main__":
    main()
