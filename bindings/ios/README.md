# NexaAI Swift Binding

This project provides Swift bindings for running large AI models (LLM, VLM, Reranker, Embedder, etc.) locally on iOS/macOS/simulator platforms. It enables seamless integration of on-device inference into your Swift applications.

## Directory Structure

- `NexaAI`  
  Swift Package with core model inference APIs and implementations
- `Example`  
  Use demo

## Quick Start

### 1. Install Dependencies

```swift
dependencies: [
    .package(url: "https://github.com/NexaAI/nexa-sdk.git", .branch("main"))
]
```

### 2. Enable interoperability with C++

Swift code interoperates with C and Objective-C APIs by default, you should enable interoperability with C++:
`Build settings` -> `Swift Compliler - Language` -> `C++ and Objective-C Interoperability` -> `C++/Objective-C++`

### 3. Requirements

| Platform              | Minimum Swift Version | Installation |
| --------------------- | --------------------- | ------------ |
| iOS 17.0+ / macOS 14+ | 5.9 / Xcode 15.0      | SPM          |

## Use Examples

### 1. Example for LLM:

```swift
import NexaAI

let llmLlama = LLMLlama()
// path of model
let modelPath = try modelPath(of: "Qwen3-0.6B-Q8_0")
// load model
try await llmLlama.load(.init(modelPath: modelPath))

let system = "You are a helpful AI assistant"
let userMsgs = [
    "Tell me a long stroy, about 100 words",
    "How are you"
]
var messages = [ChatMessage]()
messages.append(.init(role: .system, content: system))
for userMsg in userMsgs {
    messages.append(.init(role: .user, content: userMsg))
    // generation
    let stream = try await llmLlama.generationAsyncStream(messages: messages)
    var response = ""
    for try await token in stream {
        print(token, terminator: "")
        response += token
    }
    messages.append(.init(role: .assistant, content: response))
}
```

### 2. Example for VLM:

```swift
import NexaAI

let vlmLlama = VLMLlama()
let projectPath = try modelPath(of: "mmproj-SmolVLM-256M-Instruct-Q8_0")
let modelPath = try modelPath(of: "SmolVLM-256M-Instruct-Q8_0")
try await vlmLlama.load(.init(modelPath: modelPath, mmprojPath: projectPath))

// chat
let system = "You are a helpful AI assistant"
let userMsgs = [
    "Tell me a long stroy, about 100 words",
    "How are you"
]
var messages = [ChatMessage]()
messages.append(.init(role: .system, content: system))
for userMsg in userMsgs {
    messages.append(.init(role: .user, content: userMsg))
    // generation
    let stream = try await vlmLlama.generationAsyncStream(messages: messages)
    var response = ""
    for try await token in stream {
        print(token, terminator: "")
        response += token
    }
    messages.append(.init(role: .assistant, content: response))
}

// image
let images = [try assetsPath(of: "test_image.png")]
var config = GenerationConfig.default
config.imagePaths = images
let message = ChatMessage(role: .user, content: "What do you see in this image", images: images)
let stream = try await vlmLlama.generationAsyncStream(messages: [message], options: .init(config: config))

for try await token in stream {
    print(token, terminator: "")
}
print(await vlmLlama.lastProfileData?.description ?? "")
```

### 3. Example for Embedder:

```swift
import NexaAI

let modelPath = try modelPath(of: "jina-embeddings-v2-small-en-Q4_K_M")
let embedder = try Embedder(modelPath: modelPath)
let texts = [
    "Hello world", "Good morning",
    "Machine learning is fascinating",
    "Natural language processing"
]
let cfg = EmbeddingConfig(batchSize: 4, normalize: true, normalizeMethod: .l2)
let result = try embedder.embed(texts: texts, config: cfg)
let embeddings = result.embeddings
print("Texts: ", texts)
print("Total embeddings: \(embeddings.count)")
print("Embeddings:", embeddings.prefix(20))

let count = Float(embeddings.count)
let mean = (embeddings.reduce(0.0, +)) / count

let variance =
(embeddings.map {
    let diff = mean - $0
    return diff * diff
}
    .reduce(0.0, +)) / count

let std = sqrt(variance)
print("Embedding stats: min=\(embeddings.min()!), max=\(embeddings.max()!), mean=\(mean), std=\(std)")
print("ProfileData: \n", result.profileData)
```

### 4. Other

```swift
// setup log
NexaSdk.install([.error, .trace, .debug])

// version
print(NexaSdk.version)

// get llama device list
let deviceList = NexaSdk.getLlamaDeviceList()
```

See `Tests/NexaAITests` for more usage examples.

## FAQ

- **Platform Support:** Works on iOS/macOS/simulator. Metal acceleration requires a real device.
- **Custom Model Path:** Place model files in the app sandbox or add via Xcode resources.

Let me know if you need more detailed API documentation or specific usage examples!
