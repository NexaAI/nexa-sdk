# Nexa Android SDK 演示 App

[![Product Hunt](https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=1049998&theme=dark&period=daily&t=1765991451976)](https://www.producthunt.com/products/nexasdk-for-mobile)

> 📣 **NexaSDK for Android** 被 [Qualcomm 博客](https://www.qualcomm.com/developer/blog/2025/11/nexa-ai-for-android-simple-way-to-bring-on-device-ai-to-smartphones-with-snapdragon) 评价为"将端侧 AI 引入 Snapdragon 智能手机的简易方案"

## 概述

Nexa AI Android SDK 让 Android 应用可以在设备端使用 NPU 加速进行 AI 推理。支持在 Android 设备上运行大型语言模型（LLMs）、视觉语言模型（VLMs）、嵌入模型、语音识别（ASR）、重排序和计算机视觉模型，并可在 NPU、GPU 或 CPU 上进行推理。

📖 完整文档请参考 [Android SDK 文档](https://docs.nexa.ai/cn/nexa-sdk-android/overview)。

## 设备兼容性

### 支持的硬件

| 组件 | 要求 |
|-----------|-------------|
| **NPU** | 高通骁龙 8 Gen 4（已优化） |
| **GPU** | 高通 Adreno GPU |
| **CPU** | ARM64-v8a |
| **RAM** | 推荐 4GB 及以上 |
| **存储** | 100MB - 4GB（具体取决于模型） |

### 最低要求

- Android API 等级 27 及以上（Android 8.1 Oreo）
- 架构：ARM64-v8a

## 快速开始（APK 安装）

### 标准演示 App

下载并安装预编译的 APK：

```bash
# 下载地址: https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/android-demo-release/nexaai-demo-app.apk
adb install nexaai-demo-app.apk
```

### GPT-OSS NPU 演示

在高通 NPU 上运行 GPT-OSS 模型：

```bash
# 下载地址: https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/nexa_sdk/huggingface-models/gpt-oss-android-demo/nexaai-gpt-oss-npu.apk
adb install nexaai-gpt-oss-npu.apk
```

## 从源码编译

> 🎬 观看 [教程视频](https://www.linkedin.com/feed/update/urn:li:activity:7394055404562935809)，只需 40 秒即可体验如何运行示例 App。

### 详细步骤

1. **克隆仓库**

   ```bash
   git clone https://github.com/NexaAI/nexa-sdk/
   ```

2. **用 Android Studio 打开**
   
   用 Android Studio 打开 `bindings/android` 文件夹。

3. **下载模型**
   
   按照 [Android SDK 文档](https://docs.nexa.ai/cn/nexa-sdk-android/overview) 的说明下载模型。以下是一些可下载的示例：
   - [Granite-4.0-h-350M-NPU](https://huggingface.co/NexaAI/Granite-4.0-h-350M-NPU-mobile)
   - [Granite-4-Micro-NPU](https://huggingface.co/NexaAI/Granite-4-Micro-NPU-mobile)
   - [OmniNeural-4B](https://huggingface.co/NexaAI/OmniNeural-4B-mobile)
   
   将模型放到 App 的数据目录：
   ```
   /data/data/com.nexa.demo/files/models/<model-name>
   ```

4. **编译并运行** 在 Android Studio 中编译并运行应用
