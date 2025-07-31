import Foundation
import Testing

@testable import NexaAI

@Test func llmTest() async throws {
    print("===> BEGIN LLM TEST")

    let modelPath = try resoucePath(of: "Qwen3-0.6B-Q8_0")
    let llm = LLM(modelPath: modelPath)
    try llm.loadModel()

    let prompt = "Once upon a time"

    try testTokenizer(llm, prompt)
    try testKVCache(llm, "./kvcache")
    try testSampler(llm)
    try await testGenerate(llm, prompt)
    try testGetChatTemplete(llm, nil)

    try testEmbedding(llm, prompt)
    try await testApplyChatTemplate(llm)
    try await testGenerationStream(llm, prompt)

    print("===> END LLM TEST")
}

func testGenerationStream(_ llm: LLM, _ prompt: String) async throws {
    print("===> test Generation Stream")
    let streamText = try await llm.generationStream(
        prompt: prompt,
        config: .init(),
        onToken: { token in
            print(token, terminator: "")
            return true
        }
    )
    print("full text: \(streamText)")

}

func testEmbedding(_ llm: LLM, _ prompt: String) throws {
    print("===> test Embedding")
    let embedResult = try llm.embed(texts: [prompt])
    if embedResult.count > 32 {
        print(embedResult.prefix(32), "... (\(embedResult.count - 32) more)")
    } else {
        print(embedResult)
    }
}

func testApplyChatTemplate(_ llm: LLM) async throws {
    print("===> test Apply Chat Template")
    let messages: [ChatMessage] = [
        ChatMessage(role: .system, content: "You are a helpful assistant."),
        ChatMessage(role: .user, content: "What is the capital of France?"),
    ]
    let applyChatResult = try await llm.applyChatTemplate(messages: messages)
    print(applyChatResult)

    let genConfig = GenerationConfig()
    let chatResult = try await llm.generate(prompt: applyChatResult, config: genConfig)
    print(chatResult)
}

func testGetChatTemplete(_ llm: LLM, _ name: String?) throws {
    print("===> test Generate")
    let templete = try llm.getChatTemplate(name: name)
    print(templete)
}

func testGenerate(_ llm: LLM, _ prompt: String) async throws {

    print("===> test Generate")
    var config = GenerationConfig.default
    config.maxTokens = 32
    let out = try await llm.generate(prompt: prompt, config: config)
    print(out)
    llm.reset()
}

func testSampler(_ llm: LLM) throws {
    print("===> test Sampler")
    try llm.setSampler(config: .init(temperature: 0.8, topP: 0.9, topK: 40, repetitionPenalty: 1.0, presencePenalty: 0.0, seed: 42))
    llm.reset()
}

func testKVCache(_ llm: LLM, _ path: String) throws {

    print("===> Save KVCache")
    try llm.saveKVCache(to: path)

    print("===> Load KVCache")
    try llm.loadKVCache(from: path)
}

func testTokenizer(_ llm: LLM, _ prompt: String) throws {

    print("===> test encode/decode")
    let tokens = try llm.encode(text: prompt)
    print("encode tokens: \(tokens)")

    let decodeText = try llm.decode(tokens: tokens)
    #expect(decodeText == prompt)
}


@Test func testChatMultiRound() async throws {
    let modelPath = try resoucePath(of: "Qwen3-4B-Q4_K_M.F32")

    let modelConfig = ModelConfig.default
    let llm = LLM(modelPath: modelPath, config: modelConfig)
    try llm.loadModel()
    
    let conversations = [
        "1+1等于几？",
        "那2+2呢？",
        "n+n呢？"
    ]

    var messages = [ChatMessage]()
    var processedTokens: Int32 = 0
    for conversation in conversations {
        let message = ChatMessage(role: .user, content: conversation)

        print("User: ")
        print("---: ", conversation)
        messages.append(message)
        let prompt = try await llm.applyChatTemplate(messages: messages)

        var config = GenerationConfig()
        config.maxTokens = 64
        config.nPast = processedTokens
        let assistaint = try await llm.generationStream(prompt: prompt, config: config) { _ in
            return true
        }
        print("Assistant:")
        print("---: ", assistaint)

        let tokens = try llm.encode(text: prompt)
        processedTokens += Int32(tokens.count)

        let assistantMessage = ChatMessage(role: .assistant, content: assistaint)
        messages.append(assistantMessage)
    }

    llm.reset()
    print(llm.getProfileData() ?? "")
}
