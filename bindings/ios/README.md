## Overview

Nexa SDK enables on-device AI inference on iPhone / iPad / macOS (on supported platforms). It covers a wide range of capabilities including Large Language Models (LLM), Vision-Language Models (VLM), Text Embeddings (Embedder), Automatic Speech Recognition (ASR), Reranking (Reranker), and more.

The SDK supports accelerated inference using Apple Neural Engine (ANE), Metal GPU, and CPU, making it easy to integrate offline or edge AI capabilities into iOS and macOS applications.

For more detailed documentation, please refer to the official [docs](https://docs.nexa.ai/en/nexa-sdk-ios/overview)

## Device Compatibility

### Supported Inference Backends
- **Apple Neural Engine (ANE)**: Provides optimal performance on supported Apple silicon.
- **GPU (Metal)**: Accelerated inference on the GPU.
- **CPU**: Fallback option when hardware acceleration is unavailable.

### Minimum Requirements (Recommended)
- **Xcode**: 16.4 or later (recommended)
- **iOS/macOS**: iOS 17+ / macOS 15+
- **Architecture**: arm64
- **Device**: Devices with ANE are recommended for optimal performance

## Installation
- Download [NexaSdk.xcframework](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/ios/latest/NexaSdk.xcframework.zip).
- Unzip and add it to your Xcode Project / Workspace.
- Make sure the framework is listed under Frameworks, Libraries, and Embedded Content for your target, and set it to Embed & Sign if required.

## Models

Download and prepare models according to the SDK documentation or the model repository instructions. Optional example models include:
- [`NexaAI/parakeet-tdt-0.6b-v3-ane`](https://huggingface.co/NexaAI/parakeet-tdt-0.6b-v3-ane)
- [`NexaAI/EmbedNeural`](https://huggingface.co/NexaAI/EmbedNeural)
- [`NexaAI/Qwen3-0.6B-GGUF`](https://huggingface.co/NexaAI/Qwen3-0.6B-GGUF)

Place model files in a location accessible by the app:
- During development, you can drag models into App Resources in Xcode
- In production, models can be downloaded from a remote source into the app sandbox and loaded by the SDK

## Quick Start (Swift)

Below are Swift examples for common use cases

### ASR (Automatic Speech Recognition)

```swift
import NexaSdk
import Foundation

let repoDir: URL = URL(string: "...")!
let asr = try Asr()
try await asr.load(from: repoDir)

print(asr.supportedLanguages())

let audioPath = Bundle.main.path(forResource: "test", ofType: "wav")!
let result = try await asr.transcribe(options: .init(audioPath: audioPath))

print(result.asrResult)
```

### Embedder (Text Embeddings)

```swift
import NexaSdk
import Foundation

let repoDir: URL = URL(string: "...")!
let embedder = try Embedder(from: repoDir)

let texts = [
    "Hello world", "Good morning",
    "Machine learning is fascinating",
    "Natural language processing"
]

let result = try embedder.embed(
    texts: texts,
    config: .init(batchSize: Int32(texts.count))
)

print(result.embeddings.prefix(10))
```

### LLM (Large Language Model)

```swift
import NexaSdk
import Foundation

let llm = try LLM()

let repoDir: URL = URL(string: "...")!
try await llm.load(from: repoDir)

var messages: [ChatMessage] = [
    ChatMessage(role: .user, content: "Hello, tell me a story")
]

let prompt = try await llm.applyChatTemplate(messages: messages)
let stream = await llm.generateAsyncStream(prompt: prompt)

var response = ""
for try await token in stream {
    print(token, terminator: "")
    response += token
}
```

For more examples, please refer to the API Reference￼.

### Additional

```swift
// Set log levels
NexaAI.install([.error, .trace, .debug])

// Print SDK version
print(NexaAI.version)
```

## FAQ & Debugging
- If model loading fails, verify that the model path is correct and that all model files are complete.
- When testing ANE/Metal acceleration on real devices, first confirm the device OS version and chip support.
- If you encounter crashes or missing symbols, ensure that `NexaSdk.xcframework` is correctly set to `Embed & Sign` in your target settings.