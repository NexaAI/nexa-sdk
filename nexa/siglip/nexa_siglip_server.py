from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from fastapi import Request
from fastapi.responses import HTMLResponse
from fastapi.middleware.cors import CORSMiddleware
import uvicorn
import os
import socket
import time
import argparse
from PIL import Image
import torch
from transformers import AutoProcessor, AutoModel
import base64
from io import BytesIO

app = FastAPI(title="Nexa AI SigLIP Image-Text Matching Service")
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Allows all origins
    allow_credentials=True,
    allow_methods=["*"],  # Allows all methods
    allow_headers=["*"],  # Allows all headers
)

# Global variables
hostname = socket.gethostname()
siglip_model = None
siglip_processor = None
images_dict = {}

class ImagePathRequest(BaseModel):
    image_dir: str

class SearchResponse(BaseModel):
    image_path: str
    image_base64: str
    similarity_score: float
    latency: float

def init_model():
    """Initialize SigLIP model and processor"""
    global siglip_model, siglip_processor
    siglip_model = AutoModel.from_pretrained("google/siglip-base-patch16-384")
    siglip_processor = AutoProcessor.from_pretrained("google/siglip-base-patch16-384")

def load_images_from_directory(image_dir, valid_extensions=('.jpg', '.jpeg', '.png', '.webp')):
    """Load images from directory"""
    images_dict = {}
    
    if not os.path.exists(image_dir):
        raise ValueError(f"Directory {image_dir} does not exist")
        
    for filename in os.listdir(image_dir):
        if filename.lower().endswith(valid_extensions):
            image_path = os.path.join(image_dir, filename)
            try:
                image = Image.open(image_path).convert("RGB")
                images_dict[image_path] = image
            except Exception as e:
                print(f"Failed to load image {filename}: {str(e)}")
                
    if not images_dict:
        raise ValueError(f"No valid image files found in {image_dir}")
        
    return images_dict

@app.on_event("startup")
async def startup_event():
    """Initialize model and load images when service starts"""
    init_model()
    # Add image loading if image_dir is provided
    if hasattr(app, "image_dir") and app.image_dir:
        global images_dict
        try:
            images_dict = load_images_from_directory(app.image_dir)
            print(f"Successfully loaded {len(images_dict)} images from {app.image_dir}")
        except Exception as e:
            print(f"Failed to load images: {str(e)}")

@app.get("/", response_class=HTMLResponse, tags=["Root"])
async def read_root(request: Request):
    return HTMLResponse(
        content=f"<h1>Welcome to Nexa AI SigLIP Image-Text Matching Service</h1><p>Hostname: {hostname}</p>"
    )

@app.get("/v1/list_images")
async def list_images():
    """Return current image directory path and loaded images"""
    current_dir = getattr(app, "image_dir", None)
    return {
        "image_dir": current_dir,
        "images_count": len(images_dict),
        "images": list(images_dict.keys()),
        "status": "active" if current_dir and images_dict else "no_images_loaded"
    }

@app.post("/v1/load_images")
async def load_images(request: ImagePathRequest):
    """Load images from specified directory, replacing any previously loaded images"""
    global images_dict
    try:
        temp_images = load_images_from_directory(request.image_dir)
        
        if not temp_images:
            raise ValueError("No valid images found in the specified directory")
            
        images_dict.clear()
        images_dict.update(temp_images)
        app.image_dir = request.image_dir
        
        return {
            "message": f"Successfully loaded {len(images_dict)} images from {request.image_dir}",
            "images": list(images_dict.keys())
        }
    except Exception as e:
        current_count = len(images_dict)
        error_message = f"Failed to load images: {str(e)}. Keeping existing {current_count} images."
        raise HTTPException(status_code=400, detail=error_message)

@app.post("/v1/find_similar", response_model=SearchResponse)
async def find_similar(text: str):
    """Find image most similar to input text"""
    if not images_dict:
        raise HTTPException(status_code=400, detail="No images available, please load images first")
    
    try:
        start_time = time.time()
        image_paths = list(images_dict.keys())
        images = list(images_dict.values())
        
        inputs = siglip_processor(text=[text], images=images, padding="max_length", return_tensors="pt")
        
        with torch.no_grad():
            outputs = siglip_model(**inputs)
        
        logits_per_image = outputs.logits_per_image
        probs = torch.sigmoid(logits_per_image)
        max_prob_index = torch.argmax(probs).item()
        max_prob = probs[max_prob_index][0].item()
        
        # Convert the PIL Image to base64
        matched_image = images[max_prob_index]
        buffered = BytesIO()
        matched_image.save(buffered, format="JPEG")
        img_str = "data:image/jpeg;base64," + base64.b64encode(buffered.getvalue()).decode()
        
        return SearchResponse(
            image_path=image_paths[max_prob_index],
            image_base64=img_str,
            similarity_score=max_prob,
            latency=round(time.time() - start_time, 3)
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error processing request: {str(e)}")


def run_nexa_ai_siglip_service(**kwargs):
    host = kwargs.get("host", "localhost")
    port = kwargs.get("port", 8100)
    reload = kwargs.get("reload", False)
    if kwargs.get("image_dir"):
        app.image_dir = kwargs.get("image_dir")
    uvicorn.run(app, host=host, port=port, reload=reload)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Run the Nexa AI SigLIP Service"
    )
    parser.add_argument(
        "--image_dir", type=str, help="Directory of images to load"
    )
    parser.add_argument(
        "--host", type=str, default="localhost", help="Host to bind the server to"
    )
    parser.add_argument(
        "--port", type=int, default=8100, help="Port to bind the server to"
    )
    parser.add_argument(
        "--reload", type=bool, default=False, help="Reload the server on code changes"
    )
    args = parser.parse_args()
    run_nexa_ai_siglip_service(
        image_dir=args.image_dir,
        host=args.host,
        port=args.port,
        reload=args.reload
    )