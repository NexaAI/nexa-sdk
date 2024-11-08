import Foundation


public struct ChatCompletionRequestSystemMessage: Codable {
    public var role: Role = .system
    public let content: String?
    
    public init(content: String?) {
        self.content = content
    }
}

public struct ChatCompletionRequestUserMessage: Codable {
    public var role: Role = .user
    public let content: UserMessageContent
    
    public init(content: UserMessageContent) {
        self.content = content
    }
}

public enum UserMessageContent: Codable {
    case text(String)
    case image(ImageContent)
    
    enum CodingKeys: String, CodingKey {
        case type, text, imageUrl
    }

    enum ContentType: String, Codable {
        case text
        case imageUrl
    }

    public init(from decoder: Decoder) throws {
        let container = try decoder.container(keyedBy: CodingKeys.self)
        let type = try container.decode(ContentType.self, forKey: .type)
        
        switch type {
        case .text:
            let text = try container.decode(String.self, forKey: .text)
            self = .text(text)
        case .imageUrl:
            let imageUrl = try container.decode(ImageContent.self, forKey: .imageUrl)
            self = .image(imageUrl)
        }
    }

    public func encode(to encoder: Encoder) throws {
        var container = encoder.container(keyedBy: CodingKeys.self)
        switch self {
        case .text(let text):
            try container.encode(ContentType.text, forKey: .type)
            try container.encode(text, forKey: .text)
        case .image(let imageUrl):
            try container.encode(ContentType.imageUrl, forKey: .type)
            try container.encode(imageUrl, forKey: .imageUrl)
        }
    }
}

public struct ImageContent: Codable {
    public let url: String
    public let detail: String?
    
    public init(url: String, detail: String? = nil) {
        self.url = url
        self.detail = detail
    }
}

public struct ChatCompletionRequestAssistantMessage: Codable {
    public var role: Role = .assistant
    public let content: String?
    public let toolCalls: [ChatCompletionMessageToolCall]?
    public let functionCall: ChatCompletionRequestAssistantMessageFunctionCall?
    
    public init(content: String?, toolCalls: [ChatCompletionMessageToolCall]? = nil, functionCall: ChatCompletionRequestAssistantMessageFunctionCall? = nil) {
        self.content = content
        self.toolCalls = toolCalls
        self.functionCall = functionCall
    }
}

public struct ChatCompletionRequestToolMessage: Codable {
    public var role: Role = .tool
    public let content: String?
    public let toolCallID: String
    
    public init(content: String?, toolCallID: String) {
        self.content = content
        self.toolCallID = toolCallID
    }
}

public struct ChatCompletionRequestFunctionMessage: Codable {
    public var role: Role = .function
    public let content: String?
    public let name: String
    
    public init(content: String?, name: String) {
        self.content = content
        self.name = name
    }
}

public struct ChatCompletionRequestAssistantMessageFunctionCall: Codable {
    public let name: String
    public let arguments: String
    
    public init(name: String, arguments: String) {
        self.name = name
        self.arguments = arguments
    }
}


class ChatFormatterRegistry {
    private var formatters = [String: ChatFormatter]()
    
    init() {
        register(name: ChatCompletionModel.octopusv2.rawValue, formatter: OctopusV2Formatter())
        register(name: ChatCompletionModel.llama.rawValue, formatter: LlamaFormatter())
        register(name: ChatCompletionModel.llama3.rawValue, formatter: Llama3Formatter())
        register(name: ChatCompletionModel.gemma.rawValue, formatter: GemmaFormatter())
        register(name: ChatCompletionModel.qwen.rawValue, formatter: QwenFormatter())
        register(name: ChatCompletionModel.mistral.rawValue, formatter: MistralFormatter())
    }
    
    func register(name: String, formatter: ChatFormatter) {
        formatters[name] = formatter
    }
    
    func getFormatter(name: String?) -> ChatFormatter? {
        return formatters[getFormatterName(name: name)]
    }
    
