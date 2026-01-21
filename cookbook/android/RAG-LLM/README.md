# RAG-LLM Demo

## Overview
This project is a demo for building a Retrieval-Augmented Generation (RAG) pipeline using NexaSDK. It showcases how to combine embedding and LLM models to answer questions over your own documents, also shows the performance of Nexa AI's infra for NPU vs CPU/GPU, especcially for prefilling speed gains.

## APK download
- [nexa-android-rag.apk](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/android/20260121/nexa-android-rag.apk)

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

### Way 1: use apk
Download [nexa-android-rag.apk](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/android/20260121/nexa-android-rag.apk) and install it on your device.
```bash
adb install nexa-android-rag.apk
```

### way 2: build from source with Android Studio
1. Clone the repository

```bash
git clone https://github.com/NexaAI/nexa-sdk/
```

2. Open this folder `cookbook/android/RAG-LLM` in Android Studio

3. Build and run the app