# !/usr/bin/env python3

from dataclasses import dataclass
from typing import List
import numpy as np
from pathlib import Path

from nexa_client import NexaClient

IMAGE_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.tiff', '.tif'}


@dataclass
class ImageSearchResult:
    """Result of image search."""
    path: str
    score: float


class NexaImageSearch:
    """Image search using Nexa API with L2 distance."""
    
    def __init__(self, base_url: str = "http://localhost:18181", model: str = "NexaAI/EmbedNeural"):
        """
        Initialize Nexa image search.
        
        Args:
            base_url: Base URL of nexa serve API
            model: Model name to use for embeddings
        """
        self.client = NexaClient(base_url=base_url, model=model)
        self._image_embeddings = {}  # Cache: {image_path: embedding_vector}
    
    def index_images(self, image_paths: List[str]) -> None:
        """
        Calculate and cache embeddings for all images.
        
        Args:
            image_paths: List of image file paths
        """
        if not image_paths:
            return
        
        print(f"Indexing {len(image_paths)} images...")
        
        # Clear old cache before indexing new images
        self._image_embeddings = {}
        
        # Filter to only image files
        valid_image_paths = [
            path for path in image_paths 
            if Path(path).suffix.lower() in IMAGE_EXTENSIONS
        ]
        
        if not valid_image_paths:
            print("No valid image files found.")
            return
        
        # Calculate embeddings for all images (one at a time)
        try:
            for i, image_path in enumerate(valid_image_paths):
                print(f"Indexing image {i+1}/{len(valid_image_paths)}: {image_path}")
                # Request embedding for single image (API only supports one at a time)
                embedding = self.client.get_embedding(image_path)
                self._image_embeddings[image_path] = embedding
            
            print(f"✓ Successfully indexed {len(valid_image_paths)} images")
            
        except Exception as e:
            print(f"✗ Error indexing images: {e}")
            raise
    
    def search(self, query: str, metric: str = "l2", k: int = 5) -> List[ImageSearchResult]:
        """
        Search images using text query with L2 distance.
        
        Args:
            query: Text query string
            metric: Distance metric (only "l2" is supported for now)
            k: Number of top results to return
            
        Returns:
            List of ImageSearchResult sorted by similarity (best first)
        """
        if not self._image_embeddings:
            raise ValueError("No images indexed. Please index images first.")
        
        if not query or not query.strip():
            raise ValueError("Query cannot be empty.")
        
        # Calculate query embedding
        print(f"Calculating embedding for query: {query}")
        query_embedding = self.client.get_embedding(query)
        
        # Calculate distances
        distances = []
        for image_path, image_embedding in self._image_embeddings.items():
            if metric == "l2":
                distance = np.linalg.norm(query_embedding - image_embedding)
            else:
                raise ValueError(f"Unsupported metric: {metric}. Only 'l2' is supported.")
            
            distances.append((image_path, distance))
        
        # Sort by distance (smaller is better) and get top-k
        distances.sort(key=lambda x: x[1])
        top_k_results = distances[:k]
        
        # Convert to ImageSearchResult (use negative distance as score for consistency with UI)
        results = [
            ImageSearchResult(path=path, score=-distance)
            for path, distance in top_k_results
        ]
        
        return results
    
    def clear_cache(self):
        """Clear cached embeddings."""
        self._image_embeddings = {}

