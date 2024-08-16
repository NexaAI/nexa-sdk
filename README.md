# Nexa SDK

The Nexa SDK is a comprehensive toolkit for supporting **ONNX** and **GGML** models. It supports text generation, image generation, vision-language models (VLM), and text-to-speech (TTS) capabilities. Additionally, it offers an OpenAI-compatible API server with JSON schema mode for function calling and streaming support, and a user-friendly Streamlit UI.

## Features

- **Model Support:**
  - **ONNX & GGML models**
  - **Conversion Engine**
  - **Inference Engine**:
    - **Text Generation**
    - **Image Generation**
    - **Vision-Language Models (VLM)**
    - **Text-to-Speech (TTS)**
- **Server:**
  - OpenAI-compatible API
  - JSON schema mode for function calling
  - Streaming support
- **Streamlit UI** for interactive model deployment and testing

## Installation

For CPU version

```
curl -fsSL https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexa-sdk/scripts/build_nexaai.sh | sh
```

For GPU version

```
curl -fsSL https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexa-sdk/scripts/build_nexaai_gpu.sh | sh
```

## Nexa CLI commands

## Model Commands

### NLP Models

| Model            | Type | Format    | Command                              |
| ---------------- | ---- | --------- | ------------------------------------ |
| octopus-v2       | NLP  | GGUF      | `nexa-cli gen-text octopus-v2`       |
| octopus-v4       | NLP  | GGUF      | `nexa-cli gen-text octopus-v4`       |
| tinyllama        | NLP  | GGUF      | `nexa-cli gen-text tinyllama`        |
| llama2           | NLP  | GGUF/ONNX | `nexa-cli gen-text llama2`           |
| llama3           | NLP  | GGUF/ONNX | `nexa-cli gen-text llama3`           |
| llama3.1         | NLP  | GGUF/ONNX | `nexa-cli gen-text llama3.1`         |
| gemma            | NLP  | GGUF/ONNX | `nexa-cli gen-text gemma`            |
| gemma2           | NLP  | GGUF      | `nexa-cli gen-text gemma2`           |
| qwen1.5          | NLP  | GGUF      | `nexa-cli gen-text qwen1.5`          |
| qwen2            | NLP  | GGUF/ONNX | `nexa-cli gen-text qwen2`            |
| mistral          | NLP  | GGUF/ONNX | `nexa-cli gen-text mistral`          |
| codegemma        | NLP  | GGUF      | `nexa-cli gen-text codegemma`        |
| codellama        | NLP  | GGUF      | `nexa-cli gen-text codellama`        |
| codeqwen         | NLP  | GGUF      | `nexa-cli gen-text codeqwen`         |
| deepseek-coder   | NLP  | GGUF      | `nexa-cli gen-text deepseek-coder`   |
| dolphin-mistral  | NLP  | GGUF      | `nexa-cli gen-text dolphin-mistral`  |
| nomic-embed-text | NLP  | GGUF      | `nexa-cli gen-text nomic-embed-text` |
| phi2             | NLP  | GGUF      | `nexa-cli gen-text phi2`             |
| phi3             | NLP  | GGUF/ONNX | `nexa-cli gen-text phi3`             |

### Multimodal Models

| Model            | Type       | Format | Command                         |
| ---------------- | ---------- | ------ | ------------------------------- |
| nanollava        | Multimodal | GGUF   | `nexa-cli vlm nanollava`        |
| llava-phi3       | Multimodal | GGUF   | `nexa-cli vlm llava-phi3`       |
| llava-llama3     | Multimodal | GGUF   | `nexa-cli vlm llava-llama3`     |
| llava1.6-mistral | Multimodal | GGUF   | `nexa-cli vlm llava1.6-mistral` |
| llava1.6-vicuna  | Multimodal | GGUF   | `nexa-cli vlm llava1.6-vicuna`  |

### Computer Vision Models

| Model                 | Type            | Format    | Command                              |
| --------------------- | --------------- | --------- | ------------------------------------ |
| stable-diffusion-v1-4 | Computer Vision | GGUF      | `nexa-cli gen-image sd1-4`           |
| stable-diffusion-v1-5 | Computer Vision | GGUF/ONNX | `nexa-cli gen-image sd1-5`           |
| lcm-dreamshaper       | Computer Vision | GGUF/ONNX | `nexa-cli gen-image lcm-dreamshaper` |
