import Foundation
import NexaSwift
import SwiftUI
import Combine

@Observable
class ViewModel {
    let nexaSwift: NexaTextInference
    var result = ""
    var usingStream = true
    private var messages: [ChatCompletionRequestMessage] = []
    private let maxHistory = 1
    private var cancallable: Set<AnyCancellable> = []
    
    init() {
        let configuration = Configuration(maxNewToken: 128, stopTokens: [])
        let model_path = Bundle.main.path(forResource: "llama3_2_3b_q4_K_M", ofType: "gguf") ?? ""
        nexaSwift = (try? NexaTextInference(modelPath: model_path, modelConfiguration: configuration))!
    }

    func run(for userMessage: String) {
        result = ""
        let userMessageText = ChatCompletionRequestMessage.user(
            ChatCompletionRequestUserMessage(content: .text(userMessage))
        )

        messages.append(userMessageText)
        if messages.count > maxHistory * 2 {
            messages.removeFirst(2)
        }

        Task {
            switch usingStream {
            case true:
                for try await value in await nexaSwift.createChatCompletionStream(for: messages) {
                    let delta = value.choices[0].delta.content ?? ""
                    result += delta
                }
            case false:
                if let completionResponse = try? await nexaSwift.createChatCompletion(for: messages) {
                    let content = completionResponse.choices[0].message.content ?? ""
                    result += content
                }
            }

            // Add assistant's response to history
            let assistantMessage = ChatCompletionRequestMessage.assistant(
                ChatCompletionRequestAssistantMessage(
                    content: result,
                    toolCalls: nil,
                    functionCall: nil
                )
            )
            messages.append(assistantMessage)
        }
    }
}
