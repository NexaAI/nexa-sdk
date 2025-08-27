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
    guard let modelFileURL = Bundle.main.url(forResource: name, withExtension: ext) else {
        throw NSError(domain: "file load error", code: 404)
    }
    return modelFileURL.path()
#endif
}

struct LLamaLLMTest {

    init() {
        NexaSdk.install()
    }

    func loadModel(name: String = "Qwen3-0.6B-Q8_0") async throws -> LLMLlama {
        let llmLlama = LLMLlama()
        let modelPath = try modelPath(of: name)
        try await llmLlama.load(.init(modelPath: modelPath))
        return llmLlama
    }

    @Test func testLlamaLLMLoad() async throws {
        let llmLlama = LLMLlama()
        let modelPath = try modelPath(of: "Qwen3-0.6B-Q8_0")
        try await llmLlama.load(.init(modelPath: modelPath))
    }

    @Test func testKVCacheSave() async throws {
        let llmLlama = try await loadModel()
        let result = try await llmLlama.generationStream(messages: [.init(role: .user, content: "Tell me a story about 100 words")])
        print(result.response)
        try llmLlama.saveKVCache(to: "./kvcache")
    }

    @Test func testKVCacheLoad() async throws {
        let llmLlama = try await loadModel()
        let result = try await llmLlama.generationStream(messages: [.init(role: .user, content: "Tell me a story about 100 words")])
        print(result.response)
        try llmLlama.saveKVCache(to: "./kvcache")

        await llmLlama.reset()

        try llmLlama.loadKVCache(from: "./kvcache")
    }

    @Test func generateStream() async throws {
        let llmLlama = try await loadModel()
        let config = GenerationConfig(maxTokens: 32)
        let result = try await llmLlama.generationStream(prompt: "Tell me a story about 100 words", config: config) { token in
            print(token, terminator: "")
            return true
        }
        print("\n")
        print(result.response)
        print(result.profileData)
    }

    @Test func testGenerateChatMultiRound() async throws {
        let llmLlama = try await loadModel()

        let system = "You are a helpful AI assistant"
        let userMsgs = [
            "repeat what I wrote before and print here: 31415926535",
            "repeat what I wrote last time",
            "Tell me a long stroy, about 100 words",
            "How to learn Chinese?"
        ]
        var messages = [ChatMessage]()
        messages.append(.init(role: .system, content: system))
        for userMsg in userMsgs {
            messages.append(.init(role: .user, content: userMsg))
            let stream = try await llmLlama.generationAsyncStream(messages: messages)
            print("-----------------------------")
            print("User: ", userMsg)
            print("AI: ")
            var response = ""
            for try await token in stream {
                print(token, terminator: "")
                response += token
            }
            print("\n")
            messages.append(.init(role: .assistant, content: response))
            print(await llmLlama.lastProfileData?.description ?? "")
        }
    }

    @Test func testQwen3FunctionTools() async throws {
        try await testLLMFunctionTools("Qwen3-0.6B-Q8_0")
    }

    func testLLMFunctionTools(_ modelName: String) async throws {
        let llmLlama = try await loadModel(name: modelName)

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
        let userMsg = "What is the weather like in Boston today?"
        var messages = [ChatMessage(role: .user, content: userMsg)]
        let options = GenerationOptions(templeteOptions: .init(tools: weatherTool, enableThinking: true))

        let stream = try await llmLlama.generationAsyncStream(messages: messages, options: options)
        print("-----------------------------")
        print("User: ", userMsg)
        print("AI: ")
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
