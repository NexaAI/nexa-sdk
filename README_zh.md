<div align="center" style="text-decoration: none;">
	<img width="100%" src="assets/banner1.png" alt="Nexa AI Banner">
	<p style="font-size: 1.3em; font-weight: 600; margin-bottom: 20px;"> 
		<a href="README_zh.md"> 简体中文 </a>
		|
		<a href="README.md"> English </a>
	</p>
	<p style="font-size: 1.3em; font-weight: 600; margin-bottom: 20px;">🤝 支持的芯片厂商 </p>
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

**NexaSDK 让你用极低能耗打造最快、最聪明的本地 AI。** 它是一套高性能本地推理框架，只需几行代码即可在 Android、Windows、Linux、macOS 与 iOS 的 NPU、GPU、CPU 上运行最新的多模态 AI 模型。

NexaSDK 往往能比其他人提前数周甚至数月支持最新模型 —— Qwen3-VL、DeepSeek-OCR、Gemma3n（视觉版）等。

> ⭐ **star本仓库**，及时获取最新的本地 AI 能力更新与发布。

## 🏆 重要里程碑

- **Qualcomm** 官方博客中 3 次重点介绍我们。
	- [在 Qualcomm Hexagon NPU 上创新多模态 AI](https://www.qualcomm.com/developer/blog/2025/09/omnineural-4b-nexaml-qualcomm-hexagon-npu)。
	- [Qualcomm Hexagon NPU 在计算与移动平台（汽车与 IoT）上的首个 Day-0 模型支持](https://www.qualcomm.com/developer/blog/2025/10/granite-4-0-to-the-edge-on-device-ai-for-real-world-performance)。
	- [在 Snapdragon 手机上体验端侧 AI 的简单方式](https://www.qualcomm.com/developer/blog/2025/11/nexa-ai-for-android-simple-way-to-bring-on-device-ai-to-smartphones-with-snapdragon)
- **Qwen** 为我们在 NPU、GPU、CPU 上 [Day-0 支持 Qwen3-VL](https://x.com/Alibaba_Qwen/status/1978154384098754943) 做了官方推荐。我们在 GGUF 支持上领先 Ollama 和 llama.cpp 3 周，并且目前只有我们支持 NPU。
- **IBM** 在[官方博客](https://www.ibm.com/new/announcements/ibm-granite-4-0-hyper-efficient-high-performance-hybrid-models)中，将 NexaML 推理引擎与 vLLM、llama.cpp、MLX 并列介绍，感谢我们 Day-0 支持 Granite 4.0。
- **Google** 认可我们对 EmbeddingGemma Day-0 的 NPU 支持（[官方致谢](https://x.com/googleaidevs/status/1969188152049889511)）。
- **AMD** 在[官方博客](https://www.amd.com/en/developer/resources/technical-articles/2025/advancing-ai-with-nexa-ai--image-generation-on-amd-npu-with-sdxl.html)报道我们在 AMD NPU 上实现 SDXL-turbo 图像生成。
- **NVIDIA** 在[官方博客](https://blogs.nvidia.com/blog/rtx-ai-garage-nexa-hyperlink-local-agent/)中介绍了由 NexaSDK 支撑的火爆本地 AI 应用 Hyperlink。
- **Microsoft** 在 Microsoft Ignite 2025 上台展示了我们作为[官方合作伙伴](https://www.linkedin.com/posts/mixen_excited-to-celebrate-our-developer-partnerships-activity-7396601602327007232-AmCR?utm_source=share&utm_medium=member_desktop&rcm=ACoAAChXnS8B4gqbBLUlWfwt-ck0XAv472NzT4k)。
- **Intel** 在[官方帖子](https://www.linkedin.com/posts/intel-software_ai-ondeviceai-nexasdk-activity-7376337062087667712-xw7i?utm_source=share&utm_medium=member_desktop&rcm=ACoAAChXnS8B4gqbBLUlWfwt-ck0XAv472NzT4k)中提到我们对 Intel NPU 的支持。

## 🚀 快速开始

| 平台            | 链接                                                                                           |
| --------------- | ---------------------------------------------------------------------------------------------- |
| 🖥️ CLI          | [快速开始](#-cli) ｜ [文档](https://docs.nexa.ai/en/nexa-sdk-go/NexaCLI)                        |
| 🐍 Python       | [快速开始](#-python-sdk) ｜ [文档](https://docs.nexa.ai/en/nexa-sdk-python/overview)            |
| 🤖 Android      | [快速开始](#-android-sdk) ｜ [文档](https://docs.nexa.ai/en/nexa-sdk-android/overview)          |
| 🐳 Linux Docker | [快速开始](#-linux-docker) ｜ [文档](https://docs.nexa.ai/en/nexa-sdk-docker/overview)          |
| 🍎 iOS          | [快速开始](#-ios-sdk) ｜ [文档](https://docs.nexa.ai/en/nexa-sdk-ios/overview)                  |

---

### 🖥️ CLI

**下载：**

| Windows                                                                                                  | macOS                                                                                                   | Linux                                                                                        |
| -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| [arm64 (Qualcomm NPU)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_arm64.exe) | [arm64 (Apple Silicon)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_arm64.pkg) | [arm64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_arm64.sh) |
| [x64 (Intel/AMD NPU)](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_windows_x86_64.exe) | [x64](https://public-storage.nexa4ai.com/nexa_sdk/downloads/nexa-cli_macos_x86_64.pkg)                  | [x64](https://github.com/NexaAI/nexa-sdk/releases/latest/download/nexa-cli_linux_x86_64.sh)  |

**运行你的第一个模型：**

```bash
# 与 Qwen3 对话
nexa infer ggml-org/Qwen3-1.7B-GGUF

# 多模态：在 CLI 中拖入图片
nexa infer NexaAI/Qwen3-VL-4B-Instruct-GGUF

# NPU（Windows arm64，Snapdragon X Elite）
nexa infer NexaAI/OmniNeural-4B
```

- **模型类型：** LLM、多模态、ASR、OCR、Rerank、目标检测、图像生成、Embedding
- **格式：** GGUF、MLX、NEXA
- **NPU 模型：** [Model Hub](https://sdk.nexa.ai/model)
- 📖 [CLI 参考文档](https://docs.nexa.ai/en/nexa-sdk-go/NexaCLI)

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

- **模型类型：** LLM、多模态、ASR、OCR、Rerank、目标检测、图像生成、Embedding
- **格式：** GGUF、MLX、NEXA
- **NPU 模型：** [Model Hub](https://sdk.nexa.ai/model)
- 📖 [Python SDK 文档](https://docs.nexa.ai/en/nexa-sdk-python/quickstart)

---

### 🤖 Android SDK

在 `app/AndroidManifest.xml` 中添加：

```xml
<application android:extractNativeLibs="true">
```

在 `build.gradle.kts` 中添加：

```kotlin
dependencies {
		implementation("ai.nexa:core:0.0.15")
}
```

```kotlin
// 初始化 SDK
NexaSdk.getInstance().init(this)

// 加载并运行模型
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

- **要求：** Android minSdk 27，Qualcomm Snapdragon 8 Gen 4 芯片
- **模型类型：** LLM、多模态、ASR、OCR、Rerank、Embedding
- **NPU 模型：** [Supported Models](https://docs.nexa.ai/en/nexa-sdk-android/overview#supported-models)
- 📖 [Android SDK 文档](https://docs.nexa.ai/en/nexa-sdk-android/quickstart)

---

### 🐳 Linux Docker

```bash
docker pull nexa4ai/nexasdk:latest

export NEXA_TOKEN="your_token_here"
docker run --rm -it --privileged \
	-e NEXA_TOKEN \
	nexa4ai/nexasdk:latest infer NexaAI/Granite-4.0-h-350M-NPU
```

- **要求：** Qualcomm Dragonwing IQ9，ARM64 系统
- **模型类型：** LLM、VLM、ASR、CV、Rerank、Embedding
- **NPU 模型：** [Supported Models](https://docs.nexa.ai/en/nexa-sdk-docker/overview#supported-models)
- 📖 [Linux Docker 文档](https://docs.nexa.ai/en/nexa-sdk-docker/quickstart)

---

### 🍎 iOS SDK

下载 [NexaSdk.xcframework](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/ios/latest/NexaSdk.xcframework.zip) 并添加到 Xcode 项目。

```swift
import NexaSdk

// 示例：语音识别
let asr = try Asr(plugin: .ane)
try await asr.load(from: modelURL)

let result = try await asr.transcribe(options: .init(audioPath: "audio.wav"))
print(result.asrResult.transcript)
```

- **要求：** iOS 17.0+ / macOS 15.0+，Swift 5.9+
- **模型类型：** LLM、ASR、OCR、Rerank、Embedding
- **ANE 模型：** [Apple Neural Engine Models](https://huggingface.co/collections/NexaAI/apple-neural-engine)
- 📖 [iOS SDK 文档](https://docs.nexa.ai/en/nexa-sdk-ios/quickstart)

## ⚙️ 功能与对比

<div align="center">

| 功能                                   | **NexaSDK**                                                | **Ollama** | **llama.cpp** | **LM Studio** |
| -------------------------------------- | ---------------------------------------------------------- | ---------- | ------------- | ------------- |
| NPU 支持                               | ✅ NPU 优先                                                | ❌         | ❌            | ❌            |
| Android/iOS SDK 支持                   | ✅ NPU/GPU/CPU 支持                                        | ⚠️         | ⚠️            | ❌            |
| Linux 支持（Docker 镜像）              | ✅                                                         | ✅         | ✅            | ❌            |
| Day-0 支持 GGUF、MLX、NEXA             | ✅                                                         | ❌         | ⚠️            | ❌            |
| 完整多模态支持                         | ✅ 图像、音频、文本、Embedding、Rerank、ASR、TTS           | ⚠️         | ⚠️            | ⚠️            |
| 跨平台支持                             | ✅ 桌面、移动（Android、iOS）、车载、IoT（Linux）          | ⚠️         | ⚠️            | ⚠️            |
| 一行代码即可运行                       | ✅                                                         | ✅         | ⚠️            | ✅            |
| OpenAI 兼容 API + Function calling     | ✅                                                         | ✅         | ✅            | ✅            |

<p align="center" style="margin-top:14px">
	<i>
			<b>图例：</b>
			<span title="Full support">✅ Supported</span> &nbsp; | &nbsp;
			<span title="Partial or limited support">⚠️ Partial or limited support </span> &nbsp; | &nbsp;
			<span title="Not Supported">❌ No</span>
	</i>
</p>
</div>

## 🙏 致谢

我们感谢以下项目：

- [ggml](https://github.com/ggml-org/ggml)
- [mlx-lm](https://github.com/ml-explore/mlx-lm)
- [mlx-vlm](https://github.com/Blaizzy/mlx-vlm)
- [mlx-audio](https://github.com/Blaizzy/mlx-audio)

## 📄 许可证

NexaSDK 采用双重许可模式：

### CPU/GPU 组件

基于 [Apache License 2.0](LICENSE)。

### NPU 组件

- **个人使用**：可从 [Nexa AI Model Hub](https://sdk.nexa.ai/model) 免费获取许可密钥。每个密钥激活 1 台设备的 NPU 使用。
- **商业使用**：联系 [hello@nexa.ai](mailto:hello@nexa.ai) 获取授权。

## 🤝 联系与社区支持

### 商务合作

如需模型发布合作、商务洽谈或其他问题，请在[此处](https://nexa.ai/book-a-call)安排会议。

### 社区与支持

想要更多模型支持、后端支持、设备支持或新功能？我们很乐意听到你的声音！

欢迎在 GitHub [提交 issue](https://github.com/NexaAI/nexa-sdk/issues)，提出你的需求、建议或反馈。你的意见帮助我们确定优先级。

加入社区：

- [Discord](https://discord.gg/thRu2HaK4D)
- [Slack](https://join.slack.com/t/nexaai/shared_invite/zt-30a8yfv8k-1JqAXv~OjKJKLqvbKqHJxA)
- **[Nexa Wishlist](https://sdk.nexa.ai/wishlist)** —— 提交并为你想要运行在本地的模型投票。
