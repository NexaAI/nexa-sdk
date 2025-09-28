# Android

## Overview

NexaAI Android binding provides a complete solution for running AI models on Android platforms, supporting Large Language Models (LLM) and Vision-Language Models (VLM).

## Development Status

✅ **Available**: The SDK is now published to Maven Central and ready for production use.

**Current Status**: Production Ready
- Core functionality implemented
- JNI bindings complete
- Kotlin API wrapper ready
- Published to Maven Central
- Available via dependency management

## Features

- **Multi-Model Support**: LLM, VLM
- **Streaming Generation**: Real-time streaming output based on Kotlin Flow
- **Coroutine Support**: Fully async API based on Kotlin coroutines
- **Builder Pattern**: Clean and easy-to-use builder API
- **Android Optimized**: Memory and performance management optimized for Android
- **JNI Bridge**: Efficient C++ to Kotlin bridging

## Quick Start

### 1. Add Dependencies

The SDK is now available on Maven Central. Add the dependency to your `build.gradle.kts`:

```kotlin
dependencies {
    // Nexa AI SDK from Maven Central
    implementation("io.github.nfl-nexa:core:0.0.2")
}
```

**Note**: Make sure your project has Maven Central repository configured in your `settings.gradle.kts`:

```kotlin
dependencyResolutionManagement {
    repositories {
        google()
        mavenCentral()
    }
}
```

### 2. Configure Native Libraries (jniLibs)

The SDK requires native libraries to function properly. Configure your `build.gradle.kts`:

```kotlin
android {
    // ... other configurations
    
    sourceSets {
        getByName("main") {
            jniLibs.srcDirs("src/main/jniLibs")
        }
    }
    
    packagingOptions {
        jniLibs.useLegacyPackaging = true
    }
}
```

**Important**: Create the `src/main/jniLibs/arm64-v8a/` directory and add the `libomp.so` file:

```bash
mkdir -p src/main/jniLibs/arm64-v8a
# Copy libomp.so to this directory
```

### 3. Initialize SDK Environment

**Critical**: You must initialize the SDK environment before using any AI models:

```kotlin
import android.system.Os
import java.io.File

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Initialize Nexa SDK environment - REQUIRED
        initNexaSdk()
        
        // ... rest of your code
    }
    
    private fun initNexaSdk() {
        try {
            val nativeLibPath = applicationInfo.nativeLibraryDir
            Os.setenv("ADSP_LIBRARY_PATH", nativeLibPath, true)
            Os.setenv("LD_LIBRARY_PATH", nativeLibPath, true)
            Os.setenv("NEXA_PLUGIN_PATH", nativeLibPath, true)
            Log.d("MainActivity", "Nexa SDK environment initialized")
            Log.d("MainActivity", "Native library path: $nativeLibPath")
            
            // Verify native libraries are available
            val libDir = File(nativeLibPath)
            if (libDir.exists() && libDir.isDirectory) {
                val files = libDir.listFiles()
                Log.d("MainActivity", "Native libraries found: ${files?.size}")
            }
        } catch (e: Exception) {
            Log.e("MainActivity", "Failed to initialize Nexa SDK environment", e)
        }
    }
}
```

**⚠️ Critical**: Without proper SDK initialization, model loading will fail with error codes `-100001` and `-100000`.

### 4. Basic Usage

```kotlin
import com.nexa.sdk.*
import kotlinx.coroutines.*

class MainActivity : AppCompatActivity() {
    private var llm: LlmWrapper? = null
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Initialize model in background thread
        lifecycleScope.launch(Dispatchers.IO) {
            llm = LlmWrapper.builder()
                .llmCreateInput(LlmCreateInput(
                    model_path = "/path/to/model.gguf",
                    config = ModelConfig()
                ))
                .build()
                .getOrThrow()
        }
    }
    
    private fun generateText() {
        lifecycleScope.launch {
            llm?.generateStreamFlow("Hello, world!", GenerationConfig())
                ?.collect { result ->
                    when (result) {
                        is LlmStreamResult.Token -> {
                            // Update UI with generated text
                            runOnUiThread { 
                                textView.append(result.text) 
                            }
                        }
                        is LlmStreamResult.Completed -> {
                            // Generation completed
                            Log.d("LLM", "Generation completed")
                        }
                        is LlmStreamResult.Error -> {
                            // Handle error
                            Log.e("LLM", "Generation error", result.throwable)
                        }
                    }
                }
        }
    }
    
    override fun onDestroy() {
        super.onDestroy()
        llm?.close() // Release resources
    }
}
```

