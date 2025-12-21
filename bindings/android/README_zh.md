# Nexa Android SDK 演示 App

[![Product Hunt](https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=1049998&theme=dark&period=daily&t=1765991451976)](https://www.producthunt.com/products/nexasdk-for-mobile)

> 📣 **NexaSDK for Android** 被 [Qualcomm 博客](https://www.qualcomm.com/developer/blog/2025/11/nexa-ai-for-android-simple-way-to-bring-on-device-ai-to-smartphones-with-snapdragon) 评价为"将端侧 AI 引入 Snapdragon 智能手机的简易方案"

## 概述

Nexa AI Android SDK 让 Android 应用可以在设备端使用 NPU 加速进行 AI 推理。支持在 Android 设备上运行大型语言模型（LLMs）、视觉语言模型（VLMs）、嵌入模型、语音识别（ASR）、重排序和计算机视觉模型，并可在 NPU、GPU 或 CPU 上进行推理。更多详情请参考 [Android SDK 文档](https://docs.nexa.ai/nexa-sdk-android/overview)。本文件夹包含 Android SDK 的演示应用。

## 设备兼容性

### 支持的硬件

- **NPU**：高通骁龙 8 Gen 4（已优化）
- **GPU**：高通 Adreno GPU
- **CPU**：ARM64-v8a
- **RAM**：推荐 4GB 及以上
- **存储**：100MB - 4GB（具体取决于模型）

### 最低要求

- Android API 等级 27 及以上（Android 8.1 Oreo）
- **架构**：ARM64-v8a
- **Android SDK 版本**：27 及以上

## 编译及运行

这里有一个 [教程视频演示](https://www.linkedin.com/feed/update/urn:li:activity:7394055404562935809)，只需 40 秒即可体验如何运行示例 App。也可以按照以下步骤操作：

1. 克隆仓库

```bash
git clone https://github.com/NexaAI/nexa-sdk/
```

2. 用 Android Studio 打开本文件夹 `bindings/android`

3. 按照 [Android SDK 文档](https://docs.nexa.ai/nexa-sdk-android/overview)的说明下载模型。例如，[Granite-4.0-h-350M-NPU](https://huggingface.co/NexaAI/Granite-4.0-h-350M-NPU-mobile) 或 [Granite-4-Micro-NPU](https://huggingface.co/NexaAI/Granite-4-Micro-NPU-mobile)，或 [OmniNeural-4B](https://huggingface.co/NexaAI/OmniNeural-4B-mobile)，下载后放到 App 的 assets 目录下（比如 `/data/data/com.nexa.demo/files/models/Granite-4.0-h-350M-NPU`）。

4. 编译并运行应用
