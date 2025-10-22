<div align="center">
  <p>
      <img width="100%" src="assets/banner1.png" alt="Nexa AI Banner">
      <div align="center">
  <p style="font-size: 1.3em; font-weight: 600; margin-bottom: 10px;">🤝 Trusted by Partners</p>
  <img src="assets/qualcomm.png" alt="Qualcomm" height="40" style="margin: 0 20px;">
  <img src="assets/nvidia.png" alt="NVIDIA" height="40" style="margin: 0 20px;">
  <img src="assets/AMD.png" alt="AMD" height="42" style="margin: 0 20px;">
  <img src="assets/Intel_logo.png" alt="Intel" height="45" style="margin: 0 10px;">
</div>
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


</div>

# NexaSDK - Run any AI model on any backend

NexaSDK is an easy-to-use developer toolkit for running any AI model locally — across NPUs, GPUs, and CPUs — powered by our NexaML engine, built entirely from scratch for peak performance on every hardware stack. Unlike wrappers that depend on existing runtimes, NexaML is a unified inference engine built at the kernel level. It’s what lets NexaSDK achieve Day-0 support for new model architectures (LLMs, multimodal, audio, vision). NexaML supports 3 model formats: GGUF, MLX, and Nexa AI's own `.nexa` format.

### ⚙️ Differentiation

<div align="center">

| Features | **NexaSDK** | **Ollama** | **llama.cpp** | **LM Studio** |
|----------|--------------|-------------|----------------|----------------|
| NPU support | ✅ NPU-first | ❌ | ❌ | ❌ |
| Support any model in GGUF, MLX, NEXA format | ✅ Low-level Control | ❌ | ⚠️ | ❌ |
| Full multimodality support | ✅ Image, Audio, Text | ⚠️ | ⚠️ | ⚠️ |
| Cross-platform support | ✅ Desktop, Mobile, Automotive, IoT | ⚠️ | ⚠️ | ⚠️ |
| One line of code to run | ✅ | ✅ | ⚠️ | ✅ |
| OpenAI-compatible API + Function calling | ✅ | ✅ | ✅ | ✅ |

<p align="center" style="margin-top:14px">
  <i>
      <b>Legend:</b>
      <span title="Full support">✅ Supported</span> &nbsp; | &nbsp;
      <span title="Partial or limited support">⚠️ Partial or limited support </span> &nbsp; | &nbsp;
      <span title="Not Supported">❌ No</span>
  </i>
</p>
</div>


## Recent Wins

- 📣 Day-0 Support for **Qwen3-VL-4B and 8B** in GGUF, MLX, .nexa format for NPU/GPU/CPU. We are the only framework that supports the GGUF format. [Featured in Qwen's post about our partnership](https://x.com/Alibaba_Qwen/status/1978154384098754943).
- 📣 Day-0 Support for **IBM Granite 4.0** on NPU/GPU/CPU. [NexaML engine were featured right next to vLLM, llama.cpp, and MLX in IBM's blog](https://x.com/IBM/status/1978154384098754943).
- 📣 Day-0 Support for **Google EmbeddingGemma** on NPU. We are [featured in Google's social post](https://x.com/googleaidevs/status/1969188152049889511).
- 📣 Supported **vision capability for Gemma3n**: First-ever [Gemma-3n](https://sdk.nexa.ai/model/Gemma3n-E4B) **multimodal** inference for GPU & CPU, in GGUF format.
- 📣 AMD NPU Support for [SDXL](https://huggingface.co/NexaAI/sdxl-turbo-amd-npu) image generation
- 📣 Intel NPU Support [DeepSeek-r1-distill-Qwen-1.5B](https://sdk.nexa.ai/model/DeepSeek-R1-Distill-Qwen-1.5B-Intel-NPU) and [Llama3.2-3B](https://sdk.nexa.ai/model/Llama3.2-3B-Intel-NPU)
- 📣 Apple Neural Engine Support for real-time speech recognition with [Parakeet v3 model](https://sdk.nexa.ai/model/parakeet-v3-ane)
  
# Quick Start

## Step 1: Download Nexa CLI with one click

### macOS
* [arm64 with Apple Neural Engine support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_arm64.pkg)
* [x86_64](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_x86_64.pkg)

### Windows
* [arm64 with Qualcomm NPU support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_arm64.exe)
* [x86_64 with Intel / AMD NPU support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_x86_64.exe)

### Linux
#### For x86_64:
```bash
curl -fsSL https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_x86_64.sh -o install.sh && chmod +x install.sh && ./install.sh && rm install.sh
```

#### For arm64:
```bash
curl -fsSL https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_arm64.sh -o install.sh && chmod +x install.sh && ./install.sh && rm install.sh
```

## Step 2: Run models with one line of code

You can run any compatible GGUF, MLX, or nexa model from 🤗 Hugging Face by using the `nexa infer <full repo name>`.

### GGUF models

> [!TIP]
> GGUF runs on macOS, Linux, and Windows on CPU/GPU. Note certain GGUF models are only supported by NexaSDK (e.g. Qwen3-VL-4B and 8B).

📝 Run and chat with LLMs, e.g. Qwen3:

```bash
nexa infer ggml-org/Qwen3-1.7B-GGUF
```

🖼️ Run and chat with Multimodal models, e.g. Qwen3-VL-4B:

```bash
nexa infer NexaAI/Qwen3-VL-4B-Instruct-GGUF
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

### Qualcomm NPU models
> [!TIP]
> You need to download the [arm64 with Qualcomm NPU support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_arm64.exe) and make sure you have Snapdragon® X Elite chip on your laptop.

#### Quick Start (Windows arm64, Snapdragon X Elite)

1. **Login & Get Access Token (required for Pro Models)**  
   - Create an account at [sdk.nexa.ai](https://sdk.nexa.ai)  
   - Go to **Deployment → Create Token**  
   - Run this once in your terminal (replace with your token):  
     ```bash
     nexa config set license '<your_token_here>'
     ```

2. Run and chat with our multimodal model, OmniNeural-4B, or other models on NPU

```bash
nexa infer NexaAI/OmniNeural-4B
nexa infer NexaAI/Granite-4-Micro-NPU
nexa infer NexaAI/Qwen3-VL-4B-Instruct-NPU
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
- [ggml](https://github.com/ggml-org/ggml)
- [mlx-lm](https://github.com/ml-explore/mlx-lm)
- [mlx-vlm](https://github.com/Blaizzy/mlx-vlm)
- [mlx-audio](https://github.com/Blaizzy/mlx-audio)
