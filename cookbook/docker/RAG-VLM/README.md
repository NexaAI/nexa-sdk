# AutoNeural Video Inference Demo

This demo showcases video analysis using AutoNeural model through Nexa SDK. It extracts frames from uploaded videos at 8-second intervals and performs real-time AI inference using the AutoNeural vision-language model.

## Features

- **Video Upload**: Upload video files for analysis
- **Frame Extraction**: Automatically extracts frames at 8-second intervals
- **Real-time Inference**: Processes frames sequentially and displays results in real-time
- **Interactive UI**: Left panel shows current frame, right panel shows accumulated inference results
- **Docker Support**: Complete Docker setup with nexa serve and Gradio UI

## Architecture

```
┌─────────────────┐
│  Gradio UI      │  (Port 7860)
│  - Video upload │
│  - Frame extract│
│  - Real-time    │
│    display      │
└────────┬────────┘
         │ HTTP API
         ▼
┌─────────────────┐
│  Nexa Serve     │  (Port 18181)
│  AutoNeural     │
│  /v1/chat/      │
│  completions    │
└─────────────────┘
```

## Prerequisites

- Docker installed
- AutoNeural model downloaded (will be pulled automatically by nexa serve if not cached)

## Quick Start

### Build Docker Image

```bash
cd cookbook/PC/Q
docker build -t autoneural-video-demo .
```

### Run Docker Container

```bash
docker run -d \
  --name autoneural-demo \
  -p 18181:18181 \
  -p 7860:7860 \
  -v /path/to/model/cache:/root/.cache/nexa.ai \
  autoneural-video-demo
```

**Note**: Replace `/path/to/model/cache` with your local model cache directory, or omit the `-v` flag to use container's internal cache.

### Access the UI

Open your browser and navigate to:
```
http://localhost:7860
```

## Usage

1. **Upload Video**: Click on the video upload area and select a video file
2. **Configure Settings** (optional):
   - Model name: Default is `NexaAI/AutoNeural`
   - Endpoint: Default is `http://127.0.0.1:18181`
   - Prompt: Customize the analysis prompt (default: "Describe what you see in this image in detail.")
3. **Start Processing**: Click "Start Processing" button
4. **View Results**: 
   - Left panel shows the current frame being processed
   - Right panel shows accumulated inference results for all processed frames
5. **Stop Processing**: Click "Stop" button to interrupt processing

## Local Development (Without Docker)

### Prerequisites

- Python 3.11+
- Nexa SDK installed and `nexa` command available
- Python dependencies installed

### Setup

1. Install dependencies:
```bash
pip install -r requirements.txt
```

2. Start nexa serve in a separate terminal:
```bash
nexa serve --host 127.0.0.1:18181
```

3. Run Gradio UI:
```bash
python gradio_ui.py
```

4. Access the UI at `http://localhost:7860`

## Configuration

### Frame Extraction

- **Frame Interval**: 8 seconds (configurable in `gradio_ui.py` via `FRAME_INTERVAL_SECONDS`)
- **Clip Length**: 8 seconds (configurable via `CLIP_LENGTH_SECONDS`)

### Model Settings

- **Default Model**: `NexaAI/AutoNeural`
- **Default Endpoint**: `http://127.0.0.1:18181`

These can be changed in the UI's "Model Settings" accordion.

## File Structure

```
cookbook/PC/Q/
├── gradio_ui.py          # Gradio UI main file
├── requirements.txt      # Python dependencies
├── Dockerfile            # Docker image configuration
├── start.sh              # Startup script
└── README.md             # This file
```

## Troubleshooting

### nexa serve not starting

- Check if port 18181 is already in use
- Verify Docker container has proper permissions
- Check container logs: `docker logs autoneural-demo`

### Model not found

- Ensure AutoNeural model is downloaded: `nexa pull NexaAI/AutoNeural`
- Check model cache volume mount in Docker run command
- Verify model cache directory permissions

### Video processing errors

- Ensure video file format is supported (MP4, AVI, MOV, etc.)
- Check video file is not corrupted
- Verify sufficient disk space for temporary frame files

### API connection errors

- Verify nexa serve is running and accessible
- Check endpoint URL in UI settings
- Ensure firewall allows connections on port 18181

## License

Copyright 2024-2025 Nexa AI, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

