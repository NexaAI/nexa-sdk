# NexaSwift

**NexaSwift** is a Swift wrapper for the [llama.cpp](https://github.com/ggerganov/llama.cpp.git) library. This repository provides a Swifty API, allowing Swift developers to easily integrate and use `llama.cpp` models in their projects.  
**NOTE:** Currently, we support text inference capabilities.

## Installation

To add NexaSwift to your Swift project, add the following dependency in your `Package.swift` file:

```swift
.package(url: "https://github.com/NexaAI/nexa-sdk.git", .branch("main"))
```

## Usage

### 1. Initialize NexaSwift with Model Path

Create a configuration and initialize NexaSwift with the path to your model file:

```swift
let configuration = NexaSwift.Configuration(
    maxNewToken: 128,
    stopTokens: []
)
let modelPath = "path/to/your/model"
let nexaSwift = try NexaSwift.NexaTextInference(modelPath: modelPath, modelConfiguration: configuration)
```

### 2 Completion chat API

#### Generate messages

```swift
var messages:[ChatCompletionRequestMessage] = []
let userMessage = ChatCompletionRequestMessage.user(
    ChatCompletionRequestUserMessage(content: .text("user input"))
)
messages.append(userMessage)
```

#### Non-Streaming Mode

For non-streaming mode, simply call the start method with your prompt. This will return the complete response once it’s available.

```swift
let response = try await nexaSwift.createChatCompletion(for: messages)
print(response.choices[0].message.content ?? "")
```

#### Streaming Mode

In streaming mode, you can process the response in real-time as it’s generated:

```swift
for try await response in await nexaSwift.createChatCompletionStream(for: messages) {
    print(response.choices[0].delta.content ?? "")
}
```

### 3 Completion API

#### Non-Streaming Mode

```swift
if let response = try? await nexaSwift.createCompletion(for: prompt) {
    print(response.choices[0].text))
}
```

#### Streaming Mode

```swift
for try await response in await nexaSwift.createCompletionStream(for: prompt) {
    print(response.choices[0].text)
}
```

## Quick Start

Open the [swift test project](../examples/swift-test/) folder in Xcode and run the project.

## Download Models

NexaSwift supports all models compatible with llama.cpp. You can download models from the [Nexa AI ModelHub](https://nexa.ai/models).
