<div align="center">
  <p>
      <img width="100%" src="assets/banner1.png" alt="Nexa AI Banner">
      <div align="center">
  <p style="font-size: 1.3em; font-weight: 600; margin-bottom: 10px;">🤝 Trusted by Partners</p>
  <img src="assets/nvidia.svg" alt="NVIDIA" height="40" style="margin: 0 20px;">
  <img src="assets/amd.svg" alt="AMD" height="42" style="margin: 0 20px;">
  <img src="assets/qualcomm.png" alt="Qualcomm" height="40" style="margin: 0 20px;">
  <img src="assets/intel.svg" alt="Intel" height="45" style="margin: 0 10px;">
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
| Unified backend (NPU / GPU / CPU) | ✅ NPU, GPU, CPU | ⚠️ CPU, GPU | ⚠️ CPU, GPU | ⚠️ CPU, GPU |
| Support any model in GGUF / MLX / .nexa format | ✅ Low-level control | ❌ | ⚠️ Limited | ❌ |
| Multi-modality (Text / Image / Audio) | ✅ Full Multimodal Support | ⚠️ Limited | ⚠️ Limited | ⚠️ Limited |
| Model format flexibility | ✅ GGUF, MLX, .nexa format | ⚠️ GGUF | ⚠️ GGUF | ⚠️ GGUF & MLX |
| Cross-platform | ✅ Desktop, Mobile, Automotive, IoT | ⚠️ Desktop | ⚠️ Desktop | ⚠️ Desktop |
| Easy install & One line of code to run | ✅ | ✅ | ⚠️ Takes more time| ✅ |
| OpenAI-compatible API + Function calling | ✅ | ✅ | ✅ | ✅ |

<p align="center" style="margin-top:14px">
  <i>
      <b>Legend:</b>
      <span title="Full support">✅ Supported</span> &nbsp; | &nbsp;
      <span title="Partial or limited support">⚠️ Partial</span> &nbsp; | &nbsp;
      <span title="Not Supported">❌ No</span>
  </i>
</p>
</div>


## Recent updates

#### 📣  **2025.10.14: Day-0 Support : Qwen3-VL-4B-Instruct, Qwen3-VL-4B-Thinking, Qwen3-VL-8B-Instruct, Qwen3-VL-8B-Thinking**
- We support [Qwen3-VL-4B series models](https://huggingface.co/collections/NexaAI/qwen3vl-68d46de18fdc753a7295190a) with Nexa SDK on Day-0!
- We support Qualcomm NPU/GPU/CPU, Apple GPU/CPU, Intel/AMD/MediaTek/Nvidia GPU/CPUs and more

#### 📣  **2025.10.04: Day-0 Support : Qwen3-VL-30B-A3B-Instruct**
- We support [Qwen3-VL-30B-A3B-Instruct](https://huggingface.co/NexaAI/qwen3vl-30B-A3B-mlx) with Nexa SDK on Day-0!
- Try it on Apple GPU with `nexa infer NexaAI/qwen3vl-30B-A3B-mlx` on MLX backend.

#### 📣  **2025.10.02: Day-0 Support on NPU/GPU/CPU : IBM Granite 4.0**
- We support [IBM Granite 4.0](https://sdk.nexa.ai/model/Granite-4-Micro) with Nexa SDK on Day-0!
- Try it on AMD / Intel / Qualcomm / Apple GPU with `nexa infer NexaAI/granite-4.0-micro-GGUF` and on Qualcomm NPU with `nexa infer NexaAI/Granite-4-Micro-NPU`

#### 📣  **2025.10.01: AMD NPU Support**
- Image Generation with [SDXL](https://huggingface.co/NexaAI/sdxl-turbo-amd-npu) on AMD NPU

#### 📣  **2025.09.23: Intel NPU Support**
- LLM inference with [DeepSeek-r1-distill-Qwen-1.5B](https://sdk.nexa.ai/model/DeepSeek-R1-Distill-Qwen-1.5B-Intel-NPU) and [Llama3.2-3B](https://sdk.nexa.ai/model/Llama3.2-3B-Intel-NPU) on Intel NPU

#### 📣  **2025.09.22: Apple Neural Engine (ANE) Support**
- Real-time speech recognition with [Parakeet v3 model](https://sdk.nexa.ai/model/parakeet-v3-ane)

#### 📣  **2025.09.15: New Models Support**
- First-ever [Gemma-3n](https://sdk.nexa.ai/model/Gemma3n-E4B) **multimodal** inference for GPU & CPU, in GGUF format.
- [SDXL image generation](https://sdk.nexa.ai/model/Prefect-illustrious-XL-v2.0p) from Civitai for GPU
- [EmbeddingGemma](https://sdk.nexa.ai/model/embeddinggemma-300m-npu) for Qualcomm NPU
- [Phi4-mini turbo](https://sdk.nexa.ai/model/phi4-mini-npu-turbo) and [Phi3.5-mini](https://sdk.nexa.ai/model/phi3.5-mini-npu) for Qualcomm NPU
- [Parakeet V3 model](https://sdk.nexa.ai/model/parakeet-v3-npu) for Qualcomm NPU

#### 📣  **2025.09.05: Turbo Engine & Unified Interface**
- [Nexa ML Turbo engine](https://nexa.ai/blogs/nexaml-turbo) for optimized NPU performance
    - Try [Phi4-mini turbo](https://sdk.nexa.ai/model/phi4-mini-npu-turbo) and [Llama3.2-3B-NPU-Turbo](https://sdk.nexa.ai/model/Llama3.2-3B-NPU-Turbo)
    - 80% faster at shorter contexts (<=2048), 33% faster at longer contexts (>2048) than current NPU solutions
- [Unified interface](https://nexa.ai/blogs/sdk-unifiedarchitecture) supporting NPU/GPU/CPU backends:
    - Single installer architecture eliminating dependency conflicts
    - Lazy loading and plugin isolation for improved performance

#### 📣  **2025.08.20: Qualcomm NPU Support with NexaML Turbo Engine**
- OmniNeural-4B: the **first multimodal AI model built natively for NPUs** — handling text, images, and audio in one model
- Check the model and demos at [Hugging Face repo](https://huggingface.co/NexaAI/OmniNeural-4B)
- Check our [OmniNeural-4B technical blog](https://nexa.ai/blogs/omnineural-4b)

#### 📣  **2025.08.12: ASR & TTS Support in MLX format**
- Parakeet and Kokoro models support in MLX format.
- new `/mic` mode to transcribe live speech directly in your terminal.
  
## Installation

### macOS
* [arm64 with Apple Neural Engine support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_arm64.pkg)
* [x86_64](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_x86_64.pkg)

### Windows
* [arm64 with Qualcomm NPU support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_arm64.exe)
* [x86_64 with Intel / AMD NPU support](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_x86_64.exe)


### Linux
```bash
curl -fsSL https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_x86_64.sh -o install.sh && chmod +x install.sh && ./install.sh && rm install.sh
```

## Supported Models

You can run any compatible GGUF, MLX, or nexa model from 🤗 Hugging Face by using the `<full repo name>`.

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
nexa infer omni-neural
nexa infer NexaAI/OmniNeural-4B
nexa infer NexaAI/qwen3-1.7B-npu
```


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
