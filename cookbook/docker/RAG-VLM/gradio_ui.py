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
import base64
import shutil
import tempfile
import threading
import time
from typing import Optional, Generator, Tuple
from pathlib import Path

import gradio as gr
import requests
import numpy as np
from PIL import Image
import imageio

# Configuration
DEFAULT_MODEL = "NexaAI/AutoNeural"
DEFAULT_ENDPOINT = "http://127.0.0.1:18181"
DEFAULT_FRAME_INTERVAL = 8  # Extract frame every 8 seconds


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


def call_nexa_chat(model: str, messages: list, endpoint: str) -> str:
    """
    Call Nexa /v1/chat/completions endpoint (non-streaming).
    
    Args:
        model: Model name
        messages: List of chat messages
        endpoint: Base URL for API endpoint
        
    Returns:
        Response text content
    """
    url = f"{endpoint.rstrip('/')}/v1/chat/completions"
    payload = {
        "model": model,
        "messages": messages,
        "stream": False,
        "max_tokens": 512,
        "enable_think": False
    }
    
    try:
        resp = requests.post(url, json=payload, timeout=300)
        resp.raise_for_status()
        data = resp.json()
        return data.get("choices", [{}])[0].get("message", {}).get("content", "") or data.get("text", "") or data.get("response", "")
    except Exception as e:
        return f"Error: {str(e)}"


def extract_first_frame(video_path: str) -> Optional[np.ndarray]:
    """Extract the first frame from video."""
    try:
        with imageio.get_reader(video_path) as reader:
            frame = reader.get_data(0)  # type: ignore
            return frame if isinstance(frame, np.ndarray) else np.array(frame)
    except Exception as e:
        print(f"Error extracting first frame: {e}")
        return None


def extract_frames_from_video(video_path: str, interval_seconds: float = DEFAULT_FRAME_INTERVAL) -> list:
    """Extract frames from video at fixed intervals."""
    try:
        with imageio.get_reader(video_path) as reader:
            metadata = reader.get_meta_data()  # type: ignore
            fps = metadata.get('fps', 30.0)
            
            if fps <= 0:
                return []
            
            frame_interval = int(fps * interval_seconds)
            frames = []
            
            for frame_count, frame in enumerate(reader):  # type: ignore
                if frame_count % frame_interval == 0:
                    timestamp = frame_count / fps
                    frame_array = frame if isinstance(frame, np.ndarray) else np.array(frame)
                    frames.append((frame_array, timestamp))
            
            return frames
    except Exception as e:
        print(f"Error reading video: {e}")
        return []


def process_video_stream(
    video_file: Optional[str],
    model: str,
    endpoint: str,
    prompt: str,
    stop_event: threading.Event,
    interval_seconds: float = DEFAULT_FRAME_INTERVAL
) -> Generator[Tuple[Optional[np.ndarray], str], None, None]:
    """Process video frames and call API for inference."""
    if not video_file or not os.path.exists(video_file):
        yield None, f"Video file not found: {video_file}" if video_file else "Please upload a video file."
        return
    
    frames = extract_frames_from_video(video_file, interval_seconds)
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
            
            # Save frame and convert to base64
            frame_path = os.path.join(temp_dir, f"frame_{idx:04d}.jpg")
            Image.fromarray(frame_image).save(frame_path, "JPEG")
            image_data_uri = image_to_base64(frame_path)
            
            # Call API
            messages = [{
                "role": "user",
                "content": [
                    {"type": "text", "text": prompt},
                    {"type": "image_url", "image_url": {"url": image_data_uri}}
                ]
            }]
            
            result_text = call_nexa_chat(model, messages, endpoint)
            result_entry = f"**Frame at {timestamp:.1f}s:**\n{result_text}\n"
            accumulated_results.append(result_entry)
            
            yield frame_image, "\n\n".join(accumulated_results)
            time.sleep(0.1)
    
    finally:
        shutil.rmtree(temp_dir, ignore_errors=True)


# Global state for stop event
stop_processing = threading.Event()


# UI
CUSTOM_CSS = """
/* Make cards cleaner and add subtle shadows */
.gradio-container { max-width: 1600px !important; }
.rounded-card { border-radius: 16px; box-shadow: 0 1px 8px rgba(0,0,0,.06); background: white; }
.pad { padding: 14px; }
.section-title { font-weight: 700; font-size: 14px; opacity: .8; margin-bottom: 8px; }
#info-panel .gallery { background: #101114; } /* darker bg for images */
"""

