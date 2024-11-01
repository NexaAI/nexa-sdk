import Foundation
import NexaSwift
import SwiftUI
import Combine

@Observable
class ViewModel {
    let nexaSwift: NexaTextInference
    var result = ""
    var usingStream = true
    private var cancallable: Set<AnyCancellable> = []
    
    init() {
        let configuration = Configuration(maxNewToken: 128, stopTokens: ["<nexa_end>"])
        let path = Bundle.main.path(forResource: "octopusv2_q4_0", ofType: "gguf") ?? ""
        nexaSwift = (try? NexaTextInference(modelPath: path, modelConfiguration: configuration))!
    }

    func formatUserMessage(_ message: String) -> String {
        let formatted = """
        Below is the query from the users, please call the correct function and generate the parameters to call the function.
        
        Query: \(message)
        
        Response:
        """
        return formatted
    }

    func run(for userMessage: String) {
        result = ""
        
        let formattedUserMessage = formatUserMessage(userMessage)

        Task {
            switch usingStream {
            case true:
                for try await value in await nexaSwift.createCompletionStream(for: formattedUserMessage) {
                    print("Received content: \(value.choices[0].text)")   // DEBUG
                    result += value.choices[0].text
                }
            case false:
                if let completionResponse = try? await nexaSwift.createCompletion(for: formattedUserMessage) {
                    print("Received completion response: \(completionResponse.choices[0].text)")   // DEBUG
                    result += completionResponse.choices[0].text
                }
            }
        }
    }
}
