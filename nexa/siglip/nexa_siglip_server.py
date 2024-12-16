from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from fastapi import Request
from fastapi.responses import HTMLResponse
import uvicorn
import os
import socket
from PIL import Image
import torch
from transformers import AutoProcessor, AutoModel

app = FastAPI(title="Nexa AI SigLIP Image-Text Matching Service")

# Global variables
hostname = socket.gethostname()
siglip_model = None
siglip_processor = None
images_dict = {}

class ImagePathRequest(BaseModel):
    image_dir: str

class SearchResponse(BaseModel):
    image_path: str
    similarity_score: float

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
    """Initialize model when service starts"""
    init_model()

@app.get("/", response_class=HTMLResponse, tags=["Root"])
async def read_root(request: Request):
    return HTMLResponse(
        content=f"<h1>Welcome to Nexa AI SigLIP Image-Text Matching Service</h1><p>Hostname: {hostname}</p>"
    )


@app.post("/v1/load_images")
async def load_images(request: ImagePathRequest):
    """Load images from specified directory"""
    global images_dict
    try:
        images_dict = load_images_from_directory(request.image_dir)
        return {
            "message": f"Successfully loaded {len(images_dict)} images",
            "images": list(images_dict.keys())
        }
    except Exception as e:
        raise HTTPException(status_code=400, detail=str(e))

@app.post("v1/find_similar", response_model=SearchResponse)
async def find_similar(text: str):
    """Find image most similar to input text"""
    if not images_dict:
        raise HTTPException(status_code=400, detail="No images available, please load images first")
    
    try:
        
        image_paths = list(images_dict.keys())
        images = list(images_dict.values())
        
        inputs = siglip_processor(text=[text], images=images, padding="max_length", return_tensors="pt")
        
        with torch.no_grad():
            outputs = siglip_model(**inputs)
        
        logits_per_image = outputs.logits_per_image
        probs = torch.sigmoid(logits_per_image)
        max_prob_index = torch.argmax(probs).item()
        max_prob = probs[max_prob_index][0].item()
        
        return SearchResponse(
            image_path=image_paths[max_prob_index],
            similarity_score=max_prob
        )
    except Exception as e:
        raise HTTPException(status_code=500, detail=f"Error processing request: {str(e)}")

@app.get("/v1/list_images")
async def list_images():
    """List all loaded images"""
    return {"images": list(images_dict.keys())}



if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)