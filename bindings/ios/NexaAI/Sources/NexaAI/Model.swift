// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

import Foundation

public enum ModelType: CaseIterable {
    case llm
    case vlm
}

public protocol Model {
    @NexaAIActor func load(_ options: ModelOptions) throws

    @NexaAIActor func applyChatTemplate(messages: [ChatMessage], options: ChatTemplateOptions) throws -> String
    @NexaAIActor func generateAsyncStream(messages: [ChatMessage], options: GenerationOptions) throws -> AsyncThrowingStream<String, Error>

    @NexaAIActor func generate(prompt: String, config: GenerationConfig) throws -> GenerateResult

    @NexaAIActor var lastProfileData: ProfileData? { get }
    @NexaAIActor func reset()
    
    func stopStream()

    var isLoaded: Bool { get }
    var type: ModelType { get }
}

extension Model {

    @NexaAIActor
    public func generate(messages: [ChatMessage], options: GenerationOptions = .init()) throws -> GenerateResult  {
        let prompt = try applyChatTemplate(messages: messages, options: options.templateOptions)
        return try generate(prompt: prompt, config: options.config)
    }

}

public struct GenerateResult {
    public var response: String
    public var profileData: ProfileData
    public init(response: String, profileData: ProfileData) {
        self.response = response
        self.profileData = profileData
    }
}

public struct GenerationOptions {
    public var config: GenerationConfig
    public var templateOptions: ChatTemplateOptions
    public init(config: GenerationConfig = .default, templateOptions: ChatTemplateOptions = .init()) {
        self.config = config
        self.templateOptions = templateOptions
    }
}

public struct ChatTemplateOptions {
    public var tools: String
    public var enableThinking: Bool

    public init(tools: String = "", enableThinking: Bool = false) {
        self.tools = tools
        self.enableThinking = enableThinking
    }
}

public struct ModelOptions {
    public var modelPath: String
    public var tokenizerPath: String?
    public var mmprojPath: String?
    public var config: ModelConfig
    public var deviceId: String?
    public var gpuLayers: Int32
    public var enableThinking: Bool

    public init(
        modelPath: String,
        tokenizerPath: String? = nil,
        mmprojPath: String? = nil,
        config: ModelConfig = .default,
        deviceId: String? = nil,
        gpuLayers: Int32 = 8,
        enableThinking: Bool = false
    ) {
        self.modelPath = modelPath
        self.tokenizerPath = tokenizerPath
        self.config = config
        self.deviceId = deviceId
        self.mmprojPath = mmprojPath
        self.gpuLayers = gpuLayers
        self.enableThinking = enableThinking
    }
}