    func getFormatterName(name: String?) -> String {
        return name ?? ChatCompletionModel.llama.rawValue
    }
}

//formatter
public struct ChatFormatterResponse {
    let prompt: String
    let stop: [String]
}

public protocol ChatFormatter {
    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse
}


class OctopusV2Formatter: ChatFormatter {
    private let systemMessage = """
    Below is the query from the users, please call the correct function and generate the parameters to call the function.

    """
    private let separator = "\n\n"

    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse {
        var formattedMessages = mapRoles(messages: messages)
        
        // Assuming the last message should be the assistant's response
        formattedMessages.append(("Response:", nil))
        
        var prompt = systemMessage
        for (role, content) in formattedMessages {
            if let content = content {
                prompt += "\(role) \(content.trimmingCharacters(in: .whitespacesAndNewlines))\(separator)"
            } else {
                prompt += "\(role) "
            }
        }
        
        return ChatFormatterResponse(prompt: prompt.trimmingCharacters(in: .whitespacesAndNewlines), stop: [separator])
    }
    
    private func mapRoles(messages: [ChatCompletionRequestMessage]) -> [(String, String?)] {
        var mappedMessages = [(String, String?)]()
        let roleMapping: [Role: String] = [
            .user: "Query:",
            .assistant: "Response:"
        ]
        
        for message in messages {
            var rolePrefix = ""
            var content: String? = nil
            
            switch message {
            case .system(let systemMessage):
                // Include system message if necessary
                continue
            case .user(let userMessage):
                rolePrefix = roleMapping[.user] ?? "Query:"
                switch userMessage.content {
                case .text(let text):
                    content = text
                case .image(let imageContent):
                    content = imageContent.detail ?? imageContent.url
                }
            case .assistant(let assistantMessage):
                rolePrefix = roleMapping[.assistant] ?? "Response:"
                content = assistantMessage.content
            case .tool(let toolMessage):
                rolePrefix = "Tool:"
                content = toolMessage.content
            case .function(let functionMessage):
                rolePrefix = "Function:"
                content = functionMessage.content
            }
            
            mappedMessages.append((rolePrefix, content))
        }
        
        return mappedMessages
    }
}


//https://www.llama.com/docs/model-cards-and-prompt-formats/meta-llama-2/
class LlamaFormatter: ChatFormatter {
    private let systemTemplate = "<<SYS>>\n{system_message}\n<</SYS>>\n\n"
    private let roles: [String: String] = [
        "user": "<s>[INST] ",
        "assistant": " [/INST] "
    ]
    private let endToken = "</s>"
    
    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse {
        let formattedMessages = mapRoles(messages: messages)
        let systemMessage = getSystemMessage(messages)
        let formattedSystemMessage = systemMessage.map { msg in
            systemTemplate.replacingOccurrences(of: "{system_message}", with: msg)
        }
        let prompt = formatPrompt(systemMessage: formattedSystemMessage, messages: formattedMessages)
        return ChatFormatterResponse(prompt: prompt, stop: [endToken])
    }
    
    private func getSystemMessage(_ messages: [ChatCompletionRequestMessage]) -> String? {
        for message in messages {
            if case .system(let systemMessage) = message {
                return systemMessage.content
            }
        }
        return nil
    }
    
    private func mapRoles(messages: [ChatCompletionRequestMessage]) -> [(String, String?)] {
        return messages.compactMap { message in
            switch message {
            case .system:
                return nil
            case .user(let userMessage):
                let content: String?
                switch userMessage.content {
                case .text(let text):
                    content = text
                case .image(let imageContent):
                    content = imageContent.detail
                }
                return (roles["user"] ?? "", content)
            case .assistant(let assistantMessage):
                return (roles["assistant"] ?? "", assistantMessage.content)
            case .tool, .function:
                return nil
            }
        }
    }
    
