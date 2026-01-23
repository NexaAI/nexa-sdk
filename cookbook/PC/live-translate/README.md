# NexaAI Live Translator

Real-time speech recognition and language translation demo using NexaAI's ASR and LLM models. Supports multi-language translation with low-latency sentence-level translation.

## Features

- 🎙️ **Real-time Speech Recognition**: Uses NexaAI's Parakeet ASR model for multi-language transcription
- 🌐 **Instant Translation**: LLM provides fast, accurate translations
- 📊 **Segment-based**: Automatic sentence/paragraph boundary detection for natural translations
- 🖥️ **Modern Web UI**: Clean, responsive interface with live text display

## Requirements
- Windows/Linux with Qualcomm NPU device.
- Python 3.9+

## Installation

### 1. Clone and Setup Environment

```bash
cd cookbook/PC/live-translate
python -m venv .venv

# Windows
.venv\Scripts\activate
# Linux
source .venv/bin/activate
```

### 2. Install Dependencies

```bash
pip install -r requirements.txt
```

### 3. Download Models

Download the required models using NexaAI CLI:

```bash
nexa pull NexaAI/parakeet-tdt-0.6b-v3-npu
nexa pull NexaAI/HY-MT1.5-1.8B-npu
```

## Usage

### Quick Start

```bash
python app.py
```

Then open your browser to: **http://localhost:5000**


### How to Use

1. **Select Target Language**: Choose your desired translation language from the dropdown (e.g., Chinese, English, French, etc.)
2. **Start Recording**: Click the "Start Recording" button
3. **Speak**: Speak naturally into your microphone
4. **Watch Real-time Translation**: 
    - Left panel shows real-time transcription
    - Right panel shows real-time translation (auto-updated per sentence)
5. **Stop Recording**: Click "Stop Recording" to end
