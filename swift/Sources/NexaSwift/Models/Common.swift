public enum Role: String, Codable {
    case system
    case user
    case assistant
    case tool
    case function
}

public enum ChatCompletionRequestMessage {
    case system(ChatCompletionRequestSystemMessage)
    case user(ChatCompletionRequestUserMessage)
    case assistant(ChatCompletionRequestAssistantMessage)
    case tool(ChatCompletionRequestToolMessage)
    case function(ChatCompletionRequestFunctionMessage)
}

public enum FinishReason: String, Codable {
        case stop, length, toolCalls = "tool_calls", functionCall = "function_call"
}

public enum ChatCompletionModel: String, Codable {
    case octopusv2
    case llama
    case llama3
    case gemma
    case qwen
    case mistral
}
