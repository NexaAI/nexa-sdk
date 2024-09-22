# NexaAI SDK Demo: Voice Transcription

## Introduction

This demo application showcases the capabilities of the NexaAI SDK for real-time voice transcription. The application is built using Streamlit and leverages the NexaAI SDK to transcribe audio in real-time, providing features like language translation, text summarization, and on-device processing for enhanced data privacy.

### Key Features

- **Real-Time Voice Transcription**: Transcribe audio in real-time using the NexaAI SDK.
- **Language Translation**: Supports translation of transcribed audio into different languages.
- **Text Summarization**: Generate summaries of the transcribed text using the NexaAI Text Inference model.
- **On-device Processing**: Ensures data privacy by processing audio locally on the device.
- **Interactive User Interface**: A user-friendly interface for managing transcription, summarization, and file uploads.

### File Structure

- **`app.py`**: The main Streamlit application that handles the user interface and controls the transcription and summarization processes.
- **`utils/segmenter.py`**: A utility module for segmenting audio streams using WebRTC VAD (Voice Activity Detection).
- **`utils/transcriber.py`**: Contains classes for handling real-time transcription and text inference.

## Setup

### 1. Install Required Packages

Install the required packages by running:

```bash
pip install -r requirements.txt

```

### 2. Usage

#### Running the Application

To start the Streamlit application, use the following command:

```bash
streamlit run app.py
```

#### Features

- **Real-Time Transcription**: Use the "Start Recording" button to begin transcribing audio in real-time.
- **Stop Recording**: Click "Stop" to end the recording session.
- **Upload Audio Files**: Upload `.wav` files for transcription using the file uploader.
- **Generate Summary**: After transcription, generate a summary of the text by clicking the "Generate Summary" button.
- **Download Transcription**: Download the transcribed text as a `.txt` file.

### 3. File Processing

- **Transcription**: The application can transcribe both live audio and pre-recorded `.wav` files.
- **Summarization**: Generate a concise summary of the transcribed text.
- **Translation**: If enabled, the application can translate the transcribed text into the specified language.


## Code Overview

### `app.py`

The main application script, handling the Streamlit interface, model configuration, and the transcription process. Users can start or stop recording, upload audio files, generate summaries, and download transcriptions directly from the app interface.

### `utils/segmenter.py`

A utility for segmenting audio streams using WebRTC's Voice Activity Detection (VAD). This module manages the audio stream, detecting when speech occurs and splitting the audio into manageable chunks for transcription.

### `utils/transcriber.py`

Handles the transcription process, including initializing the NexaAI Voice Inference model, processing audio chunks, and managing the transcription lifecycle. It also includes a `TextInference` class for generating summaries from the transcribed text using the NexaAI Text Inference model.

---
