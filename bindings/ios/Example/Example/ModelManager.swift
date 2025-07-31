
import SwiftData
import Foundation
import NexaAI

@Observable
@MainActor
class ModelManager {

    private(set) var model: Model?

    var isGenerating: Bool = false
    var isLoadingModel: Bool = false
    var generationConfig: GenerationConfig = .default
    var systemPrompt: String = "You are a helpful assistant"

    func unload() async {
        await self.model?.unload()
        self.model = nil
    }

    func loadModel(_ model: Model) async throws {
        isLoadingModel = true
        defer { isLoadingModel = false }
        await self.model?.unload()
        self.model = model
        try await self.model?.loadModel()
    }

    func stopGeneration() {
        isGenerating = false
    }

    private func startGeneration() {
        isGenerating = true
    }

    func generationAsyncStream(from message: Message) async throws -> AsyncThrowingStream<String, any Error> {
        guard let model else {
            throw ModelError.notLoad
        }
        var config = generationConfig
        config.imagePaths = message.images.map { $0.path() }
        config.audioPaths = message.audios.map { $0.path() }
        let prompt = message.content

        return .init { continuation in
            Task {
                startGeneration()
                do {
                    _ = try await model.generationStream(prompt: prompt, config: config) { [weak self] token in
                        guard let self else { return false }
                        continuation.yield(token)
                        return self.isGenerating
                    }
                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
                stopGeneration()
            }
        }
    }

    func generate(from messages: [Message]) async throws -> String {
        guard let model else {
            throw ModelError.notLoad
        }

        var msgs = [ChatMessage]()

        let type = await model.type
        let config = generationConfig
        let systemPrompt = systemPrompt
        if type == .llm, !systemPrompt.isEmpty {
            msgs.append(.init(role: .system, content: systemPrompt))
        }

        msgs.append(contentsOf: messages.map {
            ChatMessage(role: $0.isUser ? .user : .assistant, content: $0.content)
        })

        startGeneration()
        let prompt = try await model.applyChatTemplate(messages: msgs)
        let result = try await model.generate(prompt: prompt, config: config)
        stopGeneration()
        return result
    }

    func generationAsyncStream(from messages: [Message]) async throws -> AsyncThrowingStream<String, any Error> {
        guard let model else {
            throw ModelError.notLoad
        }

        var msgs = [ChatMessage]()

        let type = await model.type
        let config = generationConfig
        let systemPrompt = systemPrompt
        if type == .llm, !systemPrompt.isEmpty {
            msgs.append(.init(role: .system, content: systemPrompt))
        }

        msgs.append(contentsOf: messages.map {
            ChatMessage(role: $0.isUser ? .user : .assistant, content: $0.content)
        })

        return .init { continuation in
            Task {
                startGeneration()
                do {
                    let prompt = try await model.applyChatTemplate(messages: msgs)
                    _ = try await model.generationStream(prompt: prompt, config: config) { [weak self] token in
                        guard let self else { return false }
                        continuation.yield(token)
                        return self.isGenerating
                    }
                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
                stopGeneration()
            }
        }
    }

    func reset() async {
        stopGeneration()
        await model?.reset()
    }

    func getProfileData() async -> ProfilingData? {
        await model?.getProfileData()
    }

    enum ModelError: LocalizedError {
        case notLoad
    }
}
