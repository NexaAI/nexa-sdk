# !/usr/bin/env python3

from dataclasses import dataclass
from typing import Tuple, List, Literal
import coremltools as ct

import os
import sys
sys.path.append(os.path.join(os.path.dirname(__file__), "algorithm"))

from algorithm.embed_neural import EmbedNeural
from algorithm.search import EmbedNeuralImageSearch

# Metrics for distance calculation
metrics=["inner_product", "cosine", "l2", "manhattan"]

IMAGE_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.tiff', '.tif'}

@dataclass
class SearchResult:
    url: str
    score: int = 0
    start: float = 0.0
    end: float = 0.0
    is_image: bool = True
    
AlgorithmType = Literal["embed-neural", "kmeans", "aks", "clip-based"]

class ViewModel:
    def __init__(self):
        self._files = []
        self._top_k = 2  # Default Top-K value
        self._metric = metrics[0]
        self._model = None
    
    @property
    def files(self):
        return self._files

    @files.setter
    def files(self, value):
        self._files = value
        
    def index_files(self, alogrithm: AlgorithmType="embed-neural"):
        # Placeholder for indexing logic
        print(f"Indexing the following files: {self._files}")
        
    def search(self, query: str, alogrithm:AlgorithmType="embed-neural") -> Tuple[List[SearchResult], List[SearchResult]]:
        print(f"[Debug: ] search use {alogrithm}")
        if query is None or query.strip() == "":
            raise ValueError("Query cannot be empty.")
        if self._model is None:
            raise ValueError("Model is not loaded. Please load a model before searching.")
        if self._files is None or len(self._files) == 0:
            raise ValueError("No files have been indexed. Please index files before searching.")
        
        if alogrithm == "embed-neural":
            return self._search_embed_neural(query=query)
        return [], []
    
    def _search_embed_neural(self, query: str) -> Tuple[List[SearchResult], List[SearchResult]]:
        print(f"Searching for query: {query}")
        
        from pathlib import Path
        image_paths = [f for f in self._files if Path(f).suffix.lower() in IMAGE_EXTENSIONS] 
        
        # Placeholder for embed-neural search logic
        searcher = EmbedNeuralImageSearch(model=self._model)
        # Benchmark: Search execution
        import time
        search_start = time.time()
        results = searcher.search(
            query=query,
            image_paths=image_paths,
            metric=self._metric,
            k=self._top_k
        )
        search_time = time.time() - search_start
        
        images = []
        for idx, result in enumerate(results, 1):
            if self._metric == "inner_product" or self._metric == "cosine":
                if result.score < 0:
                    continue
                images.append(SearchResult(url=result.path,score=result.score))
            elif self._metric == "l2" or self._metric == "manhattan":
                images.append(SearchResult(url=result.path,score=-result.score))
        
        return [], images, search_time
    
    # Download CoreML models from HuggingFace
    def download(self, model_name):
        import os
        from huggingface_hub import snapshot_download
        
        print(f"Downloading CoreML models from NexaAI/EmbedNeural-ANE...")
        
        # Get HF token from environment
        hf_token = os.getenv("HUGGINGFACE_HUB_TOKEN") or os.getenv("HF_TOKEN")
        
        try:
            # Download entire repository to HF cache
            cache_dir = snapshot_download(
                repo_id="NexaAI/EmbedNeural-ANE",
                token=hf_token,
                repo_type="model"
            )
            print(f"✓ Models downloaded to: {cache_dir}")
            
        except Exception as e:
            print(f"✗ Download failed: {e}")
            raise

    def load_model(self, model_name):
        import os
        from pathlib import Path
        from huggingface_hub import snapshot_download
        
        # Get HF token from environment
        hf_token = os.getenv("HUGGINGFACE_HUB_TOKEN") or os.getenv("HF_TOKEN")
        
        try:
            print(f"Loading models from HuggingFace cache: NexaAI/EmbedNeural-ANE")
            
            # Get cached models (or download if not present)
            cache_dir = snapshot_download(
                repo_id="NexaAI/EmbedNeural-ANE",
                token=hf_token,
                repo_type="model",
                allow_patterns=["*.mlmodelc/*"]
            )
            
            # Build paths to .mlmodelc folders in cache
            text_model = str(Path(cache_dir) / "EmbedNeuralText.mlmodelc")
            image_model = str(Path(cache_dir) / "EmbedNeuralVision.mlmodelc")
            
            print(f"✓ Model paths resolved:")
            print(f"  Text: {text_model}")
            print(f"  Image: {image_model}")
            
            # Set compute units
            compute_units_map = {
                "cpu": ct.ComputeUnit.CPU_ONLY,
                "npu": ct.ComputeUnit.CPU_AND_NE,
                "gpu": ct.ComputeUnit.CPU_AND_GPU,
                "all": ct.ComputeUnit.ALL,
            }
            compute_units = "npu"
            print(f"Loading models with compute units: {compute_units}")
            
            # Load EmbedNeural model
            self._model = EmbedNeural(
                text_model_path=text_model,
                image_model_path=image_model,
                compute_units=compute_units_map[compute_units],
            )
            
            print("✓ EmbedNeural models loaded successfully!")
            
        except Exception as e:
            print(f"✗ Model loading failed: {e}")
            raise
    
    def unload_model(self):
        self._model = None
        
    def update_top_k(self, top_k: int):
        self._top_k = top_k
        
    def update_metric(self, metric: str):
        self._metric = metric