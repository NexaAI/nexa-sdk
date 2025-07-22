<h1>Nexa SDK - Local On-Device Inference Framework</h1>
Nexa SDK is a comprehensive toolkit for supporting GGUF and MLX model formats. It supports LLM and VLMs.

![Nexa SDK](assets/banner.png)

Features
- Device Support: CPU, GPU (CUDA, Metal, Vulkan)
- Input Type Support: Text, Image, Audio
- Server: OpenAI-compatible API, JSON schema for function calling and streaming support
- Model Format Support: GGUF, MLX

## Latest News 🔥
- Beta launch for Nexa SDK, more updates coming soon!
  
> Welcome to submit your requests through issues, we ship weekly.

## Installation
- [MacOS](https://github.com/NexaAI/nexa-sdk/releases/latest/nexa-cli-universal.pkg)
- [Windows](https://github.com/NexaAI/nexa-sdk/releases/latest/nexa-cli_windows-setup.exe)
- [Linux](release/linux/install.sh)


## 🚀 Supported Model Types
GGUF runs on **macOS, Linux, and Windows**. MLX is **macOS-only (Apple Silicon)**.

| Type                | GGUF (Universal)  Quickstart                 | MLX (macOS only) Quickstart                     |
|---------------------|----------------------------------|------------------------------------------|
| **LLM**             | `nexa infer ggml-org/Qwen3-1.7B-GGUF`       | `nexa infer NexaAI/Qwen3-4B-4bit-MLX`               |
| **Multimodal (VLM)**| `nexa infer NexaAI/Qwen2.5-Omni-3B-GGUF`    | `nexa infer NexaAI/gemma-3n-E4B-it-4bit-MLX`        |



## 🤗 Run Models from HuggingFace
You can run any compatible GGUF or MLX model from Hugging Face by using the **full repo name**.

### GGUF models
To try other GGUF models, go to Hugging Face, find any model with GGUF format (e.g. unsloth/Qwen2.5-VL-3B-Instruct-GGUF), and run:

```bash
nexa infer unsloth/Qwen2.5-VL-3B-Instruct-GGUF
```

### MLX models
Many MLX models in the Hugging Face mlx-community organization have quality issues and may not run reliably.
We recommend starting with models from our curated [NexaAI Collection](https://huggingface.co/NexaAI/collections) for best results:

```bash
nexa infer NexaAI/Qwen3-4B-4bit-MLX
```

## 🛠️ Essential CLI Commands

| Command                          | What it does                                                        |
|----------------------------------|----------------------------------------------------------------------|
| `nexa -h`              | show all CLI commands                              |
| `nexa pull <repo>`              | Interactive download & cache of a model                              |
| `nexa infer <repo>`             | Local inference          |
| `nexa list`                     | Show all cached models with sizes                                    |
| `nexa remove <repo>` / `nexa clean` | Delete one / all cached models                                   |
| `nexa serve --host 127.0.0.1:8080` | Launch OpenAI‑compatible REST server                            |
| `nexa run <repo>`              | Chat with a model via an existing server                             |

👉 For comprehensive commands, see the full [CLI Reference →](https://nexaai.mintlify.app/nexa-sdk-go/NexaCLI)


## Acknowledgements
We would like to thank the following projects:
- [llama.cpp](https://github.com/ggml-org/llama.cpp)
- [mlx-lm](https://github.com/ml-explore/mlx-lm)
- [mlx-vlm](https://github.com/Blaizzy/mlx-vlm)
- [mlx-audio](https://github.com/Blaizzy/mlx-audio)
