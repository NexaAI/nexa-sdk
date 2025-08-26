# NexaAI Swift Binding

This directory provides Swift bindings for running large AI models (LLM, VLM, Reranker, Embedder, etc.) locally on iOS/macOS/simulator platforms. It enables seamless integration of on-device inference into your Swift applications.

## Directory Structure

- `NexaAI`  
  Swift Package with core model inference APIs and implementations
- `Example`  
  Use demo

## Quick Start

### 1. Install Dependencies

The [Swift Package Manager](https://swift.org/package-manager/) is a tool for automating the distribution of Swift code and is integrated into the `swift` compiler.

Once you have your Swift package set up, adding NexaAI as a dependency is as easy as adding it to the `dependencies` value of your `Package.swift` or the Package list in Xcode.

```swift
dependencies: [
    .package(url: "https://github.com/NexaAI/nexa-sdk.git", .branch("main"))
]
```

### 2. Requirements

| Platform               | Minimum Swift Version | Installation                                   |
| -----------------------| --------------------- | -----------------------------------------------|
| iOS 17.0+ / macOS 14+  | 5.9 / Xcode 15.0      | [Swift Package Manager](#swift-package-manager)|


### 2. Use the Model in Swift

Example for LLM:

```swift
import NexaAI

let llmLlama = LLMLlama()
// path of model
let modelPath = try modelPath(of: "Qwen3-0.6B-Q8_0")
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
    let stream = try await llmLlama.generationAsyncStream(messages: messages)
    var response = ""
    for try await token in stream {
        print(token, terminator: "")
        response += token
    }
    messages.append(.init(role: .assistant, content: response))
}
```

See `Tests/NexaAITests` for more usage examples.

## FAQ

- **Platform Support:** Works on iOS/macOS/simulator. Metal acceleration requires a real device.
- **Custom Model Path:** Place model files in the app sandbox or add via Xcode resources.

Let me know if you need more detailed API documentation or specific usage examples!