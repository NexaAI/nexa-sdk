<div align="center">
  <p>
      <img width="100%" src="assets/banner.png" alt="Nexa AI Banner">
  </p>

  <p align="center">
    <a href="https://docs.nexa.ai">
        <img src="https://img.shields.io/badge/docs-website-brightgreen?logo=readthedocs" alt="Documentation">
    </a>
   <a href="https://x.com/nexa_ai"><img alt="X account" src="https://img.shields.io/twitter/url/https/twitter.com/diffuserslib.svg?style=social&label=Follow%20%40Nexa_AI"></a>
    <a href="https://discord.com/invite/nexa-ai">
        <img src="https://img.shields.io/discord/1192186167391682711?color=5865F2&logo=discord&logoColor=white&style=flat-square" alt="Join us on Discord">
    </a>
    <a href="https://join.slack.com/t/nexa-ai-community/shared_invite/zt-3837k9xpe-LEty0disTTUnTUQ4O3uuNw">
        <img src="https://img.shields.io/badge/slack-join%20chat-4A154B?logo=slack&logoColor=white" alt="Join us on Slack">
    </a>
</p>
  
  ![OS](https://img.shields.io/badge/os-linux%20|%20macOS%20|%20windows-purple)
  ![Hardware](https://img.shields.io/badge/hardware-CPU%20|%20GPU%20|%20NPU-yellow)

</div>

# Nexa SDK

Nexa SDK is an on-device inference framework that runs any model on any device, across any backend. It runs on CPUs and GPUs with backend support for CUDA, Metal, and Vulkan. It handles multiple input modalities including text 📝, image 🖼️, and audio 🎧. The SDK includes an OpenAI-compatible API server with support for JSON schema-based function calling and streaming. It supports model formats such as GGUF and MLX, enabling efficient quantized inference across diverse platforms.

## Recent updates
#### 📣  **2025.08.12: ASR & TTS Support in MLX format
- ASR & TTS model support in MLX format.
- new "> /mic" mode to transcribe live speech directly in your terminal.
  
## Installation

### macOS
* [arm64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_macos_arm64.pkg)
* [x86_64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_macos_x86_64.pkg)

### Windows
* [x86_64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_windows_x86_64.exe)

### Linux
```bash
curl -fsSL https://raw.githubusercontent.com/NexaAI/nexa-sdk/main/release/linux/install.sh -o install.sh && chmod +x install.sh && ./install.sh
```

## Supported Models

You can run any compatible GGUF or MLX model from 🤗 Hugging Face by using the `<full repo name>`.

### GGUF models

> [!TIP]
> GGUF runs on macOS, Linux, and Windows.

📝 Run and chat with LLMs, e.g. Qwen3:

```bash
nexa infer ggml-org/Qwen3-1.7B-GGUF
```

🖼️ Run and chat with Multimodal models, e.g. Qwen2.5-Omni:

```bash
nexa infer NexaAI/Qwen2.5-Omni-3B-GGUF
```

### MLX models
> [!TIP]
> MLX is macOS-only (Apple Silicon). Many MLX models in the Hugging Face mlx-community organization have quality issues and may not run reliably.
> We recommend starting with models from our curated [NexaAI Collection](https://huggingface.co/NexaAI/collections) for best results. For example

📝 Run and chat with LLMs, e.g. Qwen3:

```bash
nexa infer NexaAI/Qwen3-4B-4bit-MLX
```

🖼️ Run and chat with Multimodal models, e.g. Gemma3n:

```bash
nexa infer NexaAI/gemma-3n-E4B-it-4bit-MLX
```

## CLI Reference

| Essential Command                          | What it does                                                        |
|----------------------------------|----------------------------------------------------------------------|
| `nexa -h`              | show all CLI commands                              |
| `nexa pull <repo>`              | Interactive download & cache of a model                              |
| `nexa infer <repo>`             | Local inference          |
| `nexa list`                     | Show all cached models with sizes                                    |
| `nexa remove <repo>` / `nexa clean` | Delete one / all cached models                                   |
| `nexa serve --host 127.0.0.1:8080` | Launch OpenAI‑compatible REST server                            |
| `nexa run <repo>`              | Chat with a model via an existing server                             |

👉 To interact with multimodal models, you can drag photos or audio clips directly into the CLI — you can even drop multiple images at once!

See [CLI Reference](https://nexaai.mintlify.app/nexa-sdk-go/NexaCLI) for full commands.

## Acknowledgements

We would like to thank the following projects:
- [llama.cpp](https://github.com/ggml-org/llama.cpp)
- [mlx-lm](https://github.com/ml-explore/mlx-lm)
- [mlx-vlm](https://github.com/Blaizzy/mlx-vlm)
- [mlx-audio](https://github.com/Blaizzy/mlx-audio)
