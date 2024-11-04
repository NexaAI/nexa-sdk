import Foundation

public struct CompletionUsage: Codable {
    public let promptTokens: Int
    public let completionTokens: Int
    public let totalTokens: Int
    
    enum CodingKeys: String, CodingKey {
        case promptTokens = "prompt_tokens"
        case completionTokens = "completion_tokens"
        case totalTokens = "total_tokens"
    }
}

public struct CompletionLogprobs: Codable {
    public let textOffset: [Int]?
    public let tokenLogprobs: [Float?]?
    public let tokens: [String]?
    public let topLogprobs: [Dictionary<String, Float>?]?
}

public struct CompletionChoice: Codable {
    public let text: String
    public let index: Int
    public let logprobs: CompletionLogprobs?
    public let finishReason: FinishReason?
    
    enum CodingKeys: String, CodingKey {
        case text
        case index
        case logprobs
        case finishReason = "finish_reason"
    }
}

public struct CompletionResponse: Codable {
    public let id: String
    public let object: String
    public let created: Int
    public let model: String
    public let choices: [CompletionChoice]
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
