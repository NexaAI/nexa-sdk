<div align="center" style="text-decoration: none;">
  <img width="100%" src="assets/banner1.png" alt="Nexa AI Banner">
  <p style="font-size: 1.3em; font-weight: 600; margin-bottom: 20px;"> 
    <a href="README_zh.md"> 简体中文 </a>
    |
    <a href="README.md"> English </a>
  </p>
  <p style="font-size: 1.3em; font-weight: 600; margin-bottom: 20px;">🤝 Supported chipmakers </p>
    <picture>
      <source srcset="assets/chipmakers-dark.png" media="(prefers-color-scheme: dark)">
      <source srcset="assets/chipmakers.png" media="(prefers-color-scheme: light)">
      <img src="assets/chipmakers.png" style="max-height:30px; height:auto; width:auto;">
    </picture>
  </p>
  <p>
    <a href="https://www.producthunt.com/products/nexasdk-for-mobile?embed=true&utm_source=badge-top-post-badge&utm_medium=badge&utm_campaign=badge-nexasdk-for-mobile" target="_blank" rel="noopener noreferrer">
        <img alt="NexaSDK for Mobile - #1 Product of the Day" width="180" height="39" src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=1049998&theme=dark&period=daily&t=1765991451976">
    </a>
    <a href="https://trendshift.io/repositories/12239" target="_blank" rel="noopener noreferrer">
        <img alt="NexaAI/nexa-sdk - #1 Repository of the Day" height="39" src="https://trendshift.io/api/badge/repositories/12239">
    </a>
  </p>
  <p>
    <a href="https://docs.nexa.ai">
        <img src="https://img.shields.io/badge/docs-website-brightgreen?logo=readthedocs" alt="Documentation">
    </a>
    <a href="https://sdk.nexa.ai/wishlist">
        <img src="https://img.shields.io/badge/🎯_Vote_for-Next_Models-ff69b4?style=flat-square" alt="Vote for Next Models">
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

# NexaSDK

**NexaSDK lets you build the smartest and fastest on-device AI with minimum energy.** It runs latest AI models locally on NPU, GPU, and CPU - across Android, Windows, Linux, macOS, and iOS devices with a few lines of code. 

NexaSDK supports latest models **weeks or months before anyone else** — Qwen3-VL, DeepSeek-OCR, Gemma3n (Vision), and more.

> ⭐ **Star this repo** to keep up with exciting updates and new releases about latest on-device AI capabilities.

## 🚀 Quick Start

**Choose your platform:**

<details>
<summary><b> CLI</b></summary>

### Download

| Windows                                                                                                    | macOS                                                                                                   | Linux                                                                                        |
| ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| [arm64 (Snapdragon NPU)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_arm64.exe) | [arm64 (Apple Silicon)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_arm64.pkg) | [arm64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_arm64.sh) |
| [x64 (Intel/AMD NPU)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_x86_64.exe)   | [x64](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_x86_64.pkg)                  | [x64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_x86_64.sh)  |

### Run your first model

```bash
# Chat with Qwen3
nexa infer ggml-org/Qwen3-1.7B-GGUF

# Multimodal: drag images into the CLI
nexa infer NexaAI/Qwen3-VL-4B-Instruct-GGUF

# NPU (Windows arm64 with Snapdragon X Elite)
nexa infer NexaAI/OmniNeural-4B
```

