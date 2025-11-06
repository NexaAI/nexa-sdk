# Nexa AI Android SDK Documentation

## Overview

The Nexa AI Android SDK enables on-device AI inference for Android applications with NPU acceleration. Run Large Language Models (LLMs), Vision-Language Models (VLMs), Embeddings, Speech Recognition (ASR), Reranking, and Computer Vision models on Android devices with support for NPU, GPU, and CPU inference. Follow [Android SDK Doc](https://docs.nexa.ai/nexa-sdk-android/overview) for more details.

### Key Features

- **Multiple Model Types**: LLM, VLM, Embeddings, ASR, Reranker, and Computer Vision
- **NPU Acceleration**: Optimized for Qualcomm Hexagon NPU (Snapdragon 8 Gen 4+)
- **Easy Integration**: Simple Kotlin API with builder pattern
- **On-Device Privacy**: All inference happens locally

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

## Quick Installation

Nexa AI SDK is available from Maven Central, add below to your `app/build.gradle.kts`:

```kotlin
dependencies {
    implementation("ai.nexa:core:0.0.9")
}
```

---

For more API usage examples, please refer to our documentation: [nexa-sdk-android-docs](https://docs.nexa.ai/nexa-sdk-android/APIReference)

## Model Download

You can download the required model from our official website: [sdk.nexa.ai](https://sdk.nexa.ai/model) or from [Hugging Face](https://huggingface.co/NexaAI).

### Adding a New Model

To add a new model for testing or use, follow these steps:

1. Locate the `model_list.json` file in the `assets` folder of the demo.
2. Add the configuration for the new model you want to test or use, including the download URL and other relevant settings.
3. Save the file and restart the demo to make the new model available.

If you want to use a local model directly, simply specify the path to the model.