## API Reference

### LlmWrapper - Large Language Model

#### Creating Instance

```kotlin
val llm = LlmWrapper.builder()
    .llmCreateInput(LlmCreateInput(
        model_path = "/path/to/model.gguf",
        tokenizer_path = "/path/to/tokenizer", // Optional
        config = ModelConfig(
            nGpuLayers = 0, // Number of GPU layers
            nCtx = 2048     // Context length
        ),
        plugin_id = "llama-cpp", // Plugin ID
        device_id = "cpu"         // Device ID
    ))
    .dispatcher(Dispatchers.IO) // Coroutine dispatcher
    .build()
    .getOrThrow()
```

#### Streaming Generation

```kotlin
// Stream text generation
llm.generateStreamFlow(prompt, GenerationConfig(
    maxTokens = 100,
    samplerConfig = SamplerConfig(
        temperature = 0.7f,
        topP = 0.9f
    )
)).collect { result ->
    when (result) {
        is LlmStreamResult.Token -> print(result.text)
        is LlmStreamResult.Completed -> println("Completed")
        is LlmStreamResult.Error -> println("Error: ${result.throwable}")
    }
}
```

#### Chat Template

```kotlin
val messages = arrayOf(
    ChatMessage(role = "system", content = "You are a helpful assistant"),
    ChatMessage(role = "user", content = "Hello!")
)

val formattedPrompt = llm.applyChatTemplate(
    messages = messages,
    tools = null,
    enableThinking = false
).getOrThrow()

println(formattedPrompt.formattedText)
```

### VlmWrapper - Vision-Language Model

#### Creating Instance

```kotlin
val vlm = VlmWrapper.builder()
    .vlmCreateInput(VlmCreateInput(
        model_path = "/path/to/vlm_model.gguf",
        mmproj_path = "/path/to/mmproj.gguf", // Multimodal projection layer
        config = ModelConfig(),
        plugin_id = "llama-cpp"
    ))
    .build()
    .getOrThrow()
```

#### Visual Question Answering

```kotlin
val messages = arrayOf(
    VlmChatMessage(
        role = "user",
        content = "What's in this image?",
        imagePath = "/path/to/image.jpg"
    )
)

val formattedPrompt = vlm.applyChatTemplate(messages, null, false)
    .getOrThrow()

vlm.generateStreamFlow(formattedPrompt.formattedText, GenerationConfig())
    .collect { result ->
        when (result) {
            is LlmStreamResult.Token -> print(result.text)
            is LlmStreamResult.Completed -> println("Answer completed")
            is LlmStreamResult.Error -> println("Error: ${result.throwable}")
        }
    }
```


## Configuration Options

### ModelConfig

```kotlin
ModelConfig(
    nGpuLayers = 0,          // Number of GPU layers (0 = CPU only)
    nCtx = 2048,            // Context length
    nBatch = 512,           // Batch size
    nThreads = 4            // Number of threads
)
```

### GenerationConfig

```kotlin
GenerationConfig(
    maxTokens = 100,         // Maximum tokens to generate
    stopWords = arrayOf("</s>", "<|endoftext|>"), // Stop words
    samplerConfig = SamplerConfig(
        temperature = 0.7f,      // Temperature parameter
        topP = 0.9f,            // Top-p sampling
        topK = 40,              // Top-k sampling
        repetitionPenalty = 1.1f // Repetition penalty
    )
)
```


## Best Practices

### 1. Lifecycle Management

```kotlin
class MainActivity : AppCompatActivity() {
    private var llm: LlmWrapper? = null
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Initialize in background thread
        lifecycleScope.launch(Dispatchers.IO) {
            llm = LlmWrapper.builder()
                .llmCreateInput(input)
                .build()
                .getOrThrow()
        }
    }
    
    override fun onDestroy() {
        super.onDestroy()
        llm?.close() // Important: Release resources
    }
}
```

