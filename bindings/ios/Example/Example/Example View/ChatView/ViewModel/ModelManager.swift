import Foundation
import NexaAI

@Observable
@MainActor
class ModelManager {

    private(set) var model: Model?
    private var isGenerating: Bool = false
    var isLoadingModel: Bool = false

    private(set) var forceRefesh: Bool = false

    init(model: Model? = nil) {
        self.model = model
    }

    var isLoaded: Bool {
        model?.isLoaded ?? false
    }

    func reload() {
        forceRefesh.toggle()
    }

    func unload() async {
        self.model = nil
    }

    func load(modelType: NexaAI.ModelType, options: ModelOptions) async throws {
        isLoadingModel = true
        defer { isLoadingModel = false }
        await unload()
        self.model = modelType == .llm ? LLMLlama() : VLMLlama()
        try await self.model?.load(options)
    }

    func stopGeneration() {
        model?.stopStream()
    }

    func generationStream(messages: [Message], options: GenerationOptions = .init()) async throws -> String {
        guard let model else {
            throw ModelError.notLoad
        }
        var msgs = [ChatMessage]()
        msgs.append(.init(role: .system, content: defaultSystemPrompt))
        msgs.append(contentsOf: messages.map {
            let images = $0.images.map { $0.path() }
            let audios = $0.audios.map { $0.path() }
            return ChatMessage(role: $0.isUser ? .user : .assistant, content: $0.content, images: images, audios: audios)
        })
        let result = try await model.generationStream(messages: msgs, options: options)
        return result.response
    }

    func generationAsyncStream(messages: [Message], options: GenerationOptions = .init()) async throws -> AsyncThrowingStream<String, any Error> {
        guard let model else {
            throw ModelError.notLoad
        }
        var msgs = [ChatMessage]()
        msgs.append(.init(role: .system, content: defaultSystemPrompt))
        msgs.append(contentsOf: messages.map {
            let images = $0.images.map { $0.path() }
            let audios = $0.audios.map { $0.path() }
            return ChatMessage(role: $0.isUser ? .user : .assistant, content: $0.content, images: images, audios: audios)
        })
        return try await model.generationAsyncStream(messages: msgs, options: options)
    }

    func reset() async {
        stopGeneration()
        await model?.reset()
    }

    func getProfileData() async -> ProfileData? {
        await model?.lastProfileData
    }

    enum ModelError: LocalizedError {
        case notLoad
    }

    let defaultSystemPrompt = """
            You are a helpful, concise, and privacy-respecting AI assistant running fully on-device. Provide accurate, unbiased answers across a wide range of topics. When unsure, state so clearly. Avoid speculation. Always prioritize clarity, relevance, and user control.
        """
}