[▶️ Watch Demo](https://youtube.com/TODO) ｜ [📖 Full CLI Reference](https://docs.nexa.ai/nexa-sdk-go/NexaCLI)

</details>

<details>
<summary><b> Python SDK </b></summary>

```bash
pip install nexaai
```

```python
from nexaai import NexaLLM

model = NexaLLM("ggml-org/Qwen3-1.7B-GGUF")
response = model.generate("Hello, tell me a joke")
print(response)
```

[▶️ Watch Demo](https://youtube.com/TODO) ｜ [📖 Full Docs](https://docs.nexa.ai/en/nexa-sdk-python/overview)

</details>

<details>
<summary><b> Android SDK </b></summary>

Add to your `build.gradle`:

```gradle
implementation 'ai.nexa:nexasdk:latest'
```

```kotlin
val llm = NexaLLM.Builder()
    .model("NexaAI/Qwen3-4B-Instruct")
    .backend(Backend.NPU)
    .build()

llm.generate("Hello!") { token -> print(token) }
```

[▶️ Watch Demo](https://youtube.com/TODO) ｜ [📖 Full Docs](https://docs.nexa.ai/en/nexa-sdk-android/overview)

</details>

<details>
<summary><b> iOS SDK </b></summary>

Download [NexaSdk.xcframework](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/ios/latest/NexaSdk.xcframework.zip) and add to your Xcode project.

```swift
import NexaSdk

let llm = try LLM()
try await llm.load(from: modelURL)

let stream = await llm.generateAsyncStream(prompt: "Hello!")
for try await token in stream {
    print(token, terminator: "")
}
```

[▶️ Watch Demo](https://youtube.com/TODO) ｜ [📖 Full Docs](https://docs.nexa.ai/en/nexa-sdk-ios/overview)

</details>

<details>
<summary><b> Linux (via Docker) </b></summary>

```bash
docker pull nexaai/nexa-sdk:latest
docker run -it nexaai/nexa-sdk nexa infer ggml-org/Qwen3-1.7B-GGUF
```

[▶️ Watch Demo](https://youtube.com/TODO) ｜ [📖 Full Docs](https://docs.nexa.ai/en/nexa-sdk-docker/overview)

</details>



## 🏆 Recent Wins

- **Qualcomm** — Featured as ["revolutionizing on-device AI"](https://www.qualcomm.com/developer/blog/2025/09/omnineural-4b-nexaml-qualcomm-hexagon-npu). Linux SDK released in partnership with Qualcomm & Docker.
- **Qwen** — [Day-0 Qwen3-VL support](https://x.com/Alibaba_Qwen/status/1978154384098754943) in GGUF — we're the only framework that supports it.
- **IBM** — [Day-0 Granite 4.0 support](https://x.com/IBM/status/1978154384098754943) — featured alongside vLLM, llama.cpp, MLX.
- **Google** — [Featured for EmbeddingGemma NPU support](https://x.com/googleaidevs/status/1969188152049889511). First Gemma-3n multimodal inference in GGUF.
- **AMD** — [SDXL-turbo image generation on AMD NPU](https://www.amd.com/en/developer/resources/technical-articles/2025/advancing-ai-with-nexa-ai--image-generation-on-amd-npu-with-sdxl.html).



## ⚙️ Features & Comparisons

<div align="center">

| Features                                    | **NexaSDK**                                                | **Ollama** | **llama.cpp** | **LM Studio** |
| ------------------------------------------- | ---------------------------------------------------------- | ---------- | ------------- | ------------- |
| NPU support                                 | ✅ NPU-first                                               | ❌          | ❌             | ❌            |
| Android/iOS SDK support                     | ✅ NPU/GPU/CPU support                                     | ⚠️         | ⚠️            | ❌            |
| Linux support (Docker image)                | ✅                                                         | ✅         | ✅            | ❌            |
| Support any model in GGUF, MLX, NEXA format | ✅ Low-level Control                                       | ❌         | ⚠️            | ❌            |
| Full multimodality support                  | ✅ Image, Audio, Text, Embedding, Rerank, ASR, TTS         | ⚠️         | ⚠️            | ⚠️            |
| Cross-platform support                      | ✅ Desktop, Mobile (Android, iOS), Automotive, IoT (Linux) | ⚠️         | ⚠️            | ⚠️            |
| One line of code to run                     | ✅                                                         | ✅         | ⚠️            | ✅            |
| OpenAI-compatible API + Function calling    | ✅                                                         | ✅         | ✅            | ✅            |

<p align="center" style="margin-top:14px">
  <i>
      <b>Legend:</b>
      <span title="Full support">✅ Supported</span> &nbsp; | &nbsp;
      <span title="Partial or limited support">⚠️ Partial or limited support </span> &nbsp; | &nbsp;
      <span title="Not Supported">❌ No</span>
  </i>
</p>
</div>



## 📊 Benchmarks

<!-- TODO: Add platform-specific benchmarks here -->
<!--
### Qualcomm NPU (Snapdragon X Elite)
| Model | Tokens/sec | TTFT |
|-------|------------|------|

### Apple Neural Engine
| Model | Tokens/sec | TTFT |
|-------|------------|------|

### GPU (Metal / CUDA)
| Model | Tokens/sec | TTFT |
|-------|------------|------|
-->

Coming soon. See individual platform docs for current benchmarks.



## 📖 CLI Reference

| Essential Command                   | What it does                             |
| ----------------------------------- | ---------------------------------------- |
| `nexa -h`                           | show all CLI commands                    |
| `nexa pull <repo>`                  | Interactive download & cache of a model  |
| `nexa infer <repo>`                 | Local inference                          |
| `nexa list`                         | Show all cached models with sizes        |
| `nexa remove <repo>` / `nexa clean` | Delete one / all cached models           |
| `nexa serve --host 127.0.0.1:8080`  | Launch OpenAI‑compatible REST server     |
| `nexa run <repo>`                   | Chat with a model via an existing server |

👉 To interact with multimodal models, you can drag photos or audio clips directly into the CLI — you can even drop multiple images at once!

See [CLI Reference](https://docs.nexa.ai/nexa-sdk-go/NexaCLI) for full commands.

### Import model from local filesystem

```bash
# hf download <model> --local-dir /path/to/modeldir
nexa pull <model> --model-hub localfs --local-path /path/to/modeldir
```



## 🎯 You Decide What Model We Support Next

**[Nexa Wishlist](https://sdk.nexa.ai/wishlist)** — Request and vote for the models you want to run on-device.

Drop a Hugging Face repo ID, pick your preferred backend (GGUF, MLX, or Nexa format for Qualcomm + Apple NPUs), and watch the community's top requests go live in NexaSDK.

👉 **[Vote now at sdk.nexa.ai/wishlist](https://sdk.nexa.ai/wishlist)**



## 💰 Join Builder Bounty Program

Earn up to 1,500 USD for building with NexaSDK.

![Developer Bounty](assets/developer_bounty.png)

Learn more in our [Participant Details](https://docs.nexa.ai/community/builder-bounty).



## 🙏 Acknowledgements

We would like to thank the following projects:

- [ggml](https://github.com/ggml-org/ggml)
- [mlx-lm](https://github.com/ml-explore/mlx-lm)
- [mlx-vlm](https://github.com/Blaizzy/mlx-vlm)
- [mlx-audio](https://github.com/Blaizzy/mlx-audio)



## 📄 License

NexaSDK uses a dual licensing model:

### CPU/GPU Components

Licensed under [Apache License 2.0](LICENSE).

### NPU Components

- **Personal Use**: Free license key available from [Nexa AI Model Hub](https://sdk.nexa.ai/model). Each key activates 1 device for NPU usage.
- **Commercial Use**: Contact [hello@nexa.ai](mailto:hello@nexa.ai) for licensing.
