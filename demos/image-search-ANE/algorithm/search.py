from dataclasses import dataclass
from pathlib import Path
import time
from io import BytesIO
import requests

import numpy as np
from PIL import Image
import coremltools as ct


from embed_neural import EmbedNeural


# ============================ Helper Functions ============================
def collect_images_from_folder(folder_path):
    IMAGE_EXTENSIONS = {'.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.tiff', '.tif'}
    folder = Path(folder_path)
    
    if not folder.is_dir():
        raise ValueError(f"'{folder_path}' is not a valid directory")
    
    image_paths = []
    for img_path in folder.iterdir():
        if img_path.is_file() and img_path.suffix.lower() in IMAGE_EXTENSIONS:
            image_paths.append(str(img_path))
    
    return image_paths


# ============================ Data Classes ============================
@dataclass
class ImageSearchResult:
    image: Image.Image
    score: float
    path: str = None


# ============================ EmbedNeuralImageSearch ============================
class EmbedNeuralImageSearch:
    def __init__(self, model: EmbedNeural):
        self.model = model

    def _load_image(self, image_path_or_url):
        if image_path_or_url.startswith('http'):
            response = requests.get(image_path_or_url)
            image = Image.open(BytesIO(response.content)).convert('RGB')
        else:
            image = Image.open(image_path_or_url).convert('RGB')
        return image

    def search(self, query, image_paths, metric="inner_product", k=5) -> list[ImageSearchResult]:
        # Benchmark: Text encoding
        text_encode_start = time.time()
        text_embedding = self.model.encode_text(query)
        self.text_encode_time = time.time() - text_encode_start
        
        images = [self._load_image(image_path) for image_path in image_paths]
        
        # Benchmark: Image encoding
        image_embeddings = []
        image_encode_start = time.time()
        for i, image in enumerate(images):
            print(f"Encoding image {i+1}/{len(images)}, from path {image_paths[i]}")
            image_embeddings.append(self.model.encode_image(image))
        self.image_encode_time = time.time() - image_encode_start
        self.avg_image_encode_time = self.image_encode_time / len(images) if images else 0

        if metric == "inner_product":
            scores = [np.dot(text_embedding, image_embedding) for image_embedding in image_embeddings]
        elif metric == "cosine":
            scores = [np.dot(text_embedding, image_embedding) / (np.linalg.norm(text_embedding) * np.linalg.norm(image_embedding)) for image_embedding in image_embeddings]
        elif metric == "l2":
            scores = [-np.linalg.norm(text_embedding - image_embedding) for image_embedding in image_embeddings]
        elif metric == "manhattan":
            scores = [-np.sum(np.abs(text_embedding - image_embedding)) for image_embedding in image_embeddings]
        else:
            raise ValueError(f"Unsupported metric: {metric}, choose from inner_product, cosine, l2, manhattan")
        


        sorted_indices = np.argsort(scores)[::-1][:k]
        
        return [ImageSearchResult(image=images[i], score=scores[i], path=image_paths[i]) for i in sorted_indices]



"""
Example usage:
python search.py \
   --text-model ./modelfiles/build/EmbedNeuralText.mlmodelc \
   --image-model ./modelfiles/build/EmbedNeuralVision.mlmodelc \
   --text "a blue cat" \
   --images https://i.pinimg.com/600x315/21/48/7e/21487e8e0970dd366dafaed6ab25d8d8.jpg \
            https://i.pinimg.com/736x/c9/f2/3e/c9f23e212529f13f19bad5602d84b78b.jpg
"""
def main():
    import argparse
    
    parser = argparse.ArgumentParser(description="Search images using text queries with Embed Neural")
    parser.add_argument(
        "--text",
        type=str,
        required=True,
        help="Text query to search for"
    )
    parser.add_argument(
        "--images",
        type=str,
        nargs="+",
        help="List of image paths or URLs to search through"
    )
    parser.add_argument(
        "--image-folder",
        type=str,
        help="Path to a folder containing images to search through"
    )
    parser.add_argument(
        "--text-model",
        type=str,
        required=True,
        help="Path to the compiled text encoder model (.mlmodelc)"
    )
    parser.add_argument(
        "--image-model",
        type=str,
        required=True,
        help="Path to the compiled image encoder model (.mlmodelc)"
    )
    parser.add_argument(
        "--metric",
        type=str,
        default="inner_product",
        choices=["inner_product", "cosine", "l2", "manhattan"],
        help="Similarity metric to use (default: inner_product)"
    )
    parser.add_argument(
        "--top-k",
        type=int,
        default=5,
        help="Number of top results to return (default: 5)"
    )
    parser.add_argument(
        "--compute-units",
        type=str,
        default="npu",
        choices=["cpu", "npu", "gpu", "all"],
        help="Core ML compute units to use (default: npu)"
    )
    
    args = parser.parse_args()
    
    # Collect image paths
    image_paths = []
    if args.images:
        image_paths.extend(args.images)
    if args.image_folder:
        folder_images = collect_images_from_folder(args.image_folder)
        image_paths.extend(folder_images)
        print(f"Found {len(folder_images)} images in folder '{args.image_folder}'")
    
    if not image_paths:
        print("Error: No images provided. Use --images or --image-folder to specify images.")
        return
    
    print(f"Total images to search: {len(image_paths)}")
    
    compute_units_map = {
        "cpu": ct.ComputeUnit.CPU_ONLY,
        "npu": ct.ComputeUnit.CPU_AND_NE,
        "gpu": ct.ComputeUnit.CPU_AND_GPU,
        "all": ct.ComputeUnit.ALL,
    }

    # Benchmark: Model loading
    print(f"\nLoading models with compute units: {args.compute_units}")
    model_load_start = time.time()
    model = EmbedNeural(
        text_model_path=args.text_model,
        image_model_path=args.image_model,
        compute_units=compute_units_map[args.compute_units],
    )
    model_load_time = time.time() - model_load_start
    print(f"✓ Models loaded in {model_load_time:.3f}s")
    
    searcher = EmbedNeuralImageSearch(model=model)
    
    # Benchmark: Search execution
    search_start = time.time()
    results = searcher.search(
        query=args.text,
        image_paths=image_paths,
        metric=args.metric,
        k=args.top_k
    )
    search_time = time.time() - search_start
    
    # Print benchmark results
    print(f"\n{'='*80}")
    print("BENCHMARK RESULTS")
    print(f"{'='*80}")
    print(f"Model loading time:           {model_load_time:.3f}s")
    print(f"Text encoding time:           {searcher.text_encode_time:.3f}s")
    print(f"Image encoding time (total):  {searcher.image_encode_time:.3f}s")
    print(f"Average per image:            {searcher.avg_image_encode_time:.3f}s")
    print(f"Total search time:            {search_time:.3f}s")
    print(f"{'='*80}\n")
    
    print(f"Search Results for query: '{args.text}'")
    print(f"Metric: {args.metric}\n")
    print("-" * 80)
    
    for idx, result in enumerate(results, 1):
        print(f"{idx}. Score: {result.score:.4f}")
        print(f"   Image size: {result.image.size} ({result.image.mode})")
        print(f"   Path: {result.path}")
        print()


if __name__ == "__main__":
    main()