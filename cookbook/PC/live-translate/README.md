# NexaAI Live Translator

Real-time speech recognition and language translation demo using NexaAI's ASR and LLM models. Supports English ↔ Chinese with low-latency sentence-level translation.

## Features

- 🎙️ **Real-time Speech Recognition**: Uses NexaAI's Parakeet ASR model for English and Chinese transcription
- 🌐 **Instant Translation**: Llama 3.2 3B LLM provides fast, accurate translations
- 📊 **Segment-based**: Automatic sentence/paragraph boundary detection for natural translations
- 🖥️ **Modern Web UI**: Clean, responsive interface with live text display

## Architecture

```
┌─────────────────┐
│   Frontend      │
│  (HTML/JS UI)   │
└────────┬────────┘
         │ WebSocket
         │ Audio Stream
         │
┌────────▼────────────────────────────────────┐
│     Flask Backend + ASR/LLM Pipeline        │
├─────────────────────────────────────────────┤
│  ASR.stream() → on_transcription callback   │
│       ↓                                      │
│  LLM.generate() → Translation                │
│       ↓                                      │
│  WebSocket emit → Return to Frontend        │
└─────────────────────────────────────────────┘
```

## Requirements

- Python 3.9+
- NexaAI SDK
- Microphone access for audio input

## Installation

### 1. Clone and Setup Environment

```bash
cd cookbook/PC/live-translate
python -m venv .venv

# Windows
.venv\Scripts\activate
# Linux/macOS
source .venv/bin/activate
```

### 2. Install Dependencies

```bash
pip install -r requirements.txt
```

### 3. Download Models

Download the required models using NexaAI CLI:

```bash
nexa pull NexaAI/parakeet-npu
nexa pull Qwen/Qwen3-1.7B-GGUF
```

## Usage

### Quick Start

```bash
python app.py
```

Then open your browser to: **http://localhost:5000**

### How to Use

1. **Select Language**: Choose source language (English or Chinese) from the dropdown
2. **Start Recording**: Click "Start Recording" button
3. **Speak**: Speak naturally into your microphone
4. **Watch Real-time Translation**: 
   - Left panel shows transcribed text
   - Right panel shows translated text (auto-updated per sentence)
5. **Stop Recording**: Click "Stop Recording" to end session


### ASR Stream Parameters

Edit `streaming.py` to adjust ASR behavior:

```python
config = ASRStreamConfig(
    sample_rate=16000,        # Audio sample rate (Hz)
    chunk_duration=3.0,       # Process audio in 3-second chunks
    overlap_duration=1.0,     # 1-second overlap for context
    max_queue_size=10,        # Max pending chunks
    buffer_size=4096,         # Internal buffer size
    timestamps='segment',     # Trigger on sentence boundaries
    beam_size=5,              # Beam search width
)
```

### Translation Model

Models are configured in `app.py`:

```python
asr_model = ASR.from_(
    model='NexaAI/parakeet-npu',
    plugin_id=None,          # Auto-detect plugin
    device_id=None,          # Auto-detect device
)

llm_model = LLM.from_(
    model='NexaAI/Llama3.2-3B-NPU-Turbo',
)
```

To use different models, modify the `model` parameter. For CPU-only systems:

```python
ASR.from_(model='NexaAI/parakeet-cpu')
LLM.from_(model='NexaAI/Llama3.2-3B-CPU')
```
