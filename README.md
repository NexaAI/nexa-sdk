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

### GPU version(optional)

check if you have GPU acceleration (torch required)
<details>
  <summary>CUDA:</summary>

  ```
  import torch
  torch.cuda.is_available()
  ```

  if True

  ```
  CMAKE_ARGS="-DGGML_CUDA=on -DSD_CUBLAS=ON" pip install nexaai
  ```
  Or you prefer to install our pre-built wheel:
  ```bash
  pip install nexaai --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple
  ```
  Optionally, you can install onnx supported version:
  ```bash
  pip install nexaai[onnx] --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple
  ```
</details>
<details>
  <summary>Apple M Chip:</summary>
  Apple icon -> about this mac -> Graphics

  if True:

  ```
  CMAKE_ARGS="-DGGML_METAL=on -DSD_METAL=ON" pip install nexaai
  ```
  Or you prefer to install our pre-built wheel:
  ```bash
  pip install nexaai --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple
  ```
  Optionally, you can install onnx supported version:
  ```bash
  pip install nexaai[onnx] --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple
  ```
</details>

### CPU version

<details>
  <summary>Mac with Intel Chips</summary>

  To install the `nexaai` package on a Mac with Intel chips, use the following command:

  ```bash
  CMAKE_ARGS="-DCMAKE_CXX_FLAGS=-fopenmp" pip install nexaai
  ```

  **Optional:** To install the version with ONNX support, use:

  ```bash
  CMAKE_ARGS="-DCMAKE_CXX_FLAGS=-fopenmp" pip install nexaai[onnx]
  ```

</details>

<details>
  <summary>Mac with M Chips or Other Operating Systems</summary>

  To install the `nexaai` package on a Mac with M chips or other operating systems, use the following command:

  ```bash
  pip install nexaai
  ```

  **Optional:** To install the version with ONNX support, use:

  ```bash
  pip install nexaai[onnx]
  ```


</details>
If you prefer to install the pre-built wheel for CPU versions:

```bash
pip install nexaai --index-url https://nexaai.github.io/nexa-sdk/whl/cpu --extra-index-url https://pypi.org/simple
```

To include ONNX support:

```bash
pip install nexaai[onnx] --index-url https://nexaai.github.io/nexa-sdk/whl/cpu --extra-index-url https://pypi.org/simple
```

### Docker Usage
Note: Docker doesn't support GPU acceleration
```bash
docker pull nexa4ai/sdk:latest
```
replace following placeholder with your path and command
```bash
docker run -v <your_model_dir>:/model -it nexa4ai/sdk:latest [nexa_command] [your_model_relative_path]
```

Example:
```bash
docker run -v /home/ubuntu/.cache/nexa/hub/official:/model -it nexa4ai/sdk:latest nexa gen-text /model/Phi-3-mini-128k-instruct/q4_0.gguf
```

will create an interactive session with text generation


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