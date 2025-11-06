
# Nexa AI Android SDK Documentation

## Overview

The Nexa AI Android SDK enables on-device AI inference for Android applications with NPU acceleration. Run Large Language Models (LLMs), Vision-Language Models (VLMs), Embeddings, Speech Recognition (ASR), Reranking, and Computer Vision models on Android devices with support for NPU, GPU, and CPU inference.

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

## Getting Started

### Step 1: Initialize SDK

Initialize once in your Application or Activity:

```kotlin
NexaSdk.getInstance().init(this)
```

### Step 2: Load Model

Load model with NPU configuration, it will automatically download the model:

```kotlin
// LLM with NPU
LlmWrapper.builder().createInput(
    model_path = "NexaAI/LFM2-1.2B-npu",
    plugin_id = "npu"
).build().onSuccess { llmWrapper = it }
```


### Step 3: Generate Text

```kotlin
llmWrapper.generateStreamFlow("Hello!", GenerationConfig()).collect { result ->
    println(result.text)
}
```

---

## LLM Usage (NPU)


### Generate Completions
```kotlin
llmWrapper.generateStreamFlow("Hello!", GenerationConfig()).onSuccess { result ->
    println(result.text)
}
```
### Streaming Completions

```kotlin
llmWrapper.generateStreamFlow("Hello!", GenerationConfig()).collect { result ->
}
```


### LLM API Reference

#### LlmCreateInput
```kotlin
data class LlmCreateInput(
    val model_name: String? = null,              // Model name (e.g., "liquid-v2")
    val model_path: String,                       // Path to .gguf model
    val tokenizer_path: String? = null,           // Path to tokenizer.json
    val config: ModelConfig,                      // Model configuration
    val plugin_id: String? = null                 // "npu" for NPU acceleration
)
```

#### ModelConfig (NPU)
```kotlin
data class ModelConfig(
    val nCtx: Int = 1024,                         // Context window size
    val max_tokens: Int = 2048,                   // Max generation tokens
    val nThreads: Int = 8,                        // Processing threads
    val nThreadsBatch: Int = 4,                   // Batch threads
    val nBatch: Int = 2048,                       // Batch size
    val nUBatch: Int = 512,                       // Physical batch size
    val nSeqMax: Int = 1,                         // Max sequences
    val qnn_lib_folder_path: String? = null,      // QNN library path (NPU)
    val qnn_model_folder_path: String? = null     // QNN model folder (NPU)
)
```

---

## VLM Usage (NPU)

### Load VLM Model

```kotlin
VlmWrapper.builder().createInput(
    VlmCreateInput(
        model_path = "NexaAI/LFM2-1.2B-npu",
        plugin_id = "npu"
    )
).build().onSuccess { vlmWrapper = it }
```

### Generate Text from image and text

```kotlin
vlmWrapper.generateStreamFlow("Hello!", GenerationConfig(
    messages = arrayOf(
        VlmChatMessage("user", listOf(
            VlmContent("image", "/path/to/image.jpg"),
            VlmContent("text", "What's in this image?")
        ))
    )
)).collect { result ->
}
```

### VLM API Reference

#### VlmCreateInput
```kotlin
data class VlmCreateInput(
    val model_path: String,                       // Path to VLM model
    val mmproj_path: String? = null,              // Vision projection weights
    val config: ModelConfig,                      // Model configuration
    val plugin_id: String? = null                 // null for CPU/GPU
)
```

---

## Embeddings Usage (NPU)

### Load Embedder Model

```kotlin
EmbedderWrapper.builder().embedderCreateInput(
    EmbedderCreateInput(
        model_name = "embed-gemma",
        model_path = "/path/to/embedder.gguf",
        tokenizer_path = "/path/to/tokenizer.json",
        config = ModelConfig(
            qnn_lib_folder_path = applicationInfo.nativeLibraryDir,
            qnn_model_folder_path = "/path/to/qnn/models"
        ),
        plugin_id = "npu"
    )
).build().onSuccess { embedderWrapper = it }
```

### Generate Embeddings

```kotlin
val texts = arrayOf("What is AI?", "Machine learning explained")
embedderWrapper.embed(texts, EmbeddingConfig()).onSuccess { embeddings ->
    val dimension = embeddings.size / texts.size
    println("Embedding dimension: $dimension")
}
```

### Embeddings API Reference

#### EmbedderCreateInput
```kotlin
data class EmbedderCreateInput(
    val model_name: String? = null,               // Model name (e.g., "embed-gemma")
    val model_path: String,                       // Path to embedder model
    val tokenizer_path: String? = null,           // Path to tokenizer
    val config: ModelConfig,                      // Model configuration
    val plugin_id: String? = null                 // "npu" for NPU
)
```

---

## ASR Usage (NPU)

### Load ASR Model

```kotlin
AsrWrapper.builder().asrCreateInput(
    AsrCreateInput(
        model_name = "parakeet",
        model_path = "/path/to/asr-model.gguf",
        config = ModelConfig(
            qnn_lib_folder_path = applicationInfo.nativeLibraryDir,
            qnn_model_folder_path = "/path/to/qnn/models"
        ),
        plugin_id = "npu"
    )
).build().onSuccess { asrWrapper = it }
```

### Transcribe Audio

```kotlin
asrWrapper.transcribe(
    AsrTranscribeInput("/path/to/audio.wav", "en", null)
).onSuccess { result ->
    println("Transcript: ${result.result.transcript}")
}
```

### ASR API Reference

#### AsrCreateInput
```kotlin
data class AsrCreateInput(
    val model_name: String? = null,               // Model name (e.g., "parakeet")
    val model_path: String,                       // Path to ASR model
    val config: ModelConfig,                      // Model configuration
    val plugin_id: String? = null                 // "npu" for NPU
)
```

