<div align="center">

<h1>Nexa SDK</h1>

  <img src="assets/banner.png" alt="icon"/>

[![MacOS][MacOS-image]][release-url] [![Linux][Linux-image]][release-url] [![Windows][Windows-image]][release-url]

[![GitHub Release](https://img.shields.io/github/v/release/NexaAI/nexa-sdk)](https://github.com/NexaAI/nexa-sdk/releases/latest) [![Build workflow](https://img.shields.io/github/actions/workflow/status/NexaAI/nexa-sdk/ci.yaml?label=CI&logo=github)](https://github.com/NexaAI/nexa-sdk/actions/workflows/ci.yaml?query=branch%3Amain) ![GitHub License](https://img.shields.io/github/license/NexaAI/nexa-sdk)

<!-- ![PyPI - Python Version](https://img.shields.io/pypi/pyversions/nexaai) ![PyPI - Downloads](https://img.shields.io/pypi/dm/nexaai?color=orange) -->

[![Discord](https://dcbadge.limes.pink/api/server/thRu2HaK4D?style=flat&compact=true)](https://discord.gg/thRu2HaK4D)

[On-device Model Hub](https://model-hub.nexa4ai.com/) / [Nexa SDK Documentation](https://docs.nexaai.com/)

[release-url]: https://github.com/NexaAI/nexa-sdk/releases
[Windows-image]: https://img.shields.io/badge/windows-0078D4?logo=windows
[MacOS-image]: https://img.shields.io/badge/-MacOS-black?logo=apple
[Linux-image]: https://img.shields.io/badge/-Linux-333?logo=ubuntu

</div>

Nexa SDK is a comprehensive toolkit for supporting **ONNX** and **GGML** models. It supports text generation, image generation, vision-language models (VLM), and text-to-speech (TTS) capabilities. Additionally, it offers an OpenAI-compatible API server with JSON schema mode for function calling and streaming support, and a user-friendly Streamlit UI.

## Features

- **Model Support:**

  - **ONNX & GGML models**
  - **Conversion Engine**
  - **Inference Engine**:
    - **Text Generation**
    - **Image Generation**
    - **Vision-Language Models (VLM)**
    - **Text-to-Speech (TTS)**

Detailed API documentation is available [here](https://docs.nexaai.com/).

- **Server:**
  - OpenAI-compatible API
  - JSON schema mode for function calling
  - Streaming support
- **Streamlit UI** for interactive model deployment and testing

## Installation

### Pre-built Wheels (Recommended)

We have released pre-built wheels for various Python versions, platforms, and backends for convenient installation on our [index page](https://nexaai.github.io/nexa-sdk/whl/).

#### CPU

```bash
pip install nexaai --index-url https://nexaai.github.io/nexa-sdk/whl/cpu --extra-index-url https://pypi.org/simple
```

#### GPU (Metal)

For the GPU version supporting **Metal (macOS)**:

```bash
pip install nexaai --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple
```

#### GPU (CUDA)

For the GPU version supporting **CUDA (Linux/Windows)**:

```bash
pip install nexaai --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple
```

> [!NOTE]
> The CUDA wheels are built with CUDA 12.4, but should be compatible with all CUDA 12.X

### Install from source code distribution

If pre-built wheels cannot meet your requirements, you can install Nexa SDK from the source code via cmake.

#### CPU

```bash
pip install nexaai
```

> [!IMPORTANT]
> If you are using a Mac with Intel chips, run the following command:
>
> ```bash
> CMAKE_ARGS="-DCMAKE_CXX_FLAGS=-fopenmp" pip install nexaai
> ```

#### GPU (Metal)

For the GPU version supporting Metal (macOS):

```bash
CMAKE_ARGS="-DGGML_METAL=ON -DSD_METAL=ON" pip install nexaai
```

#### GPU (CUDA)

For the GPU version supporting CUDA (Linux/Windows), run the following command:

```bash
CMAKE_ARGS="-DGGML_CUDA=ON -DSD_CUBLAS=ON" pip install nexaai
```

> [!TIP]
> You can accelerate the building process via parallel cmake by appending the following to the commands above:
>
> ```bash
> CMAKE_BUILD_PARALLEL_LEVEL=$(nproc)
> ```
>
> For example:
>
> ```bash
> CMAKE_BUILD_PARALLEL_LEVEL=$(nproc) CMAKE_ARGS="-DGGML_METAL=ON -DSD_METAL
> ```

> [!TIP]
> For Windows users, we recommend running the installation command in Git Bash to avoid unexpected behavior.

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

| Model            | Type | Format    | Command                          |
| ---------------- | ---- | --------- | -------------------------------- |
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

| Model            | Type       | Format | Command                     |
| ---------------- | ---------- | ------ | --------------------------- |
| nanollava        | Multimodal | GGUF   | `nexa vlm nanollava`        |
| llava-phi3       | Multimodal | GGUF   | `nexa vlm llava-phi3`       |
| llava-llama3     | Multimodal | GGUF   | `nexa vlm llava-llama3`     |
| llava1.6-mistral | Multimodal | GGUF   | `nexa vlm llava1.6-mistral` |
| llava1.6-vicuna  | Multimodal | GGUF   | `nexa vlm llava1.6-vicuna`  |

### Computer Vision Models

| Model                 | Type            | Format    | Command                          |
| --------------------- | --------------- | --------- | -------------------------------- |
| stable-diffusion-v1-4 | Computer Vision | GGUF      | `nexa gen-image sd1-4`           |
| stable-diffusion-v1-5 | Computer Vision | GGUF/ONNX | `nexa gen-image sd1-5`           |
| lcm-dreamshaper       | Computer Vision | GGUF/ONNX | `nexa gen-image lcm-dreamshaper` |