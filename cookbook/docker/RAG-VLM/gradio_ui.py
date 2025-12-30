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

import os
import json
import base64
import tempfile
import threading
import time
from typing import Optional, Generator, Tuple
from pathlib import Path

import cv2
import gradio as gr
import requests
import numpy as np
from PIL import Image

# Configuration
DEFAULT_MODEL = "NexaAI/AutoNeural"
DEFAULT_ENDPOINT = "http://127.0.0.1:18181"
FRAME_INTERVAL_SECONDS = 8  # Extract frame every 8 seconds
CLIP_LENGTH_SECONDS = 8  # Process 8-second clips


def image_to_base64(image_path: str) -> str:
    """Convert image file to base64 data URI."""
    with open(image_path, 'rb') as f:
        image_data = f.read()
    
    # Detect MIME type
    ext = Path(image_path).suffix.lower()
    mime_types = {
        '.jpg': 'image/jpeg',
        '.jpeg': 'image/jpeg',
        '.png': 'image/png',
        '.gif': 'image/gif',
        '.webp': 'image/webp',
        '.bmp': 'image/bmp',
    }
    mime_type = mime_types.get(ext, 'image/jpeg')
    
    base64_data = base64.b64encode(image_data).decode('utf-8')
    return f"data:{mime_type};base64,{base64_data}"


def call_nexa_chat(model: str, messages: list, endpoint: str, stream: bool = False) -> Generator[str, None, None]:
    """
    Call Nexa /v1/chat/completions endpoint.
    
    Args:
        model: Model name
        messages: List of chat messages
        endpoint: Base URL for API endpoint
        stream: If True, yields text pieces; if False, returns full text
        
    Yields:
        str: Text pieces (if stream=True) or full text (if stream=False)
    """
    url = endpoint.rstrip("/") + "/v1/chat/completions"
    headers = {"Content-Type": "application/json"}
    payload = {
        "model": model,
        "messages": messages,
        "stream": stream,
        "max_tokens": 512,
        "enable_think": False
    }
    
    if not stream:
        # Non-streaming: return complete text
        try:
            resp = requests.post(url, headers=headers, data=json.dumps(payload), timeout=300)
            resp.raise_for_status()
            data = resp.json()
            try:
                yield data["choices"][0]["message"]["content"]
            except (KeyError, IndexError):
                yield data.get("text", "") or data.get("response", "")
        except Exception as e:
            yield f"Error: {str(e)}"
        return
    
    # Streaming mode: yield text pieces via SSE
    try:
        with requests.post(url, headers=headers, data=json.dumps(payload), stream=True, timeout=300) as resp:
            resp.raise_for_status()
            for raw in resp.iter_lines(decode_unicode=True):
                if not raw:
                    continue
                if raw.startswith("data:"):
                    chunk = raw[len("data:"):].strip()
                    if chunk == "[DONE]":
                        break
                    try:
                        obj = json.loads(chunk)
                        choices = obj.get("choices", [])
                        if choices:
                            delta = choices[0].get("delta") or {}
                            piece = delta.get("content", "")
                            if piece:
                                yield piece
                    except json.JSONDecodeError:
                        continue
    except Exception as e:
        yield f"Error: {str(e)}"


def extract_frames_from_video(video_path: str, interval_seconds: float = FRAME_INTERVAL_SECONDS) -> list:
    """
    Extract frames from video at fixed intervals.
    
    Args:
        video_path: Path to video file
        interval_seconds: Interval between frames in seconds
        
    Returns:
        List of (frame_image, timestamp_seconds) tuples
    """
    cap = cv2.VideoCapture(video_path)
    if not cap.isOpened():
        return []
    
    fps = cap.get(cv2.CAP_PROP_FPS)
    if fps <= 0:
        cap.release()
        return []
    
    frame_interval = int(fps * interval_seconds)
    frames = []
    frame_count = 0
    
    while True:
        ret, frame = cap.read()
        if not ret:
            break
        
        # Extract frame at intervals
        if frame_count % frame_interval == 0:
            timestamp = frame_count / fps
            # Convert BGR to RGB for display
            frame_rgb = cv2.cvtColor(frame, cv2.COLOR_BGR2RGB)
            frames.append((frame_rgb, timestamp))
        
        frame_count += 1
    
    cap.release()
    return frames


