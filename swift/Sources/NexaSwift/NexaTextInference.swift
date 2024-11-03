import Foundation
import llama
import Combine

public class NexaTextInference {
    private let model: LlamaModel
    private let modelPath: String
    private var generatedTokenCache = ""
    private var contentStarted = false
    private let chatFormatterRegistry: ChatFormatterRegistry

    var maxLengthOfStopToken: Int {
        model.configuration.stopTokens.map { $0.count }.max() ?? 0
    }

    public init(modelPath: String,
                 modelConfiguration: Configuration = .init()) throws {
        if modelPath.isEmpty {
            throw NSError(domain: "InvalidParameterError", code: 400, userInfo: [NSLocalizedDescriptionKey: "Either modelPath or localPath must be provided."])
        }
        self.model = try LlamaModel(path: modelPath, configuration: modelConfiguration)
        self.modelPath = modelPath
        self.chatFormatterRegistry = .init()
    }
    
    private func updateConfiguration(
        temperature: Float?,
        maxNewToken: Int?,
        topK: Int32?,
        topP: Float?,
        stopTokens: [String]?
    ) {
        if let temperature = temperature {
            model.configuration.temperature = temperature
        }
        if let maxNewToken = maxNewToken {
            model.configuration.maxNewToken = maxNewToken
        }
        if let topK = topK {
            model.configuration.topK = topK
        }
        if let topP = topP {
            model.configuration.topP = topP
        }
        if let stopTokens = stopTokens {
            model.configuration.stopTokens = stopTokens
        }
    }

    private func getFormatterForModel() -> ChatFormatter? {
        let modelArch = model.arch.lowercased()
        let lowerModelPath = modelPath.lowercased()

        let modelType: ChatCompletionModel? = {
            switch modelArch {
            case _ where modelArch.contains("gemma"):
                // For Gemma-based models, check the model path
                if lowerModelPath.contains("octopus-v2") || lowerModelPath.contains("octopusv2") {
                    return .octopusv2
                } else {
                    return .gemma
                }
            case _ where modelArch.contains("qwen"):
                return .qwen
            case _ where modelArch.contains("llama"):
                // For Llama-based models, check the model path
                if lowerModelPath.contains("llama-2") || lowerModelPath.contains("llama2") {
                    return .llama
                } else if lowerModelPath.contains("llama-3") || lowerModelPath.contains("llama3") {
                    return .llama3
                } else if lowerModelPath.contains("mistral") {
                    return .mistral
                } else {
                    // If can't determine specific version, default to Llama2
                    print("Warning: Unable to determine specific Llama model version from path: \(modelPath). Defaulting to Llama2 format.")
                    return .llama
                }
            default:
                print("Warning: Unknown model architecture: \(modelArch). Defaulting to Llama2 format.")
                return .llama
            }
        }()
        
        return chatFormatterRegistry.getFormatter(name: modelType?.rawValue)
    }

    private func isStopToken() -> Bool {
        model.configuration.stopTokens.reduce(false) { partialResult, stopToken in
            generatedTokenCache.hasSuffix(stopToken)
        }
    }

    private func response(for prompt: String, output: (String) -> Void, finish: () -> Void) {
        func finaliseOutput() {
            model.configuration.stopTokens.forEach {
                generatedTokenCache = generatedTokenCache.replacingOccurrences(of: $0, with: "")
            }
            output(generatedTokenCache)
            finish()
            generatedTokenCache = ""
        }
        defer { model.clear() }
        do {
            try model.start(for: prompt)
            while model.shouldContinue {
                var delta = try model.continue()
                if contentStarted { // remove the prefix empty spaces
                    if needToStop(after: delta, output: output) {
                        finish()
                        break
                    }
                } else {
                    delta = delta.trimmingCharacters(in: .whitespacesAndNewlines)
                    if !delta.isEmpty {
                        contentStarted = true
                        if needToStop(after: delta, output: output) {
                            finish()
                            break
                        }
                    }
                }
            }
            finaliseOutput()
        } catch {
            finaliseOutput()
        }
    }

    private func needToStop(after delta: String, output: (String) -> Void) -> Bool {
        guard maxLengthOfStopToken > 0 else {
            output(delta)
            return false
        }
        generatedTokenCache += delta

        if generatedTokenCache.count >= maxLengthOfStopToken * 2 {
            if let stopToken = model.configuration.stopTokens.first(where: { generatedTokenCache.contains($0) }),
            let index = generatedTokenCache.range(of: stopToken) {
                let outputCandidate = String(generatedTokenCache[..<index.lowerBound])
                output(outputCandidate)
                generatedTokenCache = ""
                return true
            } else { 
                let outputCandidate = String(generatedTokenCache.prefix(maxLengthOfStopToken))
                generatedTokenCache.removeFirst(outputCandidate.count)
                output(outputCandidate)
                return false
            }
        }
        return false
    }

    
    @NexaSwiftActor
    private func run(for prompt: String) -> AsyncThrowingStream<String, Error> {
        return .init { continuation in
            Task {
                response(for: prompt) { [weak self] delta in
                    continuation.yield(delta)
                } finish: { [weak self] in
                    continuation.finish()
                }
            }
        }
    }