    private func formatPrompt(systemMessage: String?, messages: [(String, String?)]) -> String {
        var conversations: [String] = []
        var currentConversation = ""
        
        for (index, (role, content)) in messages.enumerated() {
            if index % 2 == 0 { // User message
                if !currentConversation.isEmpty {
                    conversations.append(currentConversation + " " + endToken)
                }
                currentConversation = role // <s>[INST]
                if index == 0 && systemMessage != nil {
                    currentConversation += systemMessage! + content!
                } else {
                    currentConversation += content ?? ""
                }
            } else { // Assistant message
                if let content = content {
                    currentConversation += role + content // [/INST] response
                }
            }
        }
        
        // Add the last conversation if it's a user message without response
        if messages.count % 2 != 0 {
            currentConversation += roles["assistant"]!
            conversations.append(currentConversation)
        } else if !currentConversation.isEmpty {
            conversations.append(currentConversation + endToken)
        }
        
        return conversations.joined(separator: "\n")
    }
}

//https://www.llama.com/docs/model-cards-and-prompt-formats/meta-llama-3/
class Llama3Formatter: ChatFormatter {
    private let roles: [String: String] = [
        "system": "<|start_header_id|>system<|end_header_id|>\n\n",
        "user": "<|start_header_id|>user<|end_header_id|>\n\n",
        "assistant": "<|start_header_id|>assistant<|end_header_id|>\n\n"
    ]
    private let endToken = "<|eot_id|>"

    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse {
        var formattedMessages = mapRoles(messages: messages)
        
        formattedMessages.append((roles["assistant"] ?? "", nil))
        
        let prompt = formatPrompt(formattedMessages)
        
        return ChatFormatterResponse(prompt: prompt, stop: [endToken])
    }

    private func mapRoles(messages: [ChatCompletionRequestMessage]) -> [(String, String?)] {
        return messages.map { message in
            var rolePrefix = ""
            var content: String? = ""
            
            switch message {
            case .system(let systemMessage):
                rolePrefix = roles["system"] ?? ""
                content = systemMessage.content
            case .user(let userMessage):
                rolePrefix = roles["user"] ?? ""
                switch userMessage.content {
                case .text(let text):
                    content = text
                case .image(let imageContent):
                    content = imageContent.detail
                }
            case .assistant(let assistantMessage):
                rolePrefix = roles["assistant"] ?? ""
                content = assistantMessage.content
            case .tool(let toolMessage):
                rolePrefix = roles["tool"] ?? ""
                content = toolMessage.content
            case .function(let functionMessage):
                rolePrefix = roles["function"] ?? ""
                content = functionMessage.content
            }
            
            return (rolePrefix, content)
        }
    }

    private func formatPrompt(_ formattedMessages: [(String, String?)]) -> String {
        var prompt = "<|begin_of_text|>"
        for (role, content) in formattedMessages {
            if let content = content {
                prompt += "\(role)\(content.trimmingCharacters(in: .whitespacesAndNewlines))\(endToken)"
            } else {
                prompt += "\(role) "
            }
        }
        return prompt.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}


//https://ai.google.dev/gemma/docs/formatting
class GemmaFormatter: ChatFormatter {
    private let roles: [String: String] = [
        "user": "<start_of_turn>user\n",
        "assistant": "<start_of_turn>model\n"
    ]
    
    private let endToken = "<end_of_turn>"
    private let separator = "<end_of_turn>\n"
    
    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse {
        var formattedMessages = mapRoles(messages: messages)
        formattedMessages.append((roles["assistant"]!, nil))
        let prompt = formatPrompt(formattedMessages)
        
        return ChatFormatterResponse(prompt: prompt, stop: [endToken])
    }
    
