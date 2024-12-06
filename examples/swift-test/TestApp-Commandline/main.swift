import Foundation
import NexaSwift

let configuration = NexaSwift.Configuration(
    maxNewToken: 128,
    stopTokens: []
)

let model_path = "path/to/your/model" // For Commandline, please add the local path here. 
let nexaSwift = try NexaSwift.NexaTextInference(modelPath: model_path, modelConfiguration: configuration)

var streamMode = false
print("Do you want to enable stream mode? (yes/y or no/n):", terminator: " ")
var userInput = readLine()?.lowercased() ?? ""
if userInput == "yes" || userInput == "y" {
    streamMode = true
}
print("")

var messages:[ChatCompletionRequestMessage] = []
let maxHistory = 2

while true {
    print("You:", terminator: " ")
    userInput = readLine() ?? ""
    print("Bot:", terminator: " ")
    
    let userMessageText = ChatCompletionRequestMessage.user(
        ChatCompletionRequestUserMessage(content: .text(userInput))
    )
    
    messages.append(userMessageText)
    if messages.count > maxHistory * 2 {
        messages.removeFirst(2)
    }
    
    var currentMessage = ""
    if streamMode{
        for try await value in await nexaSwift
            .createChatCompletionStream(for: messages) {
            print(value.choices[0].delta.content ?? "", terminator: "")
            currentMessage += value.choices[0].delta.content ?? ""
        }
    }else{
        let response = try await nexaSwift.createChatCompletion(for: messages)
        print(response.choices[0].message.content ?? "", terminator: "")
        currentMessage += response.choices[0].message.content ?? ""
    }
    
    
    let assistantMessage = ChatCompletionRequestMessage.assistant(
        ChatCompletionRequestAssistantMessage(
            content: currentMessage,
            toolCalls: nil,
            functionCall: nil
        )
    )
    
    messages.append(assistantMessage)
    
    print("")
}
