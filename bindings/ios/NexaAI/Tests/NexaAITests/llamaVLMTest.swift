import Testing
import Foundation

@testable import NexaAI

struct LLamaVLMTest {

    init() {
        NexaSdk.install([])
    }
    @Test func testLLamaVLMLoad() async throws {
        let vlmLlama = VLMLlama()
        let projectPath = try modelPath(of: "mmproj-SmolVLM-256M-Instruct-Q8_0")
        let modelPath = try modelPath(of: "SmolVLM-256M-Instruct-Q8_0")
        try await vlmLlama.load(.init(modelPath: modelPath, mmprojPath: projectPath))
    }

    @Test func testLlamaVLMGenerationAsyncStream() async throws {
        let vlmLlama = VLMLlama()
        let projectPath = try modelPath(of: "mmproj-SmolVLM-256M-Instruct-Q8_0")
        let modelPath = try modelPath(of: "SmolVLM-256M-Instruct-Q8_0")
        try await vlmLlama.load(.init(modelPath: modelPath, mmprojPath: projectPath))

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
            let stream = try await vlmLlama.generationAsyncStream(messages: messages)
            print("-----------------------------")
            var response = ""
            for try await token in stream {
                print(token, terminator: "")
                response += token
            }
            print("\n")
            messages.append(.init(role: .assistant, content: response))
            print(await vlmLlama.lastProfileData?.description ?? "")
            print("-----------------------------")
        }
    }

    @Test func testVLMImage() async throws {
        let vlmLlama = VLMLlama()
        let projectPath = try modelPath(of: "mmproj-SmolVLM-256M-Instruct-Q8_0")
        let modelPath = try modelPath(of: "SmolVLM-256M-Instruct-Q8_0")
        try await vlmLlama.load(.init(modelPath: modelPath, mmprojPath: projectPath))

        let images = [try assetsPath(of: "test_image.png")]

        var config = GenerationConfig.default
        config.imagePaths = images
        let message = ChatMessage(role: .user, content: "What do you see in this image", images: images)
        let stream = try await vlmLlama.generationAsyncStream(messages: [message], options: .init(config: config))
        print("--------------------")
        for try await token in stream {
            print(token, terminator: "")
        }
        print("\n")
        print(await vlmLlama.lastProfileData?.description ?? "")
        print("--------------------")
    }


    @Test func testLlamaVLMFunctionTools() async throws {
        let vlmLlama = VLMLlama()
        let projectPath = try modelPath(of: "mmproj-SmolVLM-256M-Instruct-Q8_0")
        let modelPath = try modelPath(of: "SmolVLM-256M-Instruct-Q8_0")
        try await vlmLlama.load(.init(modelPath: modelPath, mmprojPath: projectPath))

//        {"type":"function","function":{"name":"get_current_weather","description":"Get the current weather in a given location","parameters":{"type":"object","properties":{"location":{"type":"string","description":"The city and state, e.g. San Francisco, CA"},"unit":{"type":"string","enum":["celsius","fahrenheit"]}},"required":["location"]}}}

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

        let system = "You are a helpful, concise, and privacy-respecting AI assistant running fully on-device. Provide accurate, unbiased answers across a wide range of topics. When unsure, state so clearly. Avoid speculation. Always prioritize clarity, relevance, and user control."
        let userMsgs = [
            "What is the weather like in Boston today?"
        ]
        let options = GenerationOptions(templeteOptions: .init(tools: weatherTool))
        var messages = [ChatMessage]()
        messages.append(.init(role: .system, content: system))
        for userMsg in userMsgs {
            messages.append(.init(role: .user, content: userMsg))
            let stream = try await vlmLlama.generationAsyncStream(messages: messages, options: options)
            print("-----------------------------")
            var response = ""
            for try await token in stream {
                print(token, terminator: "")
                response += token
            }
            print("\n")
            messages.append(.init(role: .assistant, content: response))
            print(await vlmLlama.lastProfileData?.description ?? "")
            print("-----------------------------")
        }
    }
}