#### AsrTranscribeInput
```kotlin
data class AsrTranscribeInput(
    val audioPath: String,                        // Path to audio file
    val language: String,                         // Language code: "en", "zh", etc.
    val timestamps: String? = null                // Optional timestamps
)
```

---

## Reranker Usage (NPU)

### Load Reranker Model

```kotlin
RerankerWrapper.builder().rerankerCreateInput(
    RerankerCreateInput(
        model_name = "jina-rerank",
        model_path = "/path/to/reranker.gguf",
        tokenizer_path = "/path/to/tokenizer.json",
        config = ModelConfig(
            qnn_lib_folder_path = applicationInfo.nativeLibraryDir,
            qnn_model_folder_path = "/path/to/qnn/models"
        ),
        plugin_id = "npu"
    )
).build().onSuccess { rerankerWrapper = it }
```

### Rerank Documents

```kotlin
val query = "What is machine learning?"
val docs = arrayOf("ML is AI subset", "Pizza recipe", "Neural networks")
rerankerWrapper.rerank(query, docs, RerankConfig()).onSuccess { result ->
    result.scores?.withIndex()?.sortedByDescending { it.value }?.forEach { (idx, score) ->
        println("${idx + 1}. Score: $score - ${docs[idx]}")
    }
}
```

### Reranker API Reference

#### RerankerCreateInput
```kotlin
data class RerankerCreateInput(
    val model_name: String? = null,               // Model name (e.g., "jina-rerank")
    val model_path: String,                       // Path to reranker model
    val tokenizer_path: String? = null,           // Path to tokenizer
    val config: ModelConfig,                      // Model configuration
    val plugin_id: String? = null                 // "npu" for NPU
)
```

---

## Computer Vision Usage (NPU)

### Load CV Model (PaddleOCR)

```kotlin
CvWrapper.builder().createInput(
    CVCreateInput(
        model_name = "paddleocr",
        config = CVModelConfig(
            capabilities = CVCapability.OCR,
            det_model_path = "/path/to/det_model",
            rec_model_path = "/path/to/rec_model",
            char_dict_path = "/path/to/char_dict",
            qnn_model_folder_path = "/path/to/qnn/models",
            qnn_lib_folder_path = applicationInfo.nativeLibraryDir
        ),
        plugin_id = "npu"
    )
).build().onSuccess { cvWrapper = it }
```

### Perform OCR

```kotlin
cvWrapper.infer("/path/to/image.jpg").onSuccess { results ->
    results.forEach { println("Text: ${it.text}, Confidence: ${it.confidence}") }
}
```

### CV API Reference

#### CVCreateInput
```kotlin
data class CVCreateInput(
    val model_name: String,                       // Model name (e.g., "paddleocr")
    val config: CVModelConfig,                    // CV model configuration
    val plugin_id: String? = null                 // "npu" for NPU
)
```

#### CVModelConfig
```kotlin
data class CVModelConfig(
    val capabilities: CVCapability,               // OCR, DETECTION, CLASSIFICATION
    val det_model_path: String? = null,           // Detection model path
    val rec_model_path: String? = null,           // Recognition model path
    val char_dict_path: String? = null,           // Character dictionary
    val qnn_model_folder_path: String? = null,    // QNN model folder (NPU)
    val qnn_lib_folder_path: String? = null       // QNN library path (NPU)
)
```

---

## Complete Example

```kotlin
class MainActivity : Activity() {
    private val modelScope = CoroutineScope(Dispatchers.IO)
    private lateinit var llmWrapper: LlmWrapper
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Step 1: Initialize SDK
        NexaSdk.getInstance().init(this)
        
        // Step 2: Load model
        loadModel()
    }
    
    private fun loadModel() {
        modelScope.launch {
            LlmWrapper.builder().llmCreateInput(
                LlmCreateInput(
                    model_name = "liquid-v2",
                    model_path = File(filesDir, "model.gguf").absolutePath,
                    config = ModelConfig(
                        nCtx = 1024,
                        qnn_lib_folder_path = applicationInfo.nativeLibraryDir,
                        qnn_model_folder_path = filesDir.absolutePath
                    ),
                    plugin_id = "npu"
                )
            ).build().onSuccess { llmWrapper = it }
        }
    }
    
    private fun generateText(prompt: String) {
        modelScope.launch {
            val chatList = arrayListOf(ChatMessage("user", prompt))
            llmWrapper.applyChatTemplate(chatList.toTypedArray(), null, false)
                .onSuccess { template ->
                    llmWrapper.generateStreamFlow(template.formattedText, GenerationConfig())
                        .collect { result ->
                            when (result) {
                                is LlmStreamResult.Token -> runOnUiThread {
                                    tvResult.append(result.text)
                                }
                                is LlmStreamResult.Completed -> Log.d("LLM", "Done!")
                                is LlmStreamResult.Error -> Log.e("LLM", "Error: ${result.throwable}")
                            }
                        }
                }
        }
    }
}
```

For more API usage examples, please refer to our documentation: [nexa-sdk-android-docs](https://docs.nexa.ai/nexa-sdk-android/APIReference)

## Model Download

You can download the required model from our official website: [sdk.nexa.ai](https://sdk.nexa.ai/model) or from [Hugging Face](https://huggingface.co/NexaAI).

### Adding a New Model

To add a new model for testing or use, follow these steps:

1. Locate the `model_list.json` file in the `assets` folder of the demo.
2. Add the configuration for the new model you want to test or use, including the download URL and other relevant settings.
3. Save the file and restart the demo to make the new model available.

If you want to use a local model directly, simply specify the path to the model.