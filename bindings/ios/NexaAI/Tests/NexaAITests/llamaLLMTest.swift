import Testing
import Foundation

@testable import NexaAI

func modelPath(of name: String) throws -> String {
    try resoucePath(of: name, ext: "gguf", dir: "llama_cpp")
}

func assetsPath(of name: String, ext: String = "") throws -> String {
    try resoucePath(of: name, ext: ext, dir: "assets")
}

func resoucePath(of name: String, ext: String, dir: String) throws -> String {
#if os(macOS)
    let currentFile = URL(fileURLWithPath: #file)
    let path: String
    if dir.isEmpty {
        path = ext.isEmpty ? "../../../../../modelfiles/\(name)" : "../../../../../modelfiles/\(name).\(ext)"
    } else {
        path = ext.isEmpty ? "../../../../../modelfiles/\(dir)/\(name)" : "../../../../../modelfiles/\(dir)/\(name).\(ext)"
    }

    let resoucePath = currentFile
        .deletingLastPathComponent()
        .appendingPathComponent(path)
        .standardized.path
    return resoucePath
#else
    guard let modelFileURL = Bundle.module.url(forResource: name, withExtension: ext) else {
        throw TestError.fileNotFound(name)
    }
    return modelFileURL.path()
#endif
}

struct LLamaLLMTest {

    init() {
        NexaSdk.install()
    }

    @Test func testLlamaLLMLoad() async throws {
        let llmLlama = LLMLlama()
        let modelPath = try modelPath(of: "Qwen3-0.6B-Q8_0")
        try await llmLlama.load(.init(modelPath: modelPath))
    }

    @Test func testLlamaLLMGenerationAsyncStream() async throws {
        let llmLlama = LLMLlama()
        let modelPath = try modelPath(of: "Qwen3-0.6B-Q8_0")
        try await llmLlama.load(.init(modelPath: modelPath))

        let system = "You are a helpful, concise, and privacy-respecting AI assistant running fully on-device. Provide accurate, unbiased answers across a wide range of topics. When unsure, state so clearly. Avoid speculation. Always prioritize clarity, relevance, and user control."
        let userMsgs = [
            "1+1=多少?",
            "那2+2呢?",
            "n+n呢?",
            "Tell me a long stroy, about 100 words",
            "今天星期几？",
            "一年多少天？",
            "2025年是闰年么？",
            "如何学习英语？"
        ]
        var messages = [ChatMessage]()
        messages.append(.init(role: .system, content: system))
        for userMsg in userMsgs {
            messages.append(.init(role: .user, content: userMsg))
            let stream = try await llmLlama.generationAsyncStream(messages: messages)
            print("-----------------------------")
            var response = ""
            for try await token in stream {
                print(token, terminator: "")
                response += token
            }
            print("\n")
            messages.append(.init(role: .assistant, content: response))
            print(await llmLlama.lastProfileData?.description ?? "")
            print("-----------------------------")
        }
    }

    @Test func testQwen3LLamaLLMFunctionTools() async throws {
        try await testLlamaLLMFunctionTools("Qwen3-0.6B-Q8_0")
    }

    @Test func testLFMLLamaLLMFunctionTools() async throws {
        try await testLlamaLLMFunctionTools("LFM2-1.2B-Q4_0")
    }

    func testLlamaLLMFunctionTools(_ modelName: String) async throws {
        let llmLlama = LLMLlama()
        let modelPath = try modelPath(of: modelName)
        try await llmLlama.load(.init(modelPath: modelPath))

        let weatherTool = """
            [
              {
                "type": "function",
                "function": {
                  "name": "get_current_weather",
                  "description": "Get the current weather in a given location",
                  "parameters": {
                    "type": "object",
                    "properties": {
                      "location": {
                        "type": "string",
                        "description": "The city and state, e.g. San Francisco, CA"
                      },
                      "unit": {
                        "type": "string",
                        "enum": ["celsius", "fahrenheit"]
                      }
                    },
                    "required": ["location"]
                  }
                }
              }
            ]
            """
        let userMsgs = [
            "What is the weather like in Boston today?"
        ]
        var messages = [ChatMessage]()
        let options = GenerationOptions(templeteOptions: .init(tools: weatherTool, enableThinking: true))
        for userMsg in userMsgs {
            messages.append(.init(role: .user, content: userMsg))
            let stream = try await llmLlama.generationAsyncStream(messages: messages, options: options)
            print("-----------------------------")
            var response = ""
            for try await token in stream {
                print(token, terminator: "")
                response += token
            }
            print("\n")
            messages.append(.init(role: .assistant, content: response))
            print(await llmLlama.lastProfileData?.description ?? "")
            print("-----------------------------")
        }
    }
}
