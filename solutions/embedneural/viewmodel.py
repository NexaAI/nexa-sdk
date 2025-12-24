# Copyright 2024-2025 Nexa AI, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# !/usr/bin/env python3

from dataclasses import dataclass
from typing import Tuple, List
from pathlib import Path

from search import NexaImageSearch, IMAGE_EXTENSIONS

# Metrics for distance calculation
metrics = ["l2"]

@dataclass
class SearchResult:
    url: str
    score: float = 0.0
    start: float = 0.0
    end: float = 0.0
    is_image: bool = True


class ViewModel:
    """ViewModel for image search using Nexa API."""
    
    def __init__(self, base_url: str = "http://localhost:18181", model: str = "NexaAI/EmbedNeural"):
        """
        Initialize ViewModel.
        
        Args:
            base_url: Base URL of nexa serve API
            model: Model name to use for embeddings
        """
        self._files = []
        self._top_k = 2  # Default Top-K value
        self._metric = metrics[0]
        self._searcher = NexaImageSearch(base_url=base_url, model=model)
    
    @property
    def files(self):
        return self._files

    @files.setter
    def files(self, value):
        # Clear cache when files change to avoid stale results
        if value != self._files:
            self._searcher.clear_cache()
        self._files = value
        
    def index_files(self) -> None:
        """
        Calculate embeddings for all image files.
        This should be called when user clicks the Index button.
        """
        if not self._files:
            print("No files to index.")
            return
        
        # Filter to only image files
        image_paths = [
            f for f in self._files 
            if Path(f).suffix.lower() in IMAGE_EXTENSIONS
        ]
        
        if not image_paths:
            print("No image files found to index.")
            return
        
        print(f"Indexing {len(image_paths)} image files...")
        self._searcher.index_images(image_paths)
        
    def search(self, query: str) -> Tuple[List[SearchResult], List[SearchResult], float]:
        """
        Search images using text query.
        
        Args:
            query: Text query string
            
        Returns:
            Tuple of (empty_list, images, search_time)
            First element is kept for compatibility but always empty
        """
        import time
        
        if query is None or query.strip() == "":
            raise ValueError("Query cannot be empty.")
        
        if not self._searcher._image_embeddings:
            raise ValueError("No images indexed. Please index images first.")
        
        # Perform search
        search_start = time.time()
        results = self._searcher.search(
            query=query,
            metric=self._metric,
            k=self._top_k
        )
        search_time = time.time() - search_start
        
        # Convert to SearchResult format
        images = [
            SearchResult(url=result.path, score=result.score)
            for result in results
        ]
        
        return [], images, search_time
    
    def update_top_k(self, top_k: int):
        """Update top-k value for search results."""
        self._top_k = top_k
        
    def update_metric(self, metric: str):
        """Update distance metric."""
        if metric not in metrics:
            raise ValueError(f"Unsupported metric: {metric}. Supported: {metrics}")
        self._metric = metric

