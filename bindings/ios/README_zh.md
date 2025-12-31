
## 概述

Nexa SDK 支持在 iPhone / iPad / macOS（在适配平台）上进行本地 AI 推理，覆盖大语言模型（LLM）、视觉-语言模型（VLM）、文本向量化（Embedder）、语音识别（ASR）、重排序（Reranker）等能力。SDK 支持使用 Apple Neural Engine（ANE）、Metal GPU 与 CPU 进行加速推理，方便将离线或边缘推理能力集成到 iOS/macOS 应用中。

更多详细内容请参见[官方文档](https://docs.nexa.ai/cn/nexa-sdk-ios/overview)

## 设备兼容性

### 支持的推理后端

- **Apple Neural Engine (ANE)**: 在支持的 Apple 芯片上可获得最优性能。  
- **GPU (Metal)**: 在 GPU 上加速推理。  
- **CPU**: 在无硬件加速时回退到 CPU。

### 最低要求（建议）

- **Xcode**: 16.4 或更高（推荐）。  
- **iOS/macOS**: iOS 17+ / macOS 15+（运行时依赖具体模型需求，建议使用较新系统以获得最佳性能）。  
- **架构**: arm64。  
- **设备**: 推荐具备 ANE 的设备以发挥最佳性能。  

## 安装

- 下载[`NexaSdk.xcframework`](https://nexa-model-hub-bucket.s3.us-west-1.amazonaws.com/public/ios/latest/NexaSdk.xcframework.zip)
- 将Framework添加到目标的 **Frameworks, Libraries, and Embedded Content**，并根据需要设置为 **Embed & Sign**  

## 模型

按照 iOS SDK 文档或模型仓库的说明下载并准备模型。示例模型（可选）包括：  
- [`NexaAI/parakeet-tdt-0.6b-v3-ane`](https://huggingface.co/NexaAI/parakeet-tdt-0.6b-v3-ane)
- [`NexaAI/EmbedNeural`](https://huggingface.co/NexaAI/EmbedNeural)
- [`NexaAI/Qwen3-0.6B-GGUF`](https://huggingface.co/NexaAI/Qwen3-0.6B-GGUF) 

将模型文件放入 App 可访问的位置：  
- 开发调试阶段可以将模型拖入 Xcode 的 App Resources  
- 生产环境中建议从远端下载到应用的沙盒目录并由 SDK 加载

## 快速使用（Swift）

下面给出若干常用功能的 Swift 示例。这些示例与 SDK 的接口一致，可直接在 iOS/macOS 项目中调用。

### ASR（语音识别）示例

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

### Embedder（文本向量化）示例

```swift
import NexaSdk
import Foundation

let repoDir: URL = URL(string: "...")
let embedder = try Embedder(from: repoDir)

let texts = [
    "Hello world", "Good morning",
    "Machine learning is fascinating",
    "Natural language processing"
]
let result = try embedder.embed(texts: texts, config: .init(batchSize: Int32(texts.count)))
print(result.embeddings.prefix(10))
```

### LLM（大语言模型）示例

```swift
import NexaSdk
import Foundation

let llm = try LLM()

let repoDir: URL = URL(string: "...")!
try await llm.load(from: repoDir)

var messages: [ChatMessage] = [
    ChatMessage(role: .user, content: "Hello, Tell me a story")
]

var prompt = try await llm.applyChatTemplate(messages: messages)
var stream = await llm.generateAsyncStream(prompt: prompt)
var response = ""
for try await token in stream {
    print(token, terminator: "")
    response += token
}
```

更多示例参考接口[说明文档](http://localhost:3000/en/nexa-sdk-ios/APIReference)

### 其他

```swift
// 设置日志等级
NexaAI.install([.error, .trace, .debug])

// 查看 SDK 版本
print(NexaAI.version)
```

## 常见问题与调试

- 如果模型加载失败，请确认模型路径正确且模型文件完整。  
- 在真机上测试 ANE/Metal 加速时，优先检查设备系统版本和芯片支持列表。  
- 若遇到崩溃或符号缺失，确保 `NexaSdk.xcframework` 已被正确 `Embed & Sign`。
