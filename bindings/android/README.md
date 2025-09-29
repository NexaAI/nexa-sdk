# Nexa AI Android Tutorial

## Overview

[`Nexa Demo`](https://github.com/NexaAI/nexa-sdk-examples/tree/main/android) is an Android tutorial project demonstrating how to use the Nexa AI SDK to run AI models on Android devices. The project includes model downloading, loading, text generation, and visual question answering functionality.

## Features

- Model download management
- Support for LLM and VLM models
- Real-time chat conversation
- Visual question answering
- Model parameter configuration
- Local file caching

## Supported Models

### LLM Models
- Qwen3-0.6B-Q8_0: Lightweight Chinese dialogue model
- Qwen3-1.8B-Q8_0: Medium-scale Chinese dialogue model

### VLM Models
- SmolVLM-256M-Instruct-Q8_0: Lightweight vision-language model
- SmolVLM-256M-Instruct-f16: High-quality vision-language model

## Quick Start

### 1. Requirements

- Android Studio Arctic Fox or higher
- Android SDK 27 or higher
- Kotlin 1.9.23 or higher

### 2. Project Setup

Configure Maven Central repository in `settings.gradle.kts`:

```kotlin
dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
    }
}
```

### 3. Dependencies

The project uses Nexa AI SDK from Maven Central in `app/build.gradle.kts`:

```kotlin
android {
	// ...
	packagingOptions {
        jniLibs.useLegacyPackaging = true
    }
}

dependencies {
    implementation("ai.nexa:core:0.0.3")
    // Other dependencies...
}
```

### 4. Running the Project

1. Clone or download the project
2. Open the project in Android Studio
3. Connect an Android device or start an emulator
4. Click the run button

## Usage Guide

### Model Download

1. Launch the app and select a model to download
2. Click the "Download" button to start downloading
3. The model will be automatically loaded after download completion

### LLM Chat

1. Ensure an LLM model is loaded
2. Enter your question in the input field
3. Click "Send" to start the conversation
4. The model will stream responses in real-time

### VLM Visual Q&A

1. Ensure a VLM model is loaded
2. Enter a question in the input field
3. Click "Send" to get the response
4. The model will process the text input

## Project Structure

```
nexa-sdk-examples/
├── android/
│   ├── app/
│   │   ├── src/main/
│   │   │   ├── java/com/nexa/demo/
│   │   │   │   ├── MainActivity.kt          # Main activity
│   │   │   │   ├── bean/
│   │   │   │   │   ├── ModelData.kt         # Model data structure
│   │   │   │   │   └── DownloadableFile.kt  # Downloadable file structure
│   │   │   │   ├── FileConfig.kt            # File configuration
│   │   │   │   └── GenerationConfigSample.kt # Generation config sample
│   │   │   ├── assets/
│   │   │   │   ├── model_list.json          # Model list configuration
│   │   │   │   ├── model_list_local.json    # Local model configuration
│   │   │   │   └── model_list_all.json      # Complete model configuration
│   │   │   └── res/                         # Resource files
│   │   └── build.gradle.kts                 # App-level build configuration
│   ├── build.gradle.kts                     # Project-level build configuration
│   ├── gradle/                              # Gradle wrapper and dependencies
│   ├── gradlew                              # Gradle wrapper script
│   ├── gradle.properties                    # Gradle properties
│   └── settings.gradle.kts                  # Project settings
└── README.md                               # Project documentation
```

## Step-by-Step Implementation

### Step 0: Preparing to Download the Model File


```kotlin
// Parse model list configuration from assets
val modelList = parseModelListFromAssets()

// Set download directory for model files
val downloadDir = setupModelDirectory()

```

**Notes:**
- Parse model list configuration from assets
- Set download directory for model files

### Step 1: Initialize Nexa SDK Environment

This step initializes the Nexa SDK with the application context.

```kotlin
    NexaSdk.init(applicationContext)
```

**Notes:**
- Initialize Nexa SDK with application context
- Must be called before other SDK operations
- Only needs to be called once

### Step 2: Add System Prompt

This step adds system prompts to both LLM and VLM chat lists for consistent AI behavior.

```kotlin
// Add system prompt for LLM
val llmSystemPrompt = ChatMessage("system", sysPrompt)
chatList.add(llmSystemPrompt)

// Add system prompt for VLM
val vlmSystemPrompt = VlmChatMessage(
    role = "system", 
    contents = listOf(VlmContent("text", sysPrompt))
)
vlmChatList.add(vlmSystemPrompt)
```

**Notes:**
- **System prompt is crucial** - defines AI assistant behavior, role, and response style
- **Controls output format** - specify response format (e.g., Markdown, JSON, plain text)
- **Sets response constraints** - control reply length, tone, and content guidelines
- **Required for consistent responses** - without system prompt, AI responses may be inconsistent
- **Sets conversation context** - tells the AI how to behave and what role to play
- LLM uses `ChatMessage` format
- VLM uses `VlmChatMessage` format, supports text, image, audio content
- **Recommended to always include** for better conversation experience

**System Prompt Examples:**
```kotlin
// Example 1: Markdown format with length limit
val sysPrompt = "You are a helpful assistant. Always respond in Markdown format. Keep responses under 200 words."

// Example 2: JSON format for structured responses
val sysPrompt = "You are a data assistant. Always respond in valid JSON format with 'answer' and 'confidence' fields."

// Example 3: Professional tone with specific guidelines
val sysPrompt = "You are a professional consultant. Provide concise, accurate answers. Use bullet points for lists."
```

### Step 3: Download Model

This step handles model downloading with progress tracking and error handling.

```kotlin
// Download model files (example using OkDownload)
val downloadTask = DownloadTask.Builder(modelUrl, downloadDir)
    .setFilename(filename)
    .setPassIfAlreadyCompleted(true)
    .build()

downloadTask.enqueue(listener)
```

**Notes:**
- **OkDownload is optional** - you can also use other download library

### Step 4: Load Model

This step loads the selected model (LLM or VLM) with appropriate configuration.

```kotlin
// Load LLM Model
LlmWrapper.builder().llmCreateInput(
    LlmCreateInput(
        model_path = modelFile.absolutePath,
        tokenizer_path = tokenFile?.absolutePath,
        config = ModelConfig(
            nCtx = 1024,
            max_tokens = 2048,
            nThreads = 4,
            nThreadsBatch = 4,
            nBatch = 1,
            nUBatch = 1,
            nSeqMax = 1
        )
    )
).build().onSuccess { llmWrapper = it }
 .onFailure { /* handle error */ }

// Load VLM Model (plugin_id = null)
VlmWrapper.builder().vlmCreateInput(
    VlmCreateInput(
        model_path = modelFile.absolutePath,
        tokenizer_path = null,
        mmproj_path = mmprojFile?.absolutePath,
        config = ModelConfig(
            nCtx = 1024,
            max_tokens = 2048,
            nThreads = 4,
            nThreadsBatch = 4,
            nBatch = 1,
            nUBatch = 1,
            nSeqMax = 1
        ),
        plugin_id = null
    )
).build().onSuccess { vlmWrapper = it }
 .onFailure { /* handle error */ }
```

**Notes:**
- Choose LLM or VLM loader based on model type
- LLM requires model file path and tokenizer file path (no plugin_id needed)
- VLM requires model file path and mmproj file path (plugin_id = null)
- Use `ModelConfig` to configure model parameters
- Recommend executing loading operations on background threads

### Step 5: Send Message

This step handles sending messages to the loaded model with proper template conversion.

**Important: You cannot directly pass text to `generateStreamFlow`. You must first use `applyChatTemplate` to convert chat messages.**

```kotlin
// Generate text with LLM
chatList.add(ChatMessage("user", inputString))
llmWrapper.applyChatTemplate(chatList.toTypedArray(), tools, false).onSuccess { result ->
    llmWrapper.generateStreamFlow(
        result.formattedText,  // Converted text, not raw input
        GenerationConfigSample().toGenerationConfig(grammarString)
    ).collect { streamResult ->
        handleResult(sb, streamResult)
    }
}.onFailure { /* handle error */ }

// Generate response with VLM
val sendMsg = VlmChatMessage(
    role = "user", 
    contents = listOf(VlmContent("text", inputString))
)
vlmChatList.add(sendMsg)

vlmWrapper.applyChatTemplate(vlmChatList.toTypedArray(), tools, false)
    .onSuccess { result ->
        vlmWrapper.generateStreamFlow(
            result.formattedText,  // Converted text, not raw input
            GenerationConfigSample().toGenerationConfig(grammarString)
        ).collect { streamResult ->
            handleResult(sb, streamResult)
        }
    }.onFailure { /* handle error */ }

// Handle streaming results
when (streamResult) {
    is LlmStreamResult.Token -> {
        // Process generated token text
        // Update UI with streamResult.text
    }
    is LlmStreamResult.Completed -> {
        // Generation finished successfully
        // Access performance data via streamResult.profile
    }
    is LlmStreamResult.Error -> {
        // Handle generation error
        // Check streamResult.throwable for error details
    }
}
```

**Notes:**
- **Must call `applyChatTemplate` first**: Convert chat messages to formatted text
- **Cannot pass text directly**: `generateStreamFlow` requires converted text
- **Streaming processing**: Use `collect` to receive streaming generation results

**Stream Result Types:**
- **`Token`**: Contains generated text (`streamResult.text`) - append to UI progressively
- **`Completed`**: Generation finished - contains performance data (`streamResult.profile`)
- **`Error`**: Generation failed - contains error details (`streamResult.throwable`)

### Step 6: Others (Stop & Unload)

This step handles stopping stream generation and unloading models.

```kotlin
// Stop current stream generation
llmWrapper.stopStream()
// or for VLM
vlmWrapper.stopStream()

// Reset model state (clear chat context)
llmWrapper.reset()
// or for VLM
vlmWrapper.reset()

// Unload model and release resources
llmWrapper.destroy()
// or for VLM
vlmWrapper.destroy()
```

**Notes:**
- **`stopStream()`**: Stop current streaming generation
- **`reset()`**: Clear chat context and reset model state
- **`destroy()`**: Release model resources and free memory
- **Resource management**: Recommend calling when app exits or switching models


## Configuration Options

### ModelConfig

```kotlin
ModelConfig(
    nCtx = 1024,           // Context length
    max_tokens = 2048,     // Maximum generation tokens
    nThreads = 4,          // CPU threads
    nThreadsBatch = 4,     // Batch processing threads
    nBatch = 1,            // Batch size
    nUBatch = 1,           // Physical batch size
    nSeqMax = 1,           // Maximum sequences
    nGpuLayers = 0,        // GPU layers (0 for CPU only)
    config_file_path = "", // Config file path
    verbose = false        // Verbose logging
)
```

### GenerationConfig

```kotlin
// Create GenerationConfig using GenerationConfigSample
val config = GenerationConfigSample(
    maxTokens = 32,
    topK = 40,
    topP = 0.95f,
    temperature = 0.8f,
    penaltyLastN = 1.0f,
    penaltyPresent = 0.0f,
    seed = -1
).toGenerationConfig(grammarString)
```

## Best Practices

### Prompt Template Conversion

```kotlin
// ❌ Wrong: Direct text input
llm.generateStreamFlow("Hello, how are you?", config)

// ✅ Correct: Use applyChatTemplate first
chatList.add(ChatMessage("user", "Hello, how are you?"))
llmWrapper.applyChatTemplate(chatList.toTypedArray(), tools, false).onSuccess { result ->
    llm.generateStreamFlow(result.formattedText, config)
}
```

### Memory Management

```kotlin
// Stop streaming and destroy model resources
llmWrapper.stopStream()
llmWrapper.destroy()

// or for VLM
vlmWrapper.stopStream()
vlmWrapper.destroy()
```

### Error Handling

```kotlin
// Handle initialization errors
val result = LlmWrapper.builder()
    .llmCreateInput(input)
    .build()

result.onSuccess { llm ->
    // Model loaded successfully
    Log.d("LLM", "Model loaded successfully")
}.onFailure { error ->
    // Handle loading failure
    Log.e("LLM", "Model loading failed", error)
}
```

### Async Operations

```kotlin
// Perform model operations on background thread
lifecycleScope.launch(Dispatchers.IO) {
    llm.generateStreamFlow(prompt, config)
        .collect { streamResult ->
            withContext(Dispatchers.Main) {
                updateUI(streamResult)
            }
        }
}
```

## Troubleshooting

### Common Issues

1. **Model Loading Failure**
   - Check if model file path is correct
   - Verify model file integrity
   - Ensure sufficient device memory

2. **Slow Generation Speed**
   - Reduce `nCtx` and `max_tokens` parameters
   - Lower `nThreads` count
   - Ensure no other apps are consuming CPU

3. **Out of Memory**
   - Choose smaller models
   - Reduce batch size
   - Close unnecessary applications

### Debug Tips

```kotlin
// Enable verbose logging
ModelConfig(
    verbose = true
)

// Check model status
Log.d("Model", "Loaded: ${llm != null}")
Log.d("Model", "Context size: ${config.nCtx}")
```

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](https://github.com/NexaAI/nexa-sdk/blob/main/LICENSE) file for details.

## Contributing

Issues and Pull Requests are welcome to improve this tutorial project.

## Related Links

- [Nexa AI SDK Documentation](https://github.com/NexaAI/nexa-sdk)
- [Project Source Code](https://github.com/NexaAI/nexa-sdk)
- [Issue Tracker](https://github.com/NexaAI/nexa-sdk/issues)