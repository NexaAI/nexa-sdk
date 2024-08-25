<div align="center">

<h1>Nexa SDK</h1>

  <img src="./assets/banner.png" alt="icon"/>

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

<details>
<summary><strong>FAQ: Building Issues for llava</strong></summary>

If you encounter the following issue while building:

![](docs/.media/error.jpeg)

try the following command:

```bash
CMAKE_ARGS="-DCMAKE_CXX_FLAGS=-fopenmp" pip install nexaai
```

</details>

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

## Supported Models

| Model                                                                                                   | Type            | Format    | Command                            |
| ------------------------------------------------------------------------------------------------------- | --------------- | --------- | ---------------------------------- |
| [octopus-v2](https://www.nexaai.com/NexaAI/Octopus-v2/gguf-q4_0/readme)                                 | NLP             | GGUF      | `nexa run octopus-v2`              |
| [octopus-v4](https://www.nexaai.com/NexaAI/Octopus-v4/gguf-q4_0/readme)                                 | NLP             | GGUF      | `nexa run octopus-v4`              |
| [tinyllama](https://www.nexaai.com/TinyLlama/TinyLlama-1.1B-Chat-v1.0/gguf-fp16/readme)                 | NLP             | GGUF      | `nexa run tinyllama`               |
| [llama2](https://www.nexaai.com/meta/Llama2-7b-chat/gguf-q4_0/readme)                                   | NLP             | GGUF/ONNX | `nexa run llama2`                  |
| [llama3](https://www.nexaai.com/meta/Llama3-8B-Instruct/gguf-q4_0/readme)                               | NLP             | GGUF/ONNX | `nexa run llama3`                  |
| [llama3.1](https://www.nexaai.com/meta/Llama3.1-8B-Instruct/gguf-q4_0/readme)                           | NLP             | GGUF/ONNX | `nexa run llama3.1`                |
| [gemma](https://www.nexaai.com/google/gemma-1.1-2b-instruct/gguf-q4_0/readme)                           | NLP             | GGUF/ONNX | `nexa run gemma`                   |
| [gemma2](https://www.nexaai.com/google/gemma-2-2b-instruct/gguf-q4_0/readme)                            | NLP             | GGUF      | `nexa run gemma2`                  |
| [qwen1.5](https://www.nexaai.com/Qwen/Qwen1.5-7B-Instruct/gguf-q4_0/readme)                             | NLP             | GGUF      | `nexa run qwen1.5`                 |
| [qwen2](https://www.nexaai.com/Qwen/Qwen2-1.5B-Instruct/gguf-q4_0/readme)                               | NLP             | GGUF/ONNX | `nexa run qwen2`                   |
| [mistral](https://www.nexaai.com/mistralai/Mistral-7B-Instruct-v0.3/gguf-q4_0/readme)                   | NLP             | GGUF/ONNX | `nexa run mistral`                 |
| [codegemma](https://www.nexaai.com/google/codegemma-2b/gguf-q4_0/readme)                                | NLP             | GGUF      | `nexa run codegemma`               |
| [codellama](https://www.nexaai.com/meta/CodeLlama-7b-Instruct/gguf-q2_K/readme)                         | NLP             | GGUF      | `nexa run codellama`               |
| [codeqwen](https://www.nexaai.com/Qwen/CodeQwen1.5-7B-Instruct/gguf-q4_0/readme)                        | NLP             | GGUF      | `nexa run codeqwen`                |
| [deepseek-coder](https://www.nexaai.com/DeepSeek/deepseek-coder-1.3b-instruct/gguf-q4_0/readme)         | NLP             | GGUF      | `nexa run deepseek-coder`          |
| [dolphin-mistral](https://www.nexaai.com/CognitiveComputations/dolphin-2.8-mistral-7b/gguf-q4_0/readme) | NLP             | GGUF      | `nexa run dolphin-mistral`         |
| [phi2](https://www.nexaai.com/microsoft/Phi-2/gguf-q4_0/readme)                                         | NLP             | GGUF      | `nexa run phi2`                    |
| [phi3](https://www.nexaai.com/microsoft/Phi-3-mini-128k-instruct/gguf-q4_0/readme)                      | NLP             | GGUF/ONNX | `nexa run phi3`                    |
| [llama2-uncensored](https://www.nexaai.com/georgesung/Llama2-7b-chat-uncensored/gguf-q4_0/readme)       | NLP             | GGUF      | `nexa run llama2-uncensored`       |
| [llama3-uncensored](https://www.nexaai.com/Orenguteng/Llama3-8B-Lexi-Uncensored/gguf-q4_K_M/readme)     | NLP             | GGUF      | `nexa run llama3-uncensored`       |
| [llama2-function-calling](https://www.nexaai.com/Trelis/Llama2-7b-function-calling/gguf-q4_K_M/readme)  | NLP             | GGUF      | `nexa run llama2-function-calling` |
| [nanollava](https://www.nexaai.com/qnguyen3/nanoLLaVA/gguf-fp16/readme)                                 | Multimodal      | GGUF      | `nexa run nanollava`               |
| [llava-phi3](https://www.nexaai.com/xtuner/llava-phi-3-mini/gguf-q4_0/readme)                           | Multimodal      | GGUF      | `nexa run llava-phi3`              |
| [llava-llama3](https://www.nexaai.com/xtuner/llava-llama-3-8b-v1.1/gguf-q4_0/readme)                    | Multimodal      | GGUF      | `nexa run llava-llama3`            |
| [llava1.6-mistral](https://www.nexaai.com/liuhaotian/llava-v1.6-mistral-7b/gguf-q4_0/readme)            | Multimodal      | GGUF      | `nexa run llava1.6-mistral`        |
| [llava1.6-vicuna](https://www.nexaai.com/liuhaotian/llava-v1.6-vicuna-7b/gguf-q4_0/readme)              | Multimodal      | GGUF      | `nexa run llava1.6-vicuna`         |
| [stable-diffusion-v1-4](https://www.nexaai.com/runwayml/stable-diffusion-v1-4/gguf-q4_0/readme)         | Computer Vision | GGUF      | `nexa run sd1-4`                   |
| [stable-diffusion-v1-5](https://www.nexaai.com/runwayml/stable-diffusion-v1-5/gguf-q4_0/readme)         | Computer Vision | GGUF/ONNX | `nexa run sd1-5`                   |
| [lcm-dreamshaper](https://www.nexaai.com/SimianLuo/lcm-dreamshaper-v7/gguf-fp16/readme)                 | Computer Vision | GGUF/ONNX | `nexa run lcm-dreamshaper`         |
| [hassaku-lcm](https://nexaai.com/stablediffusionapi/hassaku-hentai-model-v13-LCM/gguf-fp16/readme)      | Computer Vision | GGUF      | `nexa run hassaku-lcm`             |
| [anything-lcm](https://www.nexaai.com/Linaqruf/anything-v30-LCM/gguf-fp16/readme)                       | Computer Vision | GGUF      | `nexa run anything-lcm`            |
| [faster-whisper-tiny](https://www.nexaai.com/Systran/faster-whisper-tiny/bin-cpu-fp16/readme)           | Audio           | BIN       | `nexa run faster-whisper-tiny`     |
| [faster-whisper-small](https://www.nexaai.com/Systran/faster-whisper-small/bin-cpu-fp16/readme)         | Audio           | BIN       | `nexa run faster-whisper-small`    |
| [faster-whisper-medium](https://www.nexaai.com/Systran/faster-whisper-medium/bin-cpu-fp16/readme)       | Audio           | BIN       | `nexa run faster-whisper-medium`   |
| [faster-whisper-base](https://www.nexaai.com/Systran/faster-whisper-base/bin-cpu-fp16/readme)           | Audio           | BIN       | `nexa run faster-whisper-base`     |
| [faster-whisper-large](https://www.nexaai.com/Systran/faster-whisper-large-v3/bin-cpu-fp16/readme)      | Audio           | BIN       | `nexa run faster-whisper-large`    |

## CLI Reference

```
usage: nexa [-h] [-V] {run,onnx,server,pull,remove,clean,list,login,whoami,logout} ...

Nexa CLI tool for handling various model operations.

positional arguments:
  {run,onnx,server,pull,remove,clean,list,login,whoami,logout}
                        sub-command help
    run                 Run inference for various tasks using GGUF models.
    onnx                Run inference for various tasks using ONNX models.
    server              Run the Nexa AI Text Generation Service
    pull                Pull a model from official or hub.
    remove              Remove a model from local machine.
    clean               Clean up all model files.
    list                List all models in the local machine.
    login               Login to Nexa API.
    whoami              Show current user information.
    logout              Logout from Nexa API.

options:
  -h, --help            show this help message and exit
  -V, --version         Show the version of the Nexa SDK.
```

### List Local Models

List all models on your local computer.

```
nexa list
```

### Download a Model

Download a model file to your local computer from Nexa Model Hub.

```
nexa pull MODEL_PATH
usage: nexa pull [-h] model_path

positional arguments:
  model_path  Path or identifier for the model in Nexa Model Hub

options:
  -h, --help  show this help message and exit
```

#### Example

```
nexa pull llama2
```

### Remove a Model

Remove a model from your local computer.

```
nexa remove MODEL_PATH
usage: nexa remove [-h] model_path

positional arguments:
  model_path  Path or identifier for the model in Nexa Model Hub

options:
  -h, --help  show this help message and exit
```

#### Example

```
nexa remove llama2
```

### Remove All Downloaded Models

Remove all downloaded models on your local computer.

```
nexa clean
```

### Run a Model

Run a model on your local computer. If the model file is not yet downloaded, it will be automatically fetched first.

By default, `nexa` will run gguf models. To run onnx models, use `nexa onnx MODEL_PATH`

#### Run Text-Generation Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-t TEMPERATURE] [-m MAX_NEW_TOKENS] [-k TOP_K] [-p TOP_P] [-sw [STOP_WORDS ...]] [-pf] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -pf, --profiling      Enable profiling logs for the inference process
  -st, --streamlit      Run the inference in Streamlit UI

Text generation options:
  -t, --temperature TEMPERATURE
                        Temperature for sampling
  -m, --max_new_tokens MAX_NEW_TOKENS
                        Maximum number of new tokens to generate
  -k, --top_k TOP_K     Top-k sampling parameter
  -p, --top_p TOP_P     Top-p sampling parameter
  -sw, --stop_words [STOP_WORDS ...]
                        List of stop words for early stopping
```

##### Example

```
nexa run llama2
```

#### Run Image-Generation Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-i2i] [-ns NUM_INFERENCE_STEPS] [-np NUM_IMAGES_PER_PROMPT] [-H HEIGHT] [-W WIDTH] [-g GUIDANCE_SCALE] [-o OUTPUT] [-s RANDOM_SEED] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -st, --streamlit      Run the inference in Streamlit UI

Image generation options:
  -i2i, --img2img       Whether to run image-to-image generation
  -ns, --num_inference_steps NUM_INFERENCE_STEPS
                        Number of inference steps
  -np, --num_images_per_prompt NUM_IMAGES_PER_PROMPT
                        Number of images to generate per prompt
  -H, --height HEIGHT   Height of the output image
  -W, --width WIDTH     Width of the output image
  -g, --guidance_scale GUIDANCE_SCALE
                        Guidance scale for diffusion
  -o, --output OUTPUT   Output path for the generated image
  -s, --random_seed RANDOM_SEED
                        Random seed for image generation
  --lora_dir LORA_DIR   Path to directory containing LoRA files
  --wtype WTYPE         Weight type (f32, f16, q4_0, q4_1, q5_0, q5_1, q8_0)
  --control_net_path CONTROL_NET_PATH
                        Path to control net model
  --control_image_path CONTROL_IMAGE_PATH
                        Path to image condition for Control Net
  --control_strength CONTROL_STRENGTH
                        Strength to apply Control Net
```

##### Example

```
nexa run sd1-4
```

#### Run Vision-Language Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-t TEMPERATURE] [-m MAX_NEW_TOKENS] [-k TOP_K] [-p TOP_P] [-sw [STOP_WORDS ...]] [-pf] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -pf, --profiling      Enable profiling logs for the inference process
  -st, --streamlit      Run the inference in Streamlit UI

VLM generation options:
  -t, --temperature TEMPERATURE
                        Temperature for sampling
  -m, --max_new_tokens MAX_NEW_TOKENS
                        Maximum number of new tokens to generate
  -k, --top_k TOP_K     Top-k sampling parameter
  -p, --top_p TOP_P     Top-p sampling parameter
  -sw, --stop_words [STOP_WORDS ...]
                        List of stop words for early stopping
```

##### Example

```
nexa run nanollava
```

#### Run Audio Model

```
nexa run MODEL_PATH
usage: nexa run [-h] [-o OUTPUT_DIR] [-b BEAM_SIZE] [-l LANGUAGE] [--task TASK] [-t TEMPERATURE] [-c COMPUTE_TYPE] [-st] model_path

positional arguments:
  model_path            Path or identifier for the model in Nexa Model Hub

options:
  -h, --help            show this help message and exit
  -st, --streamlit      Run the inference in Streamlit UI

Automatic Speech Recognition options:
  -b, --beam_size BEAM_SIZE
                        Beam size to use for transcription
  -l, --language LANGUAGE
                        The language spoken in the audio. It should be a language code such as 'en' or 'fr'.
  --task TASK           Task to execute (transcribe or translate)
  -c, --compute_type COMPUTE_TYPE
                        Type to use for computation (e.g., float16, int8, int8_float16)
```

##### Example

```
nexa run faster-whisper-tiny
```

### Start Local Server

Start a local server using models on your local computer.

```
nexa server MODEL_PATH
usage: nexa server [-h] [--host HOST] [--port PORT] [--reload] model_path

positional arguments:
  model_path   Path or identifier for the model in S3

options:
  -h, --help   show this help message and exit
  --host HOST  Host to bind the server to
  --port PORT  Port to bind the server to
  --reload     Enable automatic reloading on code changes
```

#### Example

```
nexa server llama2
```

### Model Path Format

For `model_path` in nexa commands, it's better to follow the standard format to ensure correct model loading and execution. The standard format for `model_path` is:

- `[user_name]/[repo_name]:[tag_name]` (user's model)
- `[repo_name]:[tag_name]` (official model)

#### Examples:

- `gemma-2b:q4_0`
- `Meta-Llama-3-8B-Instruct:onnx-cpu-int8`
- `alanzhuly/Qwen2-1B-Instruct:q4_0`

## Start Local Server

You can start a local server using models on your local computer with the `nexa server` command. Here's the usage syntax:

```
usage: nexa server [-h] [--host HOST] [--port PORT] [--reload] model_path
```

### Options:

- `--host`: Host to bind the server to
- `--port`: Port to bind the server to
- `--reload`: Enable automatic reloading on code changes

### Example Commands:

```
nexa server gemma
nexa server llama2-function-calling
nexa server sd1-5
nexa server faster-whipser-large
```

By default, `nexa server` will run gguf models. To run onnx models, simply add `onnx` after `nexa server`.

## API Endpoints

<details>
<summary><strong>1. Text Generation: <code>/v1/completions</code></strong></summary>
Generates text based on a single prompt.

#### Request body:

```json
{
  "prompt": "Tell me a story",
  "temperature": 1,
  "max_new_tokens": 128,
  "top_k": 50,
  "top_p": 1,
  "stop_words": ["string"]
}
```

#### Example Response:

```json
{
  "result": "Once upon a time, in a small village nestled among rolling hills..."
}
```

</details>

<details><summary><strong>2. Chat Completions: <code>/v1/chat/completions</code></strong></summary>

Handles chat completions with support for conversation history.

#### Request body:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Tell me a story"
    }
  ],
  "max_tokens": 128,
  "temperature": 0.1,
  "stream": false,
  "stop_words": []
}
```

#### Example Response:

```json
{
  "id": "f83502df-7f5a-4825-a922-f5cece4081de",
  "object": "chat.completion",
  "created": 1723441724.914671,
  "choices": [
    {
      "message": {
        "role": "assistant",
        "content": "In the heart of a mystical forest..."
      }
    }
  ]
}
```

</details>
<details><summary><strong>3. Function Calling: <code>/v1/function-calling</code></strong></summary>

Call the most appropriate function based on user's prompt.

#### Request body:

```json
{
  "messages": [
    {
      "role": "user",
      "content": "Extract Jason is 25 years old"
    }
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "UserDetail",
        "parameters": {
          "properties": {
            "name": {
              "description": "The user's name",
              "type": "string"
            },
            "age": {
              "description": "The user's age",
              "type": "integer"
            }
          },
          "required": ["name", "age"],
          "type": "object"
        }
      }
    }
  ],
  "tool_choice": "auto"
}
```

#### Function format:

```json
{
  "type": "function",
  "function": {
    "name": "function_name",
    "description": "function_description",
    "parameters": {
      "type": "object",
      "properties": {
        "property_name": {
          "type": "string | number | boolean | object | array",
          "description": "string"
        }
      },
      "required": ["array_of_required_property_names"]
    }
  }
}
```

#### Example Response:

```json
{
  "id": "chatcmpl-7a9b0dfb-878f-4f75-8dc7-24177081c1d0",
  "object": "chat.completion",
  "created": 1724186442,
  "model": "/home/ubuntu/.cache/nexa/hub/official/Llama2-7b-function-calling/q3_K_M.gguf",
  "choices": [
    {
      "finish_reason": "tool_calls",
      "index": 0,
      "logprobs": null,
      "message": {
        "role": "assistant",
        "content": null,
        "tool_calls": [
          {
            "id": "call__0_UserDetail_cmpl-8d5cf645-7f35-4af2-a554-2ccea1a67bdd",
            "type": "function",
            "function": {
              "name": "UserDetail",
              "arguments": "{ \"name\": \"Jason\", \"age\": 25 }"
            }
          }
        ],
        "function_call": {
          "name": "",
          "arguments": "{ \"name\": \"Jason\", \"age\": 25 }"
        }
      }
    }
  ],
  "usage": {
    "completion_tokens": 15,
    "prompt_tokens": 316,
    "total_tokens": 331
  }
}
```

</details>
<details><summary><strong>4. Text-to-Image: <code>/v1/txt2img</code></strong></summary>

Generates images based on a single prompt.

#### Request body:

```json
{
  "prompt": "A girl, standing in a field of flowers, vivid",
  "image_path": "",
  "cfg_scale": 7,
  "width": 256,
  "height": 256,
  "sample_steps": 20,
  "seed": 0,
  "negative_prompt": ""
}
```

#### Example Response:

```json
{
  "created": 1724186615.5426757,
  "data": [
    {
      "base64": "base64_of_generated_image",
      "url": "path/to/generated_image"
    }
  ]
}
```

</details>
<details><summary><strong>5. Image-to-Image: <code>/v1/img2img</code></strong></summary>

Modifies existing images based on a single prompt.

#### Request body:

```json
{
  "prompt": "A girl, standing in a field of flowers, vivid",
  "image_path": "path/to/image",
  "cfg_scale": 7,
  "width": 256,
  "height": 256,
  "sample_steps": 20,
  "seed": 0,
  "negative_prompt": ""
}
```

#### Example Response:

```json
{
  "created": 1724186615.5426757,
  "data": [
    {
      "base64": "base64_of_generated_image",
      "url": "path/to/generated_image"
    }
  ]
}
```

</details>
<details><summary><strong>6. Audio Transcriptions: <code>/v1/audio/transcriptions</code></strong></summary>

Transcribes audio files to text.

#### Parameters:

- `beam_size` (integer): Beam size for transcription (default: 5)
- `language` (string): Language code (e.g., 'en', 'fr')
- `temperature` (number): Temperature for sampling (default: 0)

#### Request body:

```
{
  "file" (form-data): The audio file to transcribe (required)
}
```

#### Example Response:

```json
{
  "text": " And so my fellow Americans, ask not what your country can do for you, ask what you can do for your country."
}
```

</details>
<details><summary><strong>7. Audio Translations: <code>/v1/audio/translations</code></strong></summary>

Translates audio files to text in English.

#### Parameters:

- `beam_size` (integer): Beam size for transcription (default: 5)
- `temperature` (number): Temperature for sampling (default: 0)

#### Request body:

```
{
  "file" (form-data): The audio file to transcribe (required)
}
```

#### Example Response:

```json
{
  "text": " Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday"
}
```

</details>
