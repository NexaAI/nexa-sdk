import Foundation


public struct ChatCompletionMessageToolCallFunction: Codable {
    public let name: String
    public let arguments: String

    public init(name: String, arguments: String) {
        self.name = name
        self.arguments = arguments
    }
}

public struct ChatCompletionMessageToolCall: Codable {
    public let id: String
    public var type: Role = .function
    public let function: ChatCompletionMessageToolCallFunction

    public init(id: String, function: ChatCompletionMessageToolCallFunction) {
        self.id = id
        self.function = function
    }
}

public struct ChatCompletionResponseFunctionCall:Codable{
    public let name: String
    public let arguments: String
}

public struct ChatCompletionResponseMessage: Codable{
    public let content: String?
    public let toolCalls: [ChatCompletionMessageToolCall]?
    public let role: String?
    public let functionCall: ChatCompletionResponseFunctionCall?
}

public struct  ChatCompletionResponseChoice: Codable{
    public let index: Int
    public let message: ChatCompletionResponseMessage
    public let logprobs: CompletionLogprobs?
    public let finishReason: FinishReason?
}

public struct ChatCompletionResponse: Codable {
    public let id: String
    public let object: String
    public let created: Int
    public let model: String
    public let choices: [ChatCompletionResponseChoice]
    public let usage: CompletionUsage?

    enum CodingKeys: String, CodingKey {
        case id
        case object
        case created
        case model
        case choices
        case usage
    }
}

public struct ChatCompletionStreamResponseDelta: Codable {
    public var content: String?
    public var functionCall: ChatCompletionStreamResponseDeltaFunctionCall? // DEPRECATED
    public var toolCalls: [ChatCompletionMessageToolCallChunk]?
    public var role: Role?
    
}

public struct ChatCompletionStreamResponseDeltaFunctionCall: Codable {

}

public struct ChatCompletionMessageToolCallChunk: Codable {

}

public struct ChatCompletionStreamResponseChoice: Codable {
    public var index: Int
    public var delta: ChatCompletionStreamResponseDelta
    public var finishReason: FinishReason?
    public var logprobs: CompletionLogprobs?
}

public struct CreateChatCompletionStreamResponse: Codable {
    public var id: String
    public var model: String
    public var object: String
    public var created: Int
    public var choices: [ChatCompletionStreamResponseChoice]
}
