import SwiftUI
import Observation
import Foundation
import NexaAI

@Observable
@MainActor
class ChatViewModel {
    let modelManager: ModelManager = .init()

    private var startIndex: Int = 0
    private let maxMessageCount: Int = 20
    private(set) var messages: [Message] = []
    private(set) var currentGenerateMessge: Message?

    var scrollPosition: String?

    var isGenerating: Bool = false
    var isLoadingModel: Bool {
        get { modelManager.isLoadingModel }
    }
    var modelLoadProgress: Float = 0
    var generationError: Error?
    var isLoadModelError: Bool = false

    var currentEditingMessage: Message?
    var prompt: String = ""
    var modelName: String = ""

    func loadModel(from modelPath: String) async {
        do {
            willLoadModel()
            let options = ModelOptions(
                modelPath: modelPath,
                config: .init(nCtx: 4096),
                deviceId: nil
            )
            try await modelManager.load(modelType: .llm, options: options)
            didLoadModel()
        } catch {
            isLoadModelError = true
        }
    }

    func willLoadModel() {
        Task {
            modelLoadProgress = 0.0
            while modelLoadProgress < 0.95 {
                let increment = Float.random(in: 0.1...0.3)
                withAnimation(.smooth) {
                    modelLoadProgress = min(modelLoadProgress + increment, 0.95)
                }
                try? await Task.sleep(for: .seconds(0.1))
            }
        }
    }

    func didLoadModel() {
        withAnimation(.smooth) {
            modelLoadProgress = 1.0
        }
        Task {
            try? await Task.sleep(for: .seconds(0.02))
            modelLoadProgress = 0.0
        }
    }

    func regerationStream(from index: Int) {
        if isGenerating {
            return
        }

        if index >= messages.count {
            return
        }

        if !modelManager.isLoaded {
            isLoadModelError = true
            return
        }

        Task {
            let isUser = messages[index].isUser
            let userMessageIndex = isUser ? index : index - 1
            let userMessage = messages[userMessageIndex]
            do {
                messages = Array(messages[0...userMessageIndex])
                await modelManager.reset()
                scrollPosition = userMessage.id
                startIndex = max(messages.count - 1, 0)
                try await _generationStream()
            } catch {
                endGenerating(with: error)
            }
        }
    }

    func generationStream() {
        if !modelManager.isLoaded {
            isLoadModelError = true
            return
        }
        let content = prompt.trimmingCharacters(in: .whitespacesAndNewlines)
        prompt = ""
        Task {
            do {
                let message = Message.user(content)
                messages.append(message)
                scrollPosition = message.id
                try await _generationStream()
            } catch {
                endGenerating(with: error)
            }
        }
    }

    private func _generationStream() async throws {
        beginGenerating()

        let genMsgs: [Message]
        if startIndex >= messages.count {
            genMsgs = messages.suffix(maxMessageCount)
        } else {
            genMsgs = Array(messages[startIndex...])
        }

        let options = GenerationOptions(config: .init(maxTokens: 4096), templeteOptions: .init(enableThinking: false))
        let stream = try await modelManager.generationAsyncStream(messages: genMsgs, options: options)

        var assistant = Message.assistant("")
        currentGenerateMessge = assistant
        var tokens = 0
        for try await value in stream {
            assistant.content += value
            tokens += 1
            if isGenerating, tokens == 3 {
                currentGenerateMessge = assistant
                tokens = 0
            }
        }
        if isGenerating {
            currentGenerateMessge = assistant
        }

        messages.append(assistant)

        endGenerating()
    }

    private func beginGenerating() {
        isGenerating = true
        generationError = nil
    }

    private func endGenerating() {
        currentGenerateMessge = nil
        isGenerating = false
        scrollPosition = nil
    }

    private func endGenerating(with error: Error) {
        currentGenerateMessge = nil
        isGenerating = false
        generationError = error
    }
}
