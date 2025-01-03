# Nexa SDK Quick Start Guide

## Prerequisites
- Android Studio
- Android device or emulator
- Git

## Installation Steps

### 1. Clone Repository
```bash
git clone https://github.com/NexaAI/nexa-sdk.git
cd nexa-sdk
```

### 2. Set Up Development Branch
```bash
git checkout -b android-omnivlm origin/demo/android-omnivlm
```

### 3. Project Setup
1. Open the `android` folder in Android Studio
2. Wait for project sync and Gradle build to complete

### 4. Model Setup
1. Download the OmniVLM model from [Nexa AI ModelHub](https://nexa.ai/NexaAI/omniVLM/gguf-q8_0/file)

2. Copy the model files (`model-q8_0.gguf`  and `projector-q8_0.gguf`) to your device storage:
   ```
   Device path: /sdcard/Android/data/ai.nexa.app_java/files
   ```

### 5. Run the Application
1. Connect your Android device or start an emulator
2. Click the \"Run\" button in Android Studio to build and launch the app
