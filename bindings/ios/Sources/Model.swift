import Foundation

public enum ModelType: CaseIterable {
    case llm
    case vlm
}

@NexaAIActor
public protocol Model {
    func loadModel() throws
    func unload()

    func applyChatTemplate(messages: [ChatMessage]) throws -> String

    func generate(prompt: String, config: GenerationConfig) async throws -> String
    func generationStream(prompt: String, config: GenerationConfig, onToken: @escaping (String) -> Bool) async throws -> String

    func getProfileData() -> ProfilingData?

    func reset()
    
    var type: ModelType { get }
}

extension Model {
    
    public func generationAsyncStream(from messages: [ChatMessage], config: GenerationConfig) -> AsyncThrowingStream<String, any Error> {
        return .init { continuation in
            Task {
                do {
                    let prompt = try applyChatTemplate(messages: messages)
                    _ = try await generationStream(prompt: prompt, config: config) { token in
                        continuation.yield(token)
                        return true
                    }
                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
            }
        }
    }
}
