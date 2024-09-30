<div align="center">

<h1>Nexa SDK</h1>

  <img src="https://public-storage.nexa4ai.com/nexa-banner.png" alt="icon" onerror="this.onerror=null; this.src='./assets/banner.png'"/>

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

Nexa SDK is a comprehensive toolkit for supporting **ONNX** and **GGML** models. It supports text generation, image generation, vision-language models (VLM), and text-to-speech (TTS) capabilities. Additionally, it offers an OpenAI-compatible API server with JSON schema mode for function calling and streaming support, and a user-friendly Streamlit UI. Users can run Nexa SDK in any device with Python environment, and GPU acceleration is supported.

## Latest News 🔥
* [2024/09] Nexa now has executables for easy installation: [Install Nexa SDK](https://nexaai.com/download-sdk)
* [2024/09] Added support for Llama 3.2 models: `nexa run llama3.2`
* [2024/09] Added support for Qwen2.5, Qwen2.5-coder and Qwen2.5-Math models: `nexa run qwen2.5`
* [2024/09] Now supporting pulling and running GGUF models from Hugging Face: `nexa run -hf <hf model id>`
* [2024/09] Added support for ROCm
* [2024/09] Added support for Phi-3.5 models: `nexa run phi3.5`
* [2024/09] Added support for OpenELM models: `nexa run openelm`
* [2024/09] Introduced logits API support for more advanced model interactions
* [2024/09] Added support for Flux models: `nexa run flux`
* [2024/09] Added support for Stable Diffusion 3 model: `nexa run sd3`
* [2024/09] Added support for Stable Diffusion 2.1 model: `nexa run sd2-1`

Welcome to submit your requests through [issues](https://github.com/NexaAI/nexa-sdk/issues/new/choose), we ship weekly.

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

Below is our differentiation from other similar tools:

| **Feature**                | **[Nexa SDK](https://github.com/NexaAI/nexa-sdk)** | **[ollama](https://github.com/ollama/ollama)** | **[Optimum](https://github.com/huggingface/optimum)** | **[LM Studio](https://github.com/lmstudio-ai)** |
| -------------------------- | :------------------------------------------------: | :--------------------------------------------: | :---------------------------------------------------: | :---------------------------------------------: |
| **GGML Support**           |                         ✅                         |                       ✅                       |                          ❌                           |                       ✅                        |
| **ONNX Support**           |                         ✅                         |                       ❌                       |                          ✅                           |                       ❌                        |
| **Text Generation**        |                         ✅                         |                       ✅                       |                          ✅                           |                       ✅                        |
| **Image Generation**       |                         ✅                         |                       ❌                       |                          ❌                           |                       ❌                        |
| **Vision-Language Models** |                         ✅                         |                       ✅                       |                          ✅                           |                       ✅                        |
| **Text-to-Speech**         |                         ✅                         |                       ❌                       |                          ✅                           |                       ❌                        |
| **Server Capability**      |                         ✅                         |                       ✅                       |                          ✅                           |                       ✅                        |
| **User Interface**         |                         ✅                         |                       ❌                       |                          ❌                           |                       ✅                        |

## Installation

### macOS
[Download](https://public-storage.nexa4ai.com/nexa-sdk-executable-installer/nexa-macos-installer.pkg)

### Linux
```bash
curl -fsSL https://public-storage.nexa4ai.com/install.sh | sh 
```
### Windows
Coming soon. Install with Python package below 👇

## Python Package

We have released pre-built wheels for various Python versions, platforms, and backends for convenient installation on our [index page](https://nexaai.github.io/nexa-sdk/whl/).

> [!NOTE]
>
> 1. If you want to use <strong>ONNX model</strong>, just replace `pip install nexaai` with `pip install "nexaai[onnx]"` in provided commands.
> 2. For Chinese developers, we recommend you to use <strong>Tsinghua Open Source Mirror</strong> as extra index url, just replace `--extra-index-url https://pypi.org/simple` with `--extra-index-url https://pypi.tuna.tsinghua.edu.cn/simple` in provided commands.

#### CPU

```bash
pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cpu --extra-index-url https://pypi.org/simple --no-cache-dir
```

#### GPU (Metal)

For the GPU version supporting **Metal (macOS)**:

```bash
CMAKE_ARGS="-DGGML_METAL=ON -DSD_METAL=ON" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple --no-cache-dir
```

<details>
<summary><strong>FAQ: cannot use Metal/GPU on M1</strong></summary>

Try the following command:

```bash
wget https://github.com/conda-forge/miniforge/releases/latest/download/Miniforge3-MacOSX-arm64.sh
bash Miniforge3-MacOSX-arm64.sh
conda create -n nexasdk python=3.10
conda activate nexasdk
CMAKE_ARGS="-DGGML_METAL=ON -DSD_METAL=ON" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/metal --extra-index-url https://pypi.org/simple --no-cache-dir
```

</details>

#### GPU (CUDA)

For **Linux**:

```bash
CMAKE_ARGS="-DGGML_CUDA=ON -DSD_CUBLAS=ON" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple --no-cache-dir
```

For **Windows PowerShell**:

```bash
$env:CMAKE_ARGS="-DGGML_CUDA=ON -DSD_CUBLAS=ON"; pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple --no-cache-dir
```

For **Windows Command Prompt**:

```bash
set CMAKE_ARGS="-DGGML_CUDA=ON -DSD_CUBLAS=ON" & pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple --no-cache-dir
```

For **Windows Git Bash**:

```bash
CMAKE_ARGS="-DGGML_CUDA=ON -DSD_CUBLAS=ON" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/cu124 --extra-index-url https://pypi.org/simple --no-cache-dir
```

<details>
<summary><strong>FAQ: Building Issues for llava</strong></summary>

If you encounter the following issue while building:

![](docs/.media/error.jpeg)

try the following command:

```bash
CMAKE_ARGS="-DCMAKE_CXX_FLAGS=-fopenmp" pip install nexaai
```

</details>

#### GPU (ROCm)

For **Linux**:

```bash
CMAKE_ARGS="-DGGML_HIPBLAS=on" pip install nexaai --prefer-binary --index-url https://nexaai.github.io/nexa-sdk/whl/rocm621 --extra-index-url https://pypi.org/simple --no-cache-dir
```

### Local Build

How to clone this repo

```bash
git clone --recursive https://github.com/NexaAI/nexa-sdk
```

If you forget to use `--recursive`, you can use below command to add submodule

```bash
git submodule update --init --recursive
```

Then you can build and install the package

```bash
pip install -e .
```

## Supported Models

| Model                                                                                                   | Type            | Format    | Command                            |
| ------------------------------------------------------------------------------------------------------- | --------------- | --------- | ---------------------------------- |
| [octopus-v2](https://www.nexaai.com/NexaAI/Octopus-v2/gguf-q4_0/readme)                                 | NLP             | GGUF      | `nexa run octopus-v2`              |
| [octopus-v4](https://www.nexaai.com/NexaAI/Octopus-v4/gguf-q4_0/readme)                                 | NLP             | GGUF      | `nexa run octopus-v4`              |
| [gpt2](https://nexaai.com/openai/gpt2/gguf-q4_0/readme)                                                 | NLP             | GGUF      | `nexa run gpt2`                    |
| [tinyllama](https://www.nexaai.com/TinyLlama/TinyLlama-1.1B-Chat-v1.0/gguf-fp16/readme)                 | NLP             | GGUF      | `nexa run tinyllama`               |
| [llama2](https://www.nexaai.com/meta/Llama2-7b-chat/gguf-q4_0/readme)                                   | NLP             | GGUF/ONNX | `nexa run llama2`                  |
| [llama2-uncensored](https://www.nexaai.com/georgesung/Llama2-7b-chat-uncensored/gguf-q4_0/readme)       | NLP             | GGUF      | `nexa run llama2-uncensored`       |
| [llama2-function-calling](https://www.nexaai.com/Trelis/Llama2-7b-function-calling/gguf-q4_K_M/readme)  | NLP             | GGUF      | `nexa run llama2-function-calling` |
| [llama3](https://www.nexaai.com/meta/Llama3-8B-Instruct/gguf-q4_0/readme)                               | NLP             | GGUF/ONNX | `nexa run llama3`                  |
| [llama3.1](https://www.nexaai.com/meta/Llama3.1-8B-Instruct/gguf-q4_0/readme)                           | NLP             | GGUF/ONNX | `nexa run llama3.1`                |
| [llama3.2](https://nexaai.com/meta/Llama3.2-3B-Instruct/gguf-q4_0/readme)                               | NLP             | GGUF      | `nexa run llama3.2`                |
| [llama3-uncensored](https://www.nexaai.com/Orenguteng/Llama3-8B-Lexi-Uncensored/gguf-q4_K_M/readme)     | NLP             | GGUF      | `nexa run llama3-uncensored`       |
| [gemma](https://www.nexaai.com/google/gemma-1.1-2b-instruct/gguf-q4_0/readme)                           | NLP             | GGUF/ONNX | `nexa run gemma`                   |
| [gemma2](https://www.nexaai.com/google/gemma-2-2b-instruct/gguf-q4_0/readme)                            | NLP             | GGUF      | `nexa run gemma2`                  |
| [qwen1.5](https://www.nexaai.com/Qwen/Qwen1.5-7B-Instruct/gguf-q4_0/readme)                             | NLP             | GGUF      | `nexa run qwen1.5`                 |
| [qwen2](https://www.nexaai.com/Qwen/Qwen2-1.5B-Instruct/gguf-q4_0/readme)                               | NLP             | GGUF/ONNX | `nexa run qwen2`                   |
| [qwen2.5](https://www.nexaai.com/Qwen/Qwen2.5-1.5B-Instruct/gguf-q4_0/readme)                           | NLP             | GGUF      | `nexa run qwen2.5`                 |
| [mathqwen](https://nexaai.com/Qwen/Qwen2.5-Math-1.5B-Instruct/gguf-q4_0/readme)                         | NLP             | GGUF      | `nexa run mathqwen`                |
| [codeqwen](https://www.nexaai.com/Qwen/CodeQwen1.5-7B-Instruct/gguf-q4_0/readme)                        | NLP             | GGUF      | `nexa run codeqwen`                |
| [mistral](https://www.nexaai.com/mistralai/Mistral-7B-Instruct-v0.3/gguf-q4_0/readme)                   | NLP             | GGUF/ONNX | `nexa run mistral`                 |
| [dolphin-mistral](https://www.nexaai.com/CognitiveComputations/dolphin-2.8-mistral-7b/gguf-q4_0/readme) | NLP             | GGUF      | `nexa run dolphin-mistral`         |
| [codegemma](https://www.nexaai.com/google/codegemma-2b/gguf-q4_0/readme)                                | NLP             | GGUF      | `nexa run codegemma`               |
| [codellama](https://www.nexaai.com/meta/CodeLlama-7b-Instruct/gguf-q2_K/readme)                         | NLP             | GGUF      | `nexa run codellama`               |
| [deepseek-coder](https://www.nexaai.com/DeepSeek/deepseek-coder-1.3b-instruct/gguf-q4_0/readme)         | NLP             | GGUF      | `nexa run deepseek-coder`          |
| [phi2](https://www.nexaai.com/microsoft/Phi-2/gguf-q4_0/readme)                                         | NLP             | GGUF      | `nexa run phi2`                    |
| [phi3](https://www.nexaai.com/microsoft/Phi-3-mini-128k-instruct/gguf-q4_0/readme)                      | NLP             | GGUF/ONNX | `nexa run phi3`                    |
| [phi3.5](https://nexaai.com/microsoft/Phi-3.5-mini-instruct/gguf-q4_0/readme)                           | NLP             | GGUF      | `nexa run phi3.5`                  |
| [openelm](https://nexaai.com/apple/OpenELM-3B/gguf-q4_K_M/readme)                                       | NLP             | GGUF      | `nexa run openelm`                 |
| [nanollava](https://www.nexaai.com/qnguyen3/nanoLLaVA/gguf-fp16/readme)                                 | Multimodal      | GGUF      | `nexa run nanollava`               |
| [llava-phi3](https://www.nexaai.com/xtuner/llava-phi-3-mini/gguf-q4_0/readme)                           | Multimodal      | GGUF      | `nexa run llava-phi3`              |
| [llava-llama3](https://www.nexaai.com/xtuner/llava-llama-3-8b-v1.1/gguf-q4_0/readme)                    | Multimodal      | GGUF      | `nexa run llava-llama3`            |
| [llava1.6-mistral](https://www.nexaai.com/liuhaotian/llava-v1.6-mistral-7b/gguf-q4_0/readme)            | Multimodal      | GGUF      | `nexa run llava1.6-mistral`        |
| [llava1.6-vicuna](https://www.nexaai.com/liuhaotian/llava-v1.6-vicuna-7b/gguf-q4_0/readme)              | Multimodal      | GGUF      | `nexa run llava1.6-vicuna`         |
| [stable-diffusion-v1-4](https://www.nexaai.com/runwayml/stable-diffusion-v1-4/gguf-q4_0/readme)         | Computer Vision | GGUF      | `nexa run sd1-4`                   |
| [stable-diffusion-v1-5](https://www.nexaai.com/runwayml/stable-diffusion-v1-5/gguf-q4_0/readme)         | Computer Vision | GGUF/ONNX | `nexa run sd1-5`                   |
| [stable-diffusion-v2-1](https://nexaai.com/StabilityAI/stable-diffusion-v2-1/gguf-q4_0/readme)          | Computer Vision | GGUF      | `nexa run sd2-1`                   |
| [stable-diffusion-3-medium](https://nexaai.com/StabilityAI/stable-diffusion-3-medium/gguf-q4_0/readme)  | Computer Vision | GGUF      | `nexa run sd3`                     |
| [FLUX.1-schnell](https://nexaai.com/BlackForestLabs/FLUX.1-schnell/gguf-q4_0/readme)                    | Computer Vision | GGUF      | `nexa run flux`                    |
| [lcm-dreamshaper](https://www.nexaai.com/SimianLuo/lcm-dreamshaper-v7/gguf-fp16/readme)                 | Computer Vision | GGUF/ONNX | `nexa run lcm-dreamshaper`         |
| [hassaku-lcm](https://nexaai.com/stablediffusionapi/hassaku-hentai-model-v13-LCM/gguf-fp16/readme)      | Computer Vision | GGUF      | `nexa run hassaku-lcm`             |
| [anything-lcm](https://www.nexaai.com/Linaqruf/anything-v30-LCM/gguf-fp16/readme)                       | Computer Vision | GGUF      | `nexa run anything-lcm`            |
| [faster-whisper-tiny](https://www.nexaai.com/Systran/faster-whisper-tiny/bin-cpu-fp16/readme)           | Audio           | BIN       | `nexa run faster-whisper-tiny`     |
| [faster-whisper-small](https://www.nexaai.com/Systran/faster-whisper-small/bin-cpu-fp16/readme)         | Audio           | BIN       | `nexa run faster-whisper-small`    |
| [faster-whisper-medium](https://www.nexaai.com/Systran/faster-whisper-medium/bin-cpu-fp16/readme)       | Audio           | BIN       | `nexa run faster-whisper-medium`   |
| [faster-whisper-base](https://www.nexaai.com/Systran/faster-whisper-base/bin-cpu-fp16/readme)           | Audio           | BIN       | `nexa run faster-whisper-base`     |
| [faster-whisper-large](https://www.nexaai.com/Systran/faster-whisper-large-v3/bin-cpu-fp16/readme)      | Audio           | BIN       | `nexa run faster-whisper-large`    |
| [whisper-tiny.en](https://nexaai.com/openai/whisper-tiny.en/onnx-cpu-fp32/readme)                       | Audio           | ONNX      | `nexa run whisper-tiny.en`         |
| [whisper-tiny](https://nexaai.com/openai/whisper-tiny/onnx-cpu-fp32/readme)                             | Audio           | ONNX      | `nexa run whisper-tiny`            |
| [whisper-small.en](https://nexaai.com/openai/whisper-small.en/onnx-cpu-fp32/readme)                     | Audio           | ONNX      | `nexa run whisper-small.en`        |
| [whisper-small](https://nexaai.com/openai/whisper-small/onnx-cpu-fp32/readme)                           | Audio           | ONNX      | `nexa run whisper-small`           |
| [whisper-base.en](https://nexaai.com/openai/whisper-base.en/onnx-cpu-fp32/readme)                       | Audio           | ONNX      | `nexa run whisper-base.en`         |
| [whisper-base](https://nexaai.com/openai/whisper-base/onnx-cpu-fp32/readme)                             | Audio           | ONNX      | `nexa run whisper-base`            |

## CLI Reference

Here's a brief overview of the main CLI commands:

- `nexa run`: Run inference for various tasks using GGUF models.
- `nexa onnx`: Run inference for various tasks using ONNX models.
- `nexa server`: Run the Nexa AI Text Generation Service.
- `nexa pull`: Pull a model from official or hub.
- `nexa remove`: Remove a model from local machine.
- `nexa clean`: Clean up all model files.
- `nexa list`: List all models in the local machine.
- `nexa login`: Login to Nexa API.
- `nexa whoami`: Show current user information.
- `nexa logout`: Logout from Nexa API.

For detailed information on CLI commands and usage, please refer to the [CLI Reference](CLI.md) document.

## Start Local Server

To start a local server using models on your local computer, you can use the `nexa server` command.
For detailed information on server setup, API endpoints, and usage examples, please refer to the [Server Reference](SERVER.md) document.

## Acknowledgements

We would like to thank the following projects:

- [llama.cpp](https://github.com/ggerganov/llama.cpp)
- [stable-diffusion.cpp](https://github.com/leejet/stable-diffusion.cpp)
- [optimum](https://github.com/huggingface/optimum)
