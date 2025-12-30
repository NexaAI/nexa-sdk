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
import io
import json
import logging
import sys
import threading
from typing import Optional, Generator, Tuple

import gradio as gr
import requests
import numpy as np
from PIL import Image
import imageio

# Configuration
DEFAULT_MODEL = "NexaAI/AutoNeural"
DEFAULT_ENDPOINT = "http://127.0.0.1:18181"
DEFAULT_FRAME_INTERVAL = 8  # Extract frame every 8 seconds


# Configure logging
def setup_logging(log_level: str = "INFO", log_file: Optional[str] = None):
    """
    Setup logging configuration.

    Args:
        log_level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
        log_file: Optional log file path. If None, logs only to console.
    """
    level = getattr(logging, log_level.upper(), logging.INFO)

    # Create formatter
    formatter = logging.Formatter(
        fmt="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )

    # Configure root logger
    logger = logging.getLogger()
    logger.setLevel(level)

    # Remove existing handlers
    logger.handlers.clear()

    # Console handler
    console_handler = logging.StreamHandler(sys.stdout)
    console_handler.setLevel(level)
    console_handler.setFormatter(formatter)
    logger.addHandler(console_handler)

    # File handler (if specified)
    if log_file:
        file_handler = logging.FileHandler(log_file, encoding="utf-8")
        file_handler.setLevel(level)
        file_handler.setFormatter(formatter)
        logger.addHandler(file_handler)

    return logger


# Initialize logger
logger = logging.getLogger(__name__)

# Setup logging on module import
setup_logging(log_level="DEBUG")


def image_array_to_base64(image_array: np.ndarray) -> str:
    """Convert numpy image array to base64 data URI (in-memory, no disk I/O)."""
    # Convert to PIL Image
    if image_array.dtype != np.uint8:
        image_array = (
            (image_array * 255).astype(np.uint8)
            if image_array.max() <= 1.0
            else image_array.astype(np.uint8)
        )

    pil_image = Image.fromarray(image_array)

    # Save to bytes buffer
    buffer = io.BytesIO()
    pil_image.save(buffer, format="JPEG", quality=95)
    image_bytes = buffer.getvalue()

    # Encode to base64
    base64_data = base64.b64encode(image_bytes).decode("utf-8")
    return f"data:image/jpeg;base64,{base64_data}"


def call_nexa_chat_stream(
    session: requests.Session, model: str, messages: list, endpoint: str
) -> Generator[str, None, None]:
    """
    Call Nexa /v1/chat/completions endpoint (streaming).

    Args:
        session: Requests session for connection reuse
        model: Model name
        messages: List of chat messages
        endpoint: Base URL for API endpoint

    Yields:
        Text chunks as they arrive
    """
    url = f"{endpoint.rstrip('/')}/v1/chat/completions"
    payload = {
        "model": model,
        "messages": messages,
        "stream": True,
        "max_tokens": 512,
        "enable_think": False,
    }

    try:
        logger.debug(f"Calling streaming API: {url} with model: {model}")
        with session.post(url, json=payload, stream=True, timeout=300) as resp:
            resp.raise_for_status()

            for line in resp.iter_lines(decode_unicode=True):
                if not line:
                    continue

                if line.startswith("data:"):
                    chunk = line[len("data:") :].strip()
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
                        logger.warning(f"Error parsing stream chunk: {e}")
                        continue

            logger.debug("Streaming API call completed")
    except requests.exceptions.RequestException as e:
        logger.error(f"Streaming API request failed: {e}")
        yield f"Error: {str(e)}"
    except Exception as e:
        logger.exception(f"Unexpected error in streaming API call: {e}")
        yield f"Error: {str(e)}"


def extract_first_frame(video_path: str) -> Optional[np.ndarray]:
    """Extract the first frame from video."""
    logger.info(f"Extracting first frame from: {video_path}")
    try:
        with imageio.get_reader(video_path) as reader:
            frame = reader.get_data(0)  # type: ignore
            result = frame if isinstance(frame, np.ndarray) else np.array(frame)
            logger.debug(f"Successfully extracted first frame, shape: {result.shape}")
            return result
    except Exception as e:
        logger.error(
            f"Error extracting first frame from {video_path}: {e}", exc_info=True
        )
        return None


def extract_frames_from_video(
    video_path: str, interval_seconds: float = DEFAULT_FRAME_INTERVAL
) -> list:
    """Extract frames from video at fixed intervals (optimized: direct frame access)."""
    logger.info(
        f"Extracting frames from video: {video_path}, interval: {interval_seconds}s"
    )
    try:
        reader = imageio.get_reader(video_path)
        metadata = reader.get_meta_data()  # type: ignore
        fps = metadata.get("fps", 30.0)
        logger.debug(f"Video metadata - FPS: {fps}")

        if fps <= 0:
            logger.warning(f"Invalid FPS: {fps}, cannot extract frames")
            reader.close()
            return []

        # Calculate target frame indices directly (skip unnecessary frames)
        frame_interval = int(fps * interval_seconds)
        if frame_interval <= 0:
            frame_interval = 1

        # Get total frame count - required for direct frame access
        total_frames = reader.count_frames()  # type: ignore
        logger.debug(f"Total frames: {total_frames}, frame interval: {frame_interval}")

        if total_frames <= 0:
            logger.warning(f"Invalid total frames: {total_frames}")
            reader.close()
            return []

        # Directly access target frames - skips all intermediate frames
        frames = []
        frame_indices = list(range(0, total_frames, frame_interval))
        logger.info(f"Extracting {len(frame_indices)} frames at intervals")

        for frame_idx in frame_indices:
            try:
                frame = reader.get_data(frame_idx)  # type: ignore
                timestamp = frame_idx / fps
                frame_array = (
                    frame if isinstance(frame, np.ndarray) else np.array(frame)
                )
                frames.append((frame_array, timestamp))
            except (IndexError, Exception) as e:
                logger.warning(f"Reached end of video at frame {frame_idx}: {e}")
                break

        reader.close()
        logger.info(f"Successfully extracted {len(frames)} frames")
        return frames
    except Exception as e:
        logger.error(f"Error reading video {video_path}: {e}", exc_info=True)
        return []


def process_video_stream(
    video_file: Optional[str],
    model: str,
    endpoint: str,
    prompt: str,
    stop_event: threading.Event,
    interval_seconds: float = DEFAULT_FRAME_INTERVAL,
) -> Generator[Tuple[Optional[np.ndarray], str], None, None]:
    """Process video frames sequentially and call API for inference."""
    logger.info(
        f"Starting video processing: {video_file}, model: {model}, endpoint: {endpoint}"
    )

    if not video_file or not os.path.exists(video_file):
        error_msg = (
            f"Video file not found: {video_file}"
            if video_file
            else "Please upload a video file."
        )
        logger.error(error_msg)
        yield None, error_msg
        return

    frames = extract_frames_from_video(video_file, interval_seconds)
    if not frames:
        error_msg = "Failed to extract frames from video. Please check the video file."
        logger.error(error_msg)
        yield None, error_msg
        return

    logger.info(f"Processing {len(frames)} frames")

    # Create session for connection reuse
    session = requests.Session()
    accumulated_results = []

    try:
        for idx, (frame_image, timestamp) in enumerate(frames, 1):
            if stop_event.is_set():
                logger.info("Processing stopped by user")
                yield frame_image, "\n\n".join(accumulated_results)
                return

            logger.debug(f"Processing frame {idx}/{len(frames)} at {timestamp:.1f}s")

            # Convert frame to base64 in memory (no disk I/O)
            image_data_uri = image_array_to_base64(frame_image)

            # Build messages
            messages = [
                {
                    "role": "user",
                    "content": [
                        {"type": "text", "text": prompt},
                        {"type": "image_url", "image_url": {"url": image_data_uri}},
                    ],
                }
            ]

            # Call streaming API and accumulate text in real-time
            result_text = ""
            result_entry_prefix = f"**Frame at {timestamp:.1f}s:**\n"

            # Create a temporary entry that will be updated as stream progresses
            current_entry = result_entry_prefix
            accumulated_results.append(current_entry)

            # Stream response and update UI in real-time
            for chunk in call_nexa_chat_stream(session, model, messages, endpoint):
                if stop_event.is_set():
                    break

                result_text += chunk
                # Update the last entry with accumulated text
                accumulated_results[-1] = result_entry_prefix + result_text + "\n"

                # Yield updated results for real-time display
                yield frame_image, "\n\n".join(accumulated_results)

            # Final update with complete result
            accumulated_results[-1] = result_entry_prefix + result_text + "\n"
            logger.debug(
                f"Frame {idx}/{len(frames)} processed successfully, length: {len(result_text)}"
            )
            yield frame_image, "\n\n".join(accumulated_results)

        logger.info(f"Video processing completed: {len(frames)} frames processed")

    finally:
        session.close()
        logger.debug("HTTP session closed")


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
    gr.Markdown("## AutoNeural Video Inference")

    with gr.Row():
        # Left column: settings / video upload
        with gr.Column(scale=1, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Video & Settings**")

            video_input = gr.Video(label="Upload Video")

            with gr.Accordion("Model Settings", open=False):
                model_name = gr.Textbox(
                    label="Model Name",
                    value=DEFAULT_MODEL,
                    info="AutoNeural model name",
                )
                endpoint_url = gr.Textbox(
                    label="Endpoint URL",
                    value=f"{DEFAULT_ENDPOINT}/v1/chat/completions",
                    info="Nexa serve endpoint URL (full path)",
                )
                frame_interval = gr.Slider(
                    1,
                    30,
                    value=DEFAULT_FRAME_INTERVAL,
                    step=1,
                    label="Frame Interval (seconds)",
                    info="Extract frame every N seconds",
                )
                prompt_input = gr.Textbox(
                    label="Prompt",
                    value="Describe what you see in this image in few sentences.",
                    placeholder="Enter your prompt for image analysis...",
                    lines=3,
                )

            with gr.Row():
                btn_start = gr.Button("Start Processing", variant="primary")
                btn_stop = gr.Button("Stop Processing", variant="stop")

            status = gr.Textbox(label="Status", value="", interactive=False)

        # Middle column: video display
        with gr.Column(scale=2, elem_classes=["rounded-card", "pad"]):
            gr.Markdown("**Current Frame**")
            current_frame = gr.Image(
                label="Current Frame", type="numpy", height=600, show_label=False
            )

        # Right column: information panel (results)
        with gr.Column(scale=2):
            with gr.Accordion("Inference Results", open=True):
                results_text = gr.Markdown(
                    value="Results will appear here as frames are processed...",
                    label="Results",
                )

    # Event handlers
    def on_video_upload(video):
        """Handle video upload - extract and display first frame."""
        logger.info(f"Video upload event: {video}")
        if not video or not os.path.exists(video):
            error_msg = (
                f"Video file not found: {video}"
                if video
                else "Please upload a video file."
            )
            logger.warning(error_msg)
            return None, error_msg

        first_frame = extract_first_frame(video)
        if first_frame is not None:
            logger.info("Video uploaded successfully")
            return first_frame, "Video uploaded successfully. Ready to process."
        logger.error("Failed to extract first frame from video")
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
        logger.info(
            f"Processing started - video: {video}, model: {model}, interval: {interval}s"
        )
        if not video:
            logger.warning("No video file provided")
            yield (
                None,
                "Please upload a video file first.",
                "Please upload a video file first.",
            )
            return

        stop_processing.clear()
        base_endpoint = normalize_endpoint(endpoint)
        logger.debug(f"Normalized endpoint: {base_endpoint}")
        results_accumulated = ""
        frame_img = None

        try:
            for frame_img, results in process_video_stream(
                video, model, base_endpoint, prompt, stop_processing, interval
            ):
                if stop_processing.is_set():
                    logger.info("Processing stopped by user in wrapper")
                    yield frame_img, results, "Processing stopped by user."
                    return
                results_accumulated = results
                frame_count = results_accumulated.count("**Frame at")
                yield (
                    frame_img,
                    results_accumulated,
                    f"Processing... {frame_count} frame(s) processed.",
                )

            frame_count = results_accumulated.count("**Frame at")
            logger.info(f"Processing completed successfully: {frame_count} frames")
            yield (
                frame_img,
                results_accumulated,
                f"Processing completed! Total: {frame_count} frame(s).",
            )
        except Exception as e:
            logger.exception(f"Error in process_wrapper: {e}")
            yield frame_img, results_accumulated, f"Error: {str(e)}"

    def stop_wrapper():
        """Stop processing and update UI."""
        logger.info("Stop processing requested")
        stop_processing.set()
        return None, "Processing stopped by user.", "Processing stopped."

    btn_start.click(
        fn=process_wrapper,
        inputs=[video_input, model_name, endpoint_url, prompt_input, frame_interval],
        outputs=[current_frame, results_text, status],
    )

    btn_stop.click(
        fn=stop_wrapper, inputs=[], outputs=[current_frame, results_text, status]
    )

    # Handle video upload - automatically display first frame
    video_input.change(
        fn=on_video_upload, inputs=[video_input], outputs=[current_frame, status]
    )

if __name__ == "__main__":
    logger.info("Starting Gradio application")
    logger.info("Server: 0.0.0.0:7860")
    logger.info(f"Default model: {DEFAULT_MODEL}")
    logger.info(f"Default endpoint: {DEFAULT_ENDPOINT}")
    demo.launch(server_name="0.0.0.0", server_port=7860)
