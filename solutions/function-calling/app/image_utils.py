
import base64
import os
from pathlib import Path
from typing import Union


def image_to_base64(image_path: Union[str, Path]) -> str:
    image_path = Path(image_path)
    if not image_path.exists():
        raise FileNotFoundError(f"Image not exists: {image_path}")
    suffix = image_path.suffix.lower()
    mime_types = {
        '.jpg': 'image/jpeg',
        '.jpeg': 'image/jpeg',
        '.png': 'image/png',
        '.gif': 'image/gif',
        '.webp': 'image/webp',
        '.bmp': 'image/bmp',
    }
    
    if suffix not in mime_types:
        raise ValueError(f"Not support: {suffix}")
    
    mime_type = mime_types[suffix]
    with open(image_path, 'rb') as f:
        image_data = f.read()
    base64_data = base64.b64encode(image_data).decode('utf-8')
    return f"data:{mime_type};base64,{base64_data}"