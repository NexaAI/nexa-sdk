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

Detailed API documentation is available [here](docs/index.html).

- **Server:**
  - OpenAI-compatible API
  - JSON schema mode for function calling
  - Streaming support
- **Streamlit UI** for interactive model deployment and testing

## Installation

**GPU version(optional)** 

check if you have GPU acceleration (torch required)
<details>
  <summary>CUDA:</summary>

  ```
  import torch
  torch.cuda.is_available()
  ```

  if True

  ```
  CMAKE_ARGS="-DGGML_CUDA=on -DSD_CUBLAS=ON" pip install nexaai-gpu
  ```
</details>
<details>
  <summary>Apple M Chip:</summary>
  Apple icon -> about this mac -> Graphics
  
  if True:

  ```
  CMAKE_ARGS="-DGGML_METAL=on -DSD_METAL=ON" pip install nexaai-gpu
  ```
</details>

<details>
  <summary>AMD graphics card:</summary>


  ```
  CMAKE_ARGS="-DGGML_HIPBLAS=on" pip install nexaai-gpu
  ```
</details>

**CPU version**

<details>
  <summary>Mac with Intel chips</summary>

  ```
  CMAKE_ARGS="-DCMAKE_CXX_FLAGS=-fopenmp" pip install nexaai
  ```
</details>

<details>
  <summary>Mac with M chips or other Operating systems:</summary>

  ```
  pip install nexaai
  ```
<details>

## Nexa CLI commands

## Model Commands

### NLP Models

| Model            | Type | Format    | Command                              |
| ---------------- | ---- | --------- | ------------------------------------ |
| octopus-v2       | NLP  | GGUF      | `nexa gen-text octopus-v2`       |
| octopus-v4       | NLP  | GGUF      | `nexa gen-text octopus-v4`       |
| tinyllama        | NLP  | GGUF      | `nexa gen-text tinyllama`        |
| llama2           | NLP  | GGUF/ONNX | `nexa gen-text llama2`           |
| llama3           | NLP  | GGUF/ONNX | `nexa gen-text llama3`           |
| llama3.1         | NLP  | GGUF/ONNX | `nexa gen-text llama3.1`         |
| gemma            | NLP  | GGUF/ONNX | `nexa gen-text gemma`            |
| gemma2           | NLP  | GGUF      | `nexa gen-text gemma2`           |
| qwen1.5          | NLP  | GGUF      | `nexa gen-text qwen1.5`          |
| qwen2            | NLP  | GGUF/ONNX | `nexa gen-text qwen2`            |
| mistral          | NLP  | GGUF/ONNX | `nexa gen-text mistral`          |
| codegemma        | NLP  | GGUF      | `nexa gen-text codegemma`        |
| codellama        | NLP  | GGUF      | `nexa gen-text codellama`        |
| codeqwen         | NLP  | GGUF      | `nexa gen-text codeqwen`         |
| deepseek-coder   | NLP  | GGUF      | `nexa gen-text deepseek-coder`   |
| dolphin-mistral  | NLP  | GGUF      | `nexa gen-text dolphin-mistral`  |
| nomic-embed-text | NLP  | GGUF      | `nexa gen-text nomic-embed-text` |
| phi2             | NLP  | GGUF      | `nexa gen-text phi2`             |
| phi3             | NLP  | GGUF/ONNX | `nexa gen-text phi3`             |

### Multimodal Models

| Model            | Type       | Format | Command                         |
| ---------------- | ---------- | ------ | ------------------------------- |
| nanollava        | Multimodal | GGUF   | `nexa vlm nanollava`        |
| llava-phi3       | Multimodal | GGUF   | `nexa vlm llava-phi3`       |
| llava-llama3     | Multimodal | GGUF   | `nexa vlm llava-llama3`     |
| llava1.6-mistral | Multimodal | GGUF   | `nexa vlm llava1.6-mistral` |
| llava1.6-vicuna  | Multimodal | GGUF   | `nexa vlm llava1.6-vicuna`  |

### Computer Vision Models

| Model                 | Type            | Format    | Command                              |
| --------------------- | --------------- | --------- | ------------------------------------ |
| stable-diffusion-v1-4 | Computer Vision | GGUF      | `nexa gen-image sd1-4`           |
| stable-diffusion-v1-5 | Computer Vision | GGUF/ONNX | `nexa gen-image sd1-5`           |
| lcm-dreamshaper       | Computer Vision | GGUF/ONNX | `nexa gen-image lcm-dreamshaper` |