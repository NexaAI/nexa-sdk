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

**NexaSDK lets you build the smartest and fastest on-device AI with minimum energy.** It is a highly performant local inference framework that runs the latest multimodal AI models locally on NPU, GPU, and CPU - across Android, Windows, Linux, macOS, and iOS devices with a few lines of code.

NexaSDK supports latest models **weeks or months before anyone else** — Qwen3-VL, DeepSeek-OCR, Gemma3n (Vision), and more.

> ⭐ **Star this repo** to keep up with exciting updates and new releases about latest on-device AI capabilities.

## 🏆 Recognized Milestones

- **Qualcomm** featured us **3 times** in official blogs.
  - [Innovating Multimodal AI on Qualcomm Hexagon NPU](https://www.qualcomm.com/developer/blog/2025/09/omnineural-4b-nexaml-qualcomm-hexagon-npu).
  - [First-ever Day-0 model support on Qualcomm Hexagon NPU for compute and mobile platforms, Auto and IoT](https://www.qualcomm.com/developer/blog/2025/10/granite-4-0-to-the-edge-on-device-ai-for-real-world-performance).
  - [A simple way to bring on-device AI to smartphones with Snapdragon](https://www.qualcomm.com/developer/blog/2025/11/nexa-ai-for-android-simple-way-to-bring-on-device-ai-to-smartphones-with-snapdragon)
- **Qwen** featured us for [Day-0 Qwen3-VL support on NPU, GPU, and CPU](https://x.com/Alibaba_Qwen/status/1978154384098754943). We were 3 weeks ahead of Ollama and llama.cpp on GGUF support, and no one else supports it on NPU to date.
- **IBM** featured our NexaML inference engine alongside vLLM, llama.cpp, and MLX in [official IBM blog](https://www.ibm.com/new/announcements/ibm-granite-4-0-hyper-efficient-high-performance-hybrid-models) and also for Day-0 Granite 4.0 support.
- **Google** featured us for [EmbeddingGemma Day-0 NPU support](https://x.com/googleaidevs/status/1969188152049889511).
- **AMD** featured us for [enabling SDXL-turbo image generation on AMD NPU](https://www.amd.com/en/developer/resources/technical-articles/2025/advancing-ai-with-nexa-ai--image-generation-on-amd-npu-with-sdxl.html).
- **NVIDIA** featured Hyperlink, a viral local AI app powered by NexaSDK, in their [official blog](https://blogs.nvidia.com/blog/rtx-ai-garage-nexa-hyperlink-local-agent/).
- **Microsoft** presented us on stage at Microsoft Ignite 2025 as [official partner](https://www.linkedin.com/posts/mixen_excited-to-celebrate-our-developer-partnerships-activity-7396601602327007232-AmCR?utm_source=share&utm_medium=member_desktop&rcm=ACoAAChXnS8B4gqbBLUlWfwt-ck0XAv472NzT4k).
- **Intel** featured us for [Intel NPU support in NexaSDK](https://www.linkedin.com/posts/intel-software_ai-ondeviceai-nexasdk-activity-7376337062087667712-xw7i?utm_source=share&utm_medium=member_desktop&rcm=ACoAAChXnS8B4gqbBLUlWfwt-ck0XAv472NzT4k).

## 🚀 Quick Start

| Platform        | Links                                                                                     |
| --------------- | ----------------------------------------------------------------------------------------- |
| 🖥️ CLI          | [Quick Start](#-cli) ｜ [Docs](https://docs.nexa.ai/en/nexa-sdk-go/NexaCLI)               |
| 🐍 Python       | [Quick Start](#-python-sdk) ｜ [Docs](https://docs.nexa.ai/en/nexa-sdk-python/overview)   |
| 🤖 Android      | [Quick Start](#-android-sdk) ｜ [Docs](https://docs.nexa.ai/en/nexa-sdk-android/overview) |
| 🐳 Linux Docker | [Quick Start](#-linux-docker) ｜ [Docs](https://docs.nexa.ai/en/nexa-sdk-docker/overview) |
| 🍎 iOS          | [Quick Start](#-ios-sdk) ｜ [Docs](https://docs.nexa.ai/en/nexa-sdk-ios/overview)         |

---

### 🖥️ CLI

**Download:**

| Windows                                                                                                  | macOS                                                                                                   | Linux                                                                                        |
| -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| [arm64 (Qualcomm NPU)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_arm64.exe) | [arm64 (Apple Silicon)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_arm64.pkg) | [arm64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_arm64.sh) |
| [x64 (Intel/AMD NPU)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_x86_64.exe) | [x64](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_x86_64.pkg)                  | [x64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_x86_64.sh)  |

**Run your first model:**

```bash
# Chat with Qwen3
nexa infer ggml-org/Qwen3-1.7B-GGUF

# Multimodal: drag images into the CLI
nexa infer NexaAI/Qwen3-VL-4B-Instruct-GGUF

# NPU (Windows arm64 with Snapdragon X Elite)
nexa infer NexaAI/OmniNeural-4B
```

- **Models:** LLM, Multimodal, ASR, OCR, Rerank, Object Detection, Image Generation, Embedding
- **Formats:** GGUF, MLX, NEXA
- **NPU Models:** [Model Hub](https://sdk.nexa.ai/model)
- 📖 [CLI Reference Docs](https://docs.nexa.ai/en/nexa-sdk-go/NexaCLI)

---

### 🐍 Python SDK

```bash
pip install nexaai
```

```python
from nexaai import LLM, GenerationConfig, ModelConfig, LlmChatMessage

llm = LLM.from_(model="NexaAI/Qwen3-0.6B-GGUF", config=ModelConfig())

conversation = [
    LlmChatMessage(role="user", content="Hello, tell me a joke")
]
prompt = llm.apply_chat_template(conversation)
for token in llm.generate_stream(prompt, GenerationConfig(max_tokens=100)):
    print(token, end="", flush=True)
```

- **Models:** LLM, Multimodal, ASR, OCR, Rerank, Object Detection, Image Generation, Embedding
- **Formats:** GGUF, MLX, NEXA
- **NPU Models:** [Model Hub](https://sdk.nexa.ai/model)
- 📖 [Python SDK Docs](https://docs.nexa.ai/en/nexa-sdk-python/quickstart)

---

### 🤖 Android SDK

Add to your `app/AndroidManifest.xml`

```xml
<application android:extractNativeLibs="true">
```

Add to your `build.gradle.kts`:

```kotlin
dependencies {
    implementation("ai.nexa:core:0.0.16")
}
```

```kotlin
// Initialize SDK
NexaSdk.getInstance().init(this)

// Load and run model
VlmWrapper.builder()
    .vlmCreateInput(VlmCreateInput(
        model_name = "omni-neural",
        model_path = "/data/data/your.app/files/models/OmniNeural-4B/files-1-1.nexa",
        plugin_id = "npu",
        config = ModelConfig()
    ))
    .build()
    .onSuccess { vlm ->
        vlm.generateStreamFlow("Hello!", GenerationConfig()).collect { print(it) }
    }
```

- **Requirements:** Android minSdk 27, Qualcomm Snapdragon 8 Gen 4 Chip
- **Models:** LLM, Multimodal, ASR, OCR, Rerank, Embedding
- **NPU Models:** [Supported Models](https://docs.nexa.ai/en/nexa-sdk-android/overview#supported-models)
- 📖 [Android SDK Docs](https://docs.nexa.ai/en/nexa-sdk-android/quickstart)

---

### 🐳 Linux Docker

```bash
docker pull nexa4ai/nexasdk:latest

export NEXA_TOKEN="your_token_here"
docker run --rm -it --privileged \
  -e NEXA_TOKEN \
  nexa4ai/nexasdk:latest infer NexaAI/Granite-4.0-h-350M-NPU
```

- **Requirements:** Qualcomm Dragonwing IQ9, ARM64 systems
- **Models:** LLM, VLM, ASR, CV, Rerank, Embedding
- **NPU Models:** [Supported Models](https://docs.nexa.ai/en/nexa-sdk-docker/overview#supported-models)
- 📖 [Linux Docker Docs](https://docs.nexa.ai/en/nexa-sdk-docker/quickstart)

---

### 🍎 iOS SDK

Download [NexaSdk.xcframework](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/ios/latest/NexaSdk.xcframework.zip) and add to your Xcode project.

```swift
import NexaSdk

// Example: Speech Recognition
let asr = try Asr(plugin: .ane)
try await asr.load(from: modelURL)

let result = try await asr.transcribe(options: .init(audioPath: "audio.wav"))
print(result.asrResult.transcript)
```

- **Requirements:** iOS 17.0+ / macOS 15.0+, Swift 5.9+
- **Models:** LLM, ASR, OCR, Rerank, Embedding
- **ANE Models:** [Apple Neural Engine Models](https://huggingface.co/collections/NexaAI/apple-neural-engine)
- 📖 [iOS SDK Docs](https://docs.nexa.ai/en/nexa-sdk-ios/quickstart)

## ⚙️ Features & Comparisons

<div align="center">

| Features                                 | **NexaSDK**                                                | **Ollama** | **llama.cpp** | **LM Studio** |
| ---------------------------------------- | ---------------------------------------------------------- | ---------- | ------------- | ------------- |
| NPU support                              | ✅ NPU-first                                               | ❌         | ❌            | ❌            |
| Android/iOS SDK support                  | ✅ NPU/GPU/CPU support                                     | ⚠️         | ⚠️            | ❌            |
| Linux support (Docker image)             | ✅                                                         | ✅         | ✅            | ❌            |
| Day-0 model support in GGUF, MLX, NEXA   | ✅                                                         | ❌         | ⚠️            | ❌            |
| Full multimodality support               | ✅ Image, Audio, Text, Embedding, Rerank, ASR, TTS         | ⚠️         | ⚠️            | ⚠️            |
| Cross-platform support                   | ✅ Desktop, Mobile (Android, iOS), Automotive, IoT (Linux) | ⚠️         | ⚠️            | ⚠️            |
| One line of code to run                  | ✅                                                         | ✅         | ⚠️            | ✅            |
| OpenAI-compatible API + Function calling | ✅                                                         | ✅         | ✅            | ✅            |

<p align="center" style="margin-top:14px">
  <i>
      <b>Legend:</b>
      <span title="Full support">✅ Supported</span> &nbsp; | &nbsp;
      <span title="Partial or limited support">⚠️ Partial or limited support </span> &nbsp; | &nbsp;
      <span title="Not Supported">❌ No</span>
  </i>
</p>
</div>

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

## 🤝 Contact & Community Support

### Business Inquiries

For model launching partner, business inquiries, or any other questions, please schedule a call with us [here](https://nexa.ai/book-a-call).

### Community & Support

Want more model support, backend support, device support or other features? We'd love to hear from you!

Feel free to [submit an issue](https://github.com/NexaAI/nexa-sdk/issues) on our GitHub repository with your requests, suggestions, or feedback. Your input helps us prioritize what to build next.

Join our community:

- [Discord](https://discord.gg/thRu2HaK4D)
- [Slack](https://join.slack.com/t/nexaai/shared_invite/zt-30a8yfv8k-1JqAXv~OjKJKLqvbKqHJxA)
- **[Nexa Wishlist](https://sdk.nexa.ai/wishlist)** — Request and vote for the models you want to run on-device.

## 🏆 Nexa × Qualcomm On-Device Bounty Program

Round 1: Build a working Android AI app that runs fully on-device on Qualcomm Hexagon NPU with NexaSDK.

Timeline (PT): Jan 15 → Feb 15
Prizes: $6,500 cash prize, Qualcomm official spotlight, flagship Snapdragon device, expert mentorship, and more

👉 Join & details: [https://sdk.nexa.ai/bounty](https://sdk.nexa.ai/bounty)