    @NexaSwiftActor
    public func createCompletion(
        for prompt: String,
        temperature: Float? = nil,
        maxNewToken: Int? = nil,
        topK: Int32? = nil,
        topP: Float? = nil,
        stopTokens: [String]? = nil) async throws -> CompletionResponse {
            updateConfiguration(
                temperature: temperature,
                maxNewToken: maxNewToken,
                topK: topK,
                topP: topP,
                stopTokens: stopTokens
            )
            model.reset()
            var result = ""
                for try await value in await run(for: prompt) {
                result += value
            }

            let completionResponse = CompletionResponse(
                id: UUID().uuidString,
                object: "text_completion",
                created: Int(Date().timeIntervalSince1970),
                model: "",
                choices: [
                    CompletionChoice(
                        text: result,
                        index: 0,
                        logprobs: nil,
                        finishReason: FinishReason.stop
                    )
                ],
                usage: CompletionUsage(
                    promptTokens: 0,
                    completionTokens: 0,
                    totalTokens: 0
                )
            )
            return completionResponse
    }
    
    @NexaSwiftActor
    public func createCompletionStream(
        for prompt: String,
        temperature: Float? = nil,
        maxNewToken: Int? = nil,
        topK: Int32? = nil,
        topP: Float? = nil,
        stopTokens: [String]? = nil) -> AsyncThrowingStream<CompletionResponse, Error>  {
            updateConfiguration(
                temperature: temperature,
                maxNewToken: maxNewToken,
                topK: topK,
                topP: topP,
                stopTokens: stopTokens
            )
            model.reset()
            return .init { continuation in
                Task {
                    var index = 0
                    response(for: prompt) { text in
                        let completionResponse = CompletionResponse(
                            id: UUID().uuidString,
                            object: "text_completion",
                            created: Int(Date().timeIntervalSince1970),
                            model: "",
                            choices: [
                                CompletionChoice(
                                    text: text,
                                    index: 0,
                                    logprobs: nil,
                                    finishReason: FinishReason.stop
                                )
                            ],
                            usage: CompletionUsage(
                                promptTokens: 0,
                                completionTokens: 0,
                                totalTokens: 0
                            )
                    )
                    
                    index += 1
                    continuation.yield(completionResponse)
                } finish: {
                    continuation.finish()
                }
            }
        }
    }
    
    @NexaSwiftActor
    public func createChatCompletion(
        for messages: [ChatCompletionRequestMessage],
        temperature: Float? = nil,
        maxNewToken: Int? = nil,
        topK: Int32? = nil,
        topP: Float? = nil,
        stopTokens: [String]? = nil,
        modelType: ChatCompletionModel? = nil) async throws -> ChatCompletionResponse {
            let formatter = modelType.map { chatFormatterRegistry.getFormatter(name: $0.rawValue) } ?? getFormatterForModel()
            let chatFormatter: ChatFormatterResponse? = formatter?.format(messages: messages)
            // let chatFormatter: ChatFormatterResponse? = chatFormatterRegistry.getFormatter(name: modelType?.rawValue)?.format(messages: messages) ?? nil
            updateConfiguration(
                temperature: temperature,
                maxNewToken: maxNewToken,
                topK: topK,
                topP: topP,
                stopTokens: stopTokens ?? (!model.configuration.stopTokens.isEmpty ? model.configuration.stopTokens : chatFormatter?.stop) ?? nil
            )
            model.reset()
            
            var result = ""
            for try await value in await run(for: chatFormatter?.prompt ?? "") {
                result += value
            }
            
        let response = ChatCompletionResponse(
            id: UUID().uuidString,
            object: "chat.completion",
            created: Int(Date().timeIntervalSince1970),
            model: chatFormatterRegistry.getFormatterName(name: modelType?.rawValue),
            choices: [
                ChatCompletionResponseChoice(
                    index: 0,
                    message: ChatCompletionResponseMessage(
                        content: result,
                        toolCalls: nil,
                        role: nil,
                        functionCall: nil
                    ),
                    logprobs: nil,
                    finishReason: FinishReason.stop
                )
            ],
            usage: nil
        )
        
        return response
    }
    
   @NexaSwiftActor
   public func createChatCompletionStream(
       for messages: [ChatCompletionRequestMessage],
       temperature: Float? = nil,
       maxNewToken: Int? = nil,
       topK: Int32? = nil,
       topP: Float? = nil,
       stopTokens: [String]? = nil,
       modelType: ChatCompletionModel? = nil
   ) -> AsyncThrowingStream<CreateChatCompletionStreamResponse, Error> {
           model.reset()
            let formatter = modelType.map { chatFormatterRegistry.getFormatter(name: $0.rawValue) } ?? getFormatterForModel()
            let chatFormatter: ChatFormatterResponse? = formatter?.format(messages: messages)
//            let chatFormatter: ChatFormatterResponse? = chatFormatterRegistry.getFormatter(name: modelType?.rawValue)?.format(messages: messages) ?? nil
           updateConfiguration(
               temperature: temperature,
               maxNewToken: maxNewToken,
               topK: topK,
               topP: topP,
               stopTokens: stopTokens ?? (!model.configuration.stopTokens.isEmpty ? model.configuration.stopTokens : chatFormatter?.stop) ?? nil
           )
           return .init { continuation in
                Task {
                    var index = 0
                    response(for: chatFormatter?.prompt ?? "") { text in
                        let response = CreateChatCompletionStreamResponse(
                            id: UUID().uuidString,
                            model: chatFormatterRegistry.getFormatterName(name: modelType?.rawValue),
                            object: "chat.completion.chunk",
                            created: Int(Date().timeIntervalSince1970),
                            choices: [
                                ChatCompletionStreamResponseChoice(
                                    index: index,
                                    delta: ChatCompletionStreamResponseDelta(
                                        content: text,
                                        functionCall: nil,
                                        toolCalls: nil,
                                        role: nil
                                    ),
                                    finishReason: FinishReason.stop,
                                    logprobs: nil
                                )
                            ]
                        )
                        
                        index += 1
                        continuation.yield(response)
                    } finish: {
                        continuation.finish()
                    }
                }
            }
   }
}
