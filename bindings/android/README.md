# Nexa Android SDK Demo App

## Overview

The Nexa AI Android SDK enables on-device AI inference for Android applications with NPU acceleration. Run Large Language Models (LLMs), Vision-Language Models (VLMs), Embeddings, Speech Recognition (ASR), Reranking, and Computer Vision models on Android devices with support for NPU, GPU, and CPU inference. Follow [Android SDK Doc](https://docs.nexa.ai/nexa-sdk-android/overview) for more details. This folder contains the demo app for the Android SDK.

## Device Compatibility

### Supported Hardware

- **NPU**: Qualcomm Snapdragon 8 Gen 4 (optimized)
- **GPU**: Qualcomm Adreno GPU
- **CPU**: ARM64-v8a
- **RAM**: 4GB+ recommended
- **Storage**: 100MB - 4GB (varies by model)

### Minimum Requirements

- Android API Level 27+ (Android 8.1 Oreo)
- **Architecture**: ARM64-v8a
- **Android SDK Version**: 27+

## Build and Run

Here is a [tutorial video demo](https://www.linkedin.com/feed/update/urn:li:activity:7394055404562935809) showing how to run the sample App in 40 seconds. Or, you can follow the instructions below step by step.

1. Clone the repository

```bash
git clone https://github.com/NexaAI/nexa-sdk/
```

2. Open this folder `bindings/android` in Android Studio

3. Follow the instructions in [Android SDK Doc](https://docs.nexa.ai/nexa-sdk-android/overview), download model. For example, [Granite-4.0-h-350M-NPU](https://huggingface.co/NexaAI/Granite-4.0-h-350M-NPU-mobile) or [Granite-4-Micro-NPU](https://huggingface.co/NexaAI/Granite-4-Micro-NPU-mobile), or [OmniNeural-4B](https://huggingface.co/NexaAI/OmniNeural-4B-mobile) and put it in App's assets folder (For example, `/data/data/com.nexa.demo/files/models/Granite-4.0-h-350M-NPU`).

4. Build and run the app