    private func mapRoles(messages: [ChatCompletionRequestMessage]) -> [(String, String?)] {
        return messages.compactMap { message in
            switch message {
            case .system:
                return nil
            case .user(let userMessage):
                let content: String?
                switch userMessage.content {
                case .text(let text):
                    content = text
                case .image(let imageContent):
                    content = imageContent.detail
                }
                return (roles["user"] ?? "", content)
            case .assistant(let assistantMessage):
                return (roles["assistant"] ?? "", assistantMessage.content)
            case .tool, .function:
                return nil
            }
        }
    }
    
    private func formatPrompt(_ formattedMessages: [(String, String?)]) -> String {
        var prompt = ""
        
        for (index, (role, content)) in formattedMessages.enumerated() {
            if index == formattedMessages.count - 1 {
                prompt += role
            } else if let content = content {
                prompt += "\(role)\(content)\(separator)"
            }
        }
        return prompt.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}

// https://qwen.readthedocs.io/zh-cn/latest/getting_started/concepts.html#control-tokens-chat-template
class QwenFormatter: ChatFormatter {
    private let roles: [String: String] = [
        "user": "<|im_start|>user",
        "assistant": "<|im_start|>assistant"
    ]
    
    private let systemTemplate = "<|im_start|>system\n{system_message}"
    private let defaultSystemMessage = "You are a helpful assistant."
    private let separator = "<|im_end|>"
    private let endToken = "<|endoftext|>"
    
    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse {
        let systemMessage = formatSystemMessage()
        var formattedMessages = mapRoles(messages: messages)
        formattedMessages.append((roles["assistant"]!, nil))
        let prompt = formatChatML(systemMessage: systemMessage, messages: formattedMessages)
        return ChatFormatterResponse(prompt: prompt, stop: [endToken])
    }
    
    private func formatSystemMessage() -> String {
        return systemTemplate.replacingOccurrences(of: "{system_message}", with: defaultSystemMessage)
    }
    
    private func mapRoles(messages: [ChatCompletionRequestMessage]) -> [(String, String?)] {
        return messages.compactMap { message in
            switch message {
            case .user(let userMessage):
                let content: String?
                switch userMessage.content {
                case .text(let text):
                    content = text
                case .image(let imageContent):
                    content = imageContent.detail
                }
                return (roles["user"]!, content)
            case .assistant(let assistantMessage):
                return (roles["assistant"]!, assistantMessage.content)
            case .system, .tool, .function:
                return nil
            }
        }
    }
    
    private func formatChatML(systemMessage: String, messages: [(String, String?)]) -> String {
        var prompt = systemMessage.isEmpty ? "" : "\(systemMessage)\(separator)\n"
        for (role, content) in messages {
            if let content = content {
                prompt += "\(role)\n\(content)\(separator)\n"
            } else {
                prompt += "\(role)\n"
            }
        }
        return prompt.trimmingCharacters(in: .whitespacesAndNewlines)
    }
}

// https://www.promptingguide.ai/models/mistral-7b#chat-template-for-mistral-7b-instruct
class MistralFormatter: ChatFormatter {
    private let endToken = "</s>"
    private let conversationStart = "<s>"
    private let instructStart = "[INST] "
    private let instructEnd = " [/INST] "
    
    func format(messages: [ChatCompletionRequestMessage]) -> ChatFormatterResponse {
        var prompt = conversationStart  // Add <s> only once at the start
        
        for (index, message) in messages.enumerated() {
            switch message {
            case .user(let userMessage):
                switch userMessage.content {
                case .text(let text):
                    prompt += "\(instructStart)\(text)"
                case .image:
                    continue
                }
                
            case .assistant(let assistantMessage):
                if let content = assistantMessage.content {
                    prompt += "\(instructEnd)\(content)\(endToken)"
                }     
            default:
                continue
            }
        }
        
        // Add instructEnd if the last message was from user (waiting for AI response)
        if messages.last.map({ if case .user = $0 { return true } else { return false } }) ?? false {
            prompt += instructEnd
        }
        
        return ChatFormatterResponse(prompt: prompt, stop: [endToken])
    }
}