with gr.Blocks(title="AutoNeural Video Inference") as demo:
    gr.Markdown("## AutoNeural Video Inference Demo")
    gr.Markdown("Upload a video to extract frames at 8-second intervals and get AI analysis using AutoNeural model.")
    
    with gr.Row():
        # Left column: settings / video upload
        with gr.Column(scale=1, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Video & Settings**")
            
            video_input = gr.Video(
                label="Upload Video"
            )
            
            with gr.Accordion("Model Settings", open=False):
                model_name = gr.Textbox(
                    label="Model Name",
                    value=DEFAULT_MODEL,
                    info="AutoNeural model name"
                )
                endpoint_url = gr.Textbox(
                    label="Endpoint URL",
                    value=f"{DEFAULT_ENDPOINT}/v1/chat/completions",
                    info="Nexa serve endpoint URL (full path)"
                )
                frame_interval = gr.Slider(
                    1, 30, value=DEFAULT_FRAME_INTERVAL, step=1,
                    label="Frame Interval (seconds)",
                    info="Extract frame every N seconds"
                )
            
            prompt_input = gr.Textbox(
                label="Prompt",
                value="Describe what you see in this image in detail.",
                placeholder="Enter your prompt for image analysis...",
                lines=3
            )
            
            with gr.Row():
                btn_start = gr.Button("Start Processing", variant="primary")
                btn_stop = gr.Button("Stop Processing", variant="stop")
            
            status = gr.Textbox(label="Status", value="", interactive=False)
        
        # Middle column: video display
        with gr.Column(scale=2, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Current Frame**")
            current_frame = gr.Image(
                label="Current Frame",
                type="numpy",
                height=600,
                show_label=False
            )
        
        # Right column: information panel (results)
        with gr.Column(scale=2):
            with gr.Accordion("Inference Results", open=True):
                results_text = gr.Markdown(
                    value="Results will appear here as frames are processed...",
                    label="Results"
                )
    
    # Event handlers
    def on_video_upload(video):
        """Handle video upload - extract and display first frame."""
        if not video or not os.path.exists(video):
            return None, f"Video file not found: {video}" if video else "Please upload a video file."
        
        first_frame = extract_first_frame(video)
        if first_frame is not None:
            return first_frame, "Video uploaded successfully. Ready to process."
        return None, "Failed to extract first frame from video."
    
    def normalize_endpoint(endpoint_url: str) -> str:
        """Normalize endpoint URL to base URL."""
        if not endpoint_url:
            return DEFAULT_ENDPOINT
        endpoint = endpoint_url.rstrip("/")
        if endpoint.endswith("/v1/chat/completions"):
            endpoint = endpoint[:-20]
        return endpoint.rstrip("/")
    
    def process_wrapper(video, model, endpoint, prompt, interval):
        """Wrapper to handle processing with UI updates."""
        if not video:
            yield None, "Please upload a video file first.", "Please upload a video file first."
            return
        
        stop_processing.clear()
        base_endpoint = normalize_endpoint(endpoint)
        results_accumulated = ""
        frame_img = None
        
        try:
            for frame_img, results in process_video_stream(video, model, base_endpoint, prompt, stop_processing, interval):
                if stop_processing.is_set():
                    yield frame_img, results, "Processing stopped by user."
                    return
                results_accumulated = results
                frame_count = results_accumulated.count("**Frame at")
                yield frame_img, results_accumulated, f"Processing... {frame_count} frame(s) processed."
            
            frame_count = results_accumulated.count("**Frame at")
            yield frame_img, results_accumulated, f"Processing completed! Total: {frame_count} frame(s)."
        except Exception as e:
            yield frame_img, results_accumulated, f"Error: {str(e)}"
    
    def stop_wrapper():
        """Stop processing and update UI."""
        stop_processing.set()
        return None, "Processing stopped by user.", "Processing stopped."
    
    btn_start.click(
        fn=process_wrapper,
        inputs=[video_input, model_name, endpoint_url, prompt_input, frame_interval],
        outputs=[current_frame, results_text, status]
    )
    
    btn_stop.click(
        fn=stop_wrapper,
        inputs=[],
        outputs=[current_frame, results_text, status]
    )
    
    # Handle video upload - automatically display first frame
    video_input.change(
        fn=on_video_upload,
        inputs=[video_input],
        outputs=[current_frame, status]
    )

if __name__ == "__main__":
    demo.launch(server_name="0.0.0.0", server_port=7860, css=CUSTOM_CSS)

