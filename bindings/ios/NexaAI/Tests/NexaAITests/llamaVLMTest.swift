import Testing
import Foundation

@testable import NexaAI

struct LLamaVLMTest {

    init() {
        NexaSdk.install([.error])
    }

    func loadModel(
        name: String = "SmolVLM-256M-Instruct-Q8_0",
        projectFileName: String = "mmproj-SmolVLM-256M-Instruct-Q8_0"
    ) async throws -> VLMLlama {
        let vlmLlama = VLMLlama()
        let projectPath = try modelPath(of: projectFileName)
        let modelPath = try modelPath(of: name)
        try await vlmLlama.load(.init(modelPath: modelPath, mmprojPath: projectPath))
        return vlmLlama
    }

    @Test func testLoadModel() async throws {
        let _ = try await loadModel()
    }

    @Test func generateStream() async throws {
        let vlmLlama = try await loadModel()
        let config = GenerationConfig(maxTokens: 32)
        let stream = await vlmLlama.generateAsyncStream(prompt: "Tell me a story about 100 words", config: config)

        for try await token in stream {
            print(token, terminator: "")
        }
        print("\n")
        print(await vlmLlama.lastProfileData ?? "")
    }

    @Test func testGenerateChatMultiRound() async throws {
        let vlmLlama = try await loadModel()

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
            let stream = try await vlmLlama.generateAsyncStream(messages: messages)
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
            print(await vlmLlama.lastProfileData?.description ?? "")
        }
    }

    @Test func testVLMImage() async throws {
        let vlmLlama = try await loadModel()

        let images = [try assetsPath(of: "test_image.png")]

        var config = GenerationConfig.default
        config.imagePaths = images
        let message = ChatMessage(role: .user, content: "What do you see in this image", images: images)
        let stream = try await vlmLlama.generateAsyncStream(messages: [message], options: .init(config: config))
        print("--------------------")
        for try await token in stream {
            print(token, terminator: "")
        }
        print("\n")
        print(await vlmLlama.lastProfileData?.description ?? "")
        print("--------------------")
    }
}