def process_video_stream(
    video_file: Optional[str],
    model: str,
    endpoint: str,
    prompt: str,
    stop_event: threading.Event
) -> Generator[Tuple[Optional[np.ndarray], str], None, None]:
    """
    Process video frames and call API for inference.
    
    Args:
        video_file: Path to uploaded video file
        model: Model name
        endpoint: API endpoint base URL
        prompt: User prompt for inference
        stop_event: Threading event to signal stop
        
    Yields:
        Tuple of (current_frame_image, accumulated_text_results)
    """
    if not video_file:
        yield None, "Please upload a video file."
        return
    
    if not os.path.exists(video_file):
        yield None, f"Video file not found: {video_file}"
        return
    
    # Extract frames
    frames = extract_frames_from_video(video_file, FRAME_INTERVAL_SECONDS)
    
    if not frames:
        yield None, "Failed to extract frames from video. Please check the video file."
        return
    
    accumulated_results = []
    temp_dir = tempfile.mkdtemp()
    
    try:
        for idx, (frame_image, timestamp) in enumerate(frames):
            if stop_event.is_set():
                yield frame_image, "\n\n".join(accumulated_results)
                return
            
            # Save frame as temporary image
            frame_path = os.path.join(temp_dir, f"frame_{idx:04d}.jpg")
            frame_pil = Image.fromarray(frame_image)
            frame_pil.save(frame_path, "JPEG")
            
            # Convert to base64
            image_data_uri = image_to_base64(frame_path)
            
            # Build messages for API call
            messages = [
                {
                    "role": "user",
                    "content": [
                        {"type": "text", "text": prompt},
                        {"type": "image_url", "image_url": {"url": image_data_uri}}
                    ]
                }
            ]
            
            # Call API (non-streaming for simplicity)
            result_text = ""
            for piece in call_nexa_chat(model, messages, endpoint, stream=False):
                result_text += piece
            
            # Format result with timestamp
            result_entry = f"**Frame at {timestamp:.1f}s:**\n{result_text}\n"
            accumulated_results.append(result_entry)
            
            # Yield current frame and accumulated results
            yield frame_image, "\n\n".join(accumulated_results)
            
            # Small delay to allow UI update
            time.sleep(0.1)
    
    finally:
        # Cleanup temporary files
        import shutil
        shutil.rmtree(temp_dir, ignore_errors=True)


# Global state for stop event
stop_processing = threading.Event()


def start_processing(video_file: Optional[str], model: str, endpoint: str, prompt: str):
    """Start video processing."""
    stop_processing.clear()
    return process_video_stream(video_file, model, endpoint, prompt, stop_processing)


def stop_processing_func():
    """Stop video processing."""
    stop_processing.set()
    return None, "Processing stopped by user."


# UI
CUSTOM_CSS = """
.gradio-container { max-width: 1600px !important; }
.rounded-card { border-radius: 16px; box-shadow: 0 1px 8px rgba(0,0,0,.06); background: white; }
.pad { padding: 14px; }
"""

with gr.Blocks(title="AutoNeural Video Inference", css=CUSTOM_CSS) as demo:
    gr.Markdown("# AutoNeural Video Inference Demo")
    gr.Markdown("Upload a video to extract frames at 8-second intervals and get AI analysis using AutoNeural model.")
    
    with gr.Row():
        # Left column: Video display and controls
        with gr.Column(scale=1, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Video & Settings**")
            
            video_input = gr.Video(label="Upload Video", type="filepath")
            
            with gr.Accordion("Model Settings", open=False):
                model_name = gr.Textbox(
                    label="Model",
                    value=DEFAULT_MODEL,
                    info="AutoNeural model name"
                )
                endpoint_url = gr.Textbox(
                    label="Endpoint",
                    value=DEFAULT_ENDPOINT,
                    info="Nexa serve endpoint URL"
                )
            
            prompt_input = gr.Textbox(
                label="Prompt",
                value="Describe what you see in this image in detail.",
                placeholder="Enter your prompt for image analysis...",
                lines=3
            )
            
            with gr.Row():
                btn_start = gr.Button("Start Processing", variant="primary")
                btn_stop = gr.Button("Stop", variant="stop")
        
        # Right column: Results display
        with gr.Column(scale=1, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Current Frame**")
            current_frame = gr.Image(label="Current Frame", type="numpy", height=400)
            
            gr.Markdown("**Inference Results**")
            results_text = gr.Markdown(
                value="Results will appear here as frames are processed...",
                label="Results"
            )
    
    # Event handlers
    def process_wrapper(video, model, endpoint, prompt):
        """Wrapper to handle processing with UI updates."""
        if not video:
            return None, "Please upload a video file first."
        
        results_accumulated = ""
        frame_img = None
        for frame_img, results in process_video_stream(video, model, endpoint, prompt, stop_processing):
            if stop_processing.is_set():
                break
            results_accumulated = results
            yield frame_img, results_accumulated
        
        return frame_img, results_accumulated
    
    btn_start.click(
        fn=process_wrapper,
        inputs=[video_input, model_name, endpoint_url, prompt_input],
        outputs=[current_frame, results_text]
    )
    
    btn_stop.click(
        fn=stop_processing_func,
        inputs=[],
        outputs=[current_frame, results_text]
    )

if __name__ == "__main__":
    demo.launch(server_name="0.0.0.0", server_port=7860)

