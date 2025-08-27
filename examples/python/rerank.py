#!/usr/bin/env python3

"""
NexaAI Rerank Example - Document Reranking

This example demonstrates how to use the NexaAI SDK to rerank documents based on a query.
It includes basic model initialization, document reranking, and score analysis.
"""

import os
from nexaai.rerank import Reranker, RerankConfig

def main():
    model_path = os.path.expanduser("~/.cache/nexa.ai/nexa_sdk/models/nexaml/jina-v2-rerank-mlx/jina-reranker-v2-base-multilingual-f16.safetensors")
    reranker: Reranker = Reranker.from_(name_or_path=model_path, plugin_id="mlx")
    documents = [
        "Machine learning is a subset of artificial intelligence.",
        "Machine learning algorithms learn patterns from data.",
        "The weather is sunny today.",
        "Deep learning is a type of machine learning."
    ]
    
    query = "What is machine learning?"

    scores = reranker.rerank(query=query, documents=documents, config=RerankConfig(batch_size=len(documents)))
    
    print(f"Query: {query}")
    print(f"Documents: {len(documents)} documents")
    print("-" * 50)
    for i, score in enumerate(scores):
        print(f"[{score:.4f}] : {documents[i]}")

if __name__ == "__main__":
    main()