### 2. Error Handling

```kotlin
lifecycleScope.launch {
    val result = llm?.generateStreamFlow(prompt, config)
        ?.catch { e ->
            Log.e("LLM", "Generation error", e)
            // Handle error
        }
        ?.collect { streamResult ->
            when (streamResult) {
                is LlmStreamResult.Token -> updateUI(streamResult.text)
                is LlmStreamResult.Completed -> onComplete()
                is LlmStreamResult.Error -> handleError(streamResult.throwable)
            }
        }
}
```

### 3. Performance Optimization

```kotlin
// Use appropriate dispatcher
val llm = LlmWrapper.builder()
    .llmCreateInput(input)
    .dispatcher(Dispatchers.IO) // For I/O intensive operations
    .build()
    .getOrThrow()

```

### 4. Memory Management

```kotlin
// Release resources promptly
try {
    val llm = LlmWrapper.builder()
        .llmCreateInput(input)
        .build()
        .getOrThrow()
    
    // Use model...
    
} finally {
    llm?.close() // Ensure resources are released
}
```

## Build Configuration

### Gradle Configuration

```kotlin
android {
    compileSdk = 35
    minSdk = 27
    
    defaultConfig {
        ndk {
            abiFilters += listOf("arm64-v8a")
        }
    }
    
    externalNativeBuild {
        cmake {
            cppFlags += "-std=c++17"
            arguments += listOf(
                "-DNEXA_DL=ON",
                "-DNEXA_ANDROID=ON"
            )
        }
    }
}
```

### Required Permissions

```xml
<uses-permission android:name="android.permission.READ_EXTERNAL_STORAGE" />
<uses-permission android:name="android.permission.WRITE_EXTERNAL_STORAGE" />
```

## Example Project

Check the `Example` directory for a complete sample project showing how to use NexaAI Android binding in real applications.

## Troubleshooting

### Common Issues

#### 1. Model Loading Fails with Error Codes -100001/-100000

**Problem**: Model initialization fails with error codes `-100001` (Invalid input) or `-100000` (Unknown error).

**Solution**: 
1. Ensure SDK environment is properly initialized:
   ```kotlin
   // Must call this before using any AI models
   initNexaSdk()
   ```

2. Verify native libraries are properly configured:
   ```kotlin
   // Check if jniLibs is configured in build.gradle.kts
   sourceSets {
       getByName("main") {
           jniLibs.srcDirs("src/main/jniLibs")
       }
   }
   ```

3. Ensure `libomp.so` is present in `src/main/jniLibs/arm64-v8a/`

#### 2. Native Library Not Found

**Problem**: `UnsatisfiedLinkError` or similar native library errors.

**Solution**:
1. Verify Maven dependency contains native libraries:
   ```bash
   # Check if the dependency is properly resolved
   ./gradlew app:dependencies | grep "io.github.nfl-nexa"
   ```

2. Check jniLibs configuration in build.gradle.kts

3. Ensure proper packaging options:
   ```kotlin
   packagingOptions {
       jniLibs.useLegacyPackaging = true
   }
   ```

#### 3. Model File Not Found

**Problem**: Model files cannot be loaded from external storage.

**Solution**:
1. Use Storage Access Framework (SAF) for file selection
2. Copy files to internal cache directory
3. Ensure proper permissions are granted

### Debug Checklist

- [ ] SDK environment initialized (`initNexaSdk()` called)
- [ ] Native libraries configured (`jniLibs.srcDirs` set)
- [ ] `libomp.so` present in jniLibs directory
- [ ] Maven dependency properly resolved
- [ ] Model files are accessible and valid
- [ ] Proper permissions granted (storage, camera if needed)

## Support

For questions or suggestions, please visit:
- GitHub Issues: [nexasdk-bridge/issues](https://github.com/nexai-ai/nexasdk-bridge/issues)
- Documentation: [NexaAI Documentation](https://docs.nexa.ai)
