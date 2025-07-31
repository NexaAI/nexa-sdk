import Foundation
import Testing
@testable import NexaAI

@Test func vlmTest() async throws {
    // Write your test here and use APIs like `#expect(...)` to check expected conditions.
//    let modelPath = try resoucePath(of: "Qwen3-0.6B-Q8_0")
//    let mmprojPath = try resoucePath(of: "mmproj-model-f16")

    let modelPath = try resoucePath(of: "SmolVLM-256M-Instruct-f16")
    let mmprojPath = try resoucePath(of: "mmproj-SmolVLM-256M-Instruct-f16")

    let imagePath = try resoucePath(of: "test1", ext: "jpg")
    let audioPath = try resoucePath(of: "test1", ext: "jpg")
    let promt = "Describe this image"

    print("test create")
    let vlm = VLM(modelPath: modelPath, mmprojPath: mmprojPath)
    try vlm.loadModel()

    print("test encode")
    let text = "Describe this image"
    let tokens = try vlm.encode(text: text)
    print("Encoded tokens: \(tokens)")

    print("test decode")
    let decodedText = try vlm.decode(tokens: tokens)
    print("Decoded text: \(decodedText)")

    print("test get/reset sampler")
    try vlm.setSampler(config: .default)
    try vlm.resetSampler()

    print("test generate")
    var generationConfig = GenerationConfig()
    generationConfig.maxTokens = 32
    generationConfig.imagePaths = [imagePath]
    generationConfig.audioPaths = [audioPath]
    let generatedText = try await vlm.generate(prompt: text, config: generationConfig)
    print("Generated text: \(generatedText)")

    print("test get chat template")
    let chatTemplate = try vlm.getChatTemplate()
    print("Chat template: \n \(chatTemplate)")

    print("test chat")
    let messages: [ChatMessage] = [
        ChatMessage(role: .system, content: "You are a helpful assistant."),
        ChatMessage(role: .user, content: promt),
    ]
    let applyChatResult = try await vlm.applyChatTemplate(messages: messages)
    print(applyChatResult)

    let applyGeneratedText = try await vlm.generate(prompt: applyChatResult, config: generationConfig)
    print("apply generated text: \(applyGeneratedText)")

    print("test embed")
    let embedding = try vlm.embed(texts: [promt])
    print(
        "Embedding vector (first 5 elements),\(embedding.prefix(5)), total count: \(embedding.count)"
    )

    print("test generate stream")
//    generationConfig.imagePaths = []
    let streamText = try await vlm.generationStream(
        prompt: text,
        config: generationConfig,
        onToken: { token in
            if token.isEmpty {
                print(token, terminator: "")
                fflush(stdout)
            } else {
                print("Stream ended")
            }
            return true
        }
    )
    print("\nFull text: \(streamText)")

}

@Test func testImageGenerate() async throws {
//    let modelPath = try resoucePath(of: "Qwen2-VL-2B-Instruct-Q4_K_M")
//    let mmprojPath = try resoucePath(of: "mmproj-Qwen2-VL-2B-Instruct-Q8_0")

    let modelPath = try resoucePath(of: "SmolVLM-256M-Instruct-f16")
    let mmprojPath = try resoucePath(of: "mmproj-SmolVLM-256M-Instruct-f16")
    let imagePath = try resoucePath(of: "test1", ext: "jpg")
    let promt = "Describe this image"

    let vlm = VLM(modelPath: modelPath, mmprojPath: mmprojPath)
    try vlm.loadModel()

    var generationConfig = GenerationConfig()
    generationConfig.maxTokens = 4096
    generationConfig.imagePaths = [imagePath]
    let generatedText = try await vlm.generate(prompt: promt, config: generationConfig)
    print(generatedText)

    generationConfig.imagePaths = [imagePath, imagePath]
    let generatedTextRound2 = try await vlm.generate(prompt: "Compare this two images", config: generationConfig)
    print(generatedTextRound2)
}

@Test func vlmAudioGeneration() async throws {
    let modelPath = try resoucePath(of: "Qwen3-4B-Q4_K_M.F32")
    let mmprojPath = try resoucePath(of: "mmproj-model-f16")
    let audioPath = try resoucePath(of: "test-2", ext: "mp3")
    let prompt = "Translate this audio to English"

    let vlm = VLM(modelPath: modelPath, mmprojPath: mmprojPath)
    try vlm.loadModel()

    var config = GenerationConfig()
    config.maxTokens = 32
    config.audioPaths = [audioPath]

    let result = try await vlm.generationStream(prompt: prompt, config: config) { token in
        return true
    }
    print(result)
}

@Test func testError() {
    let loadError = VLM.VLMError.modelLoadingFailed
    let failError = VLM.VLMError.generationStreamFailed(-300101)

    print(loadError.localizedDescription)
    print(failError.localizedDescription)
}
