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
