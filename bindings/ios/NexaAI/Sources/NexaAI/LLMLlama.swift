// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

import Foundation
import NexaBridge

private typealias llmHandle = OpaquePointer

func createLlamaPlugin() -> UnsafeMutableRawPointer? {
    guard let pluginPointer = create_plugin() else {
        return nil
    }
    return UnsafeMutableRawPointer(pluginPointer)
}

final public class LLMLlama: Model {
    private var handle: llmHandle?
    public private(set) var lastProfileData: ProfileData?
    public private(set) var isLoaded = false
    private var continueStream = true
    public let type: ModelType = .llm

    public init() { }

    private func registerLlamaPlugin() {
        unregister()
        ml_register_plugin(plugin_id, createLlamaPlugin)
    }

    private func unregister() {
        if let handle {
            ml_llm_destroy(handle)
            ml_deinit()
        }
    }

    public func load(_ options: ModelOptions) throws {
        registerLlamaPlugin()
        isLoaded = false
        let cPath = strdup(options.modelPath)
        defer { free(cPath) }

        let cTokenizerPath = (options.tokenizerPath != nil) ? strdup(options.tokenizerPath!) : nil
        defer { free(cTokenizerPath) }

        let deviceId = options.deviceId != nil ? strdup(options.deviceId!) : nil
        defer { free(deviceId) }

        let config = options.config
        let chatTemplatePath = config.chatTemplatePath != nil ? strdup(config.chatTemplatePath!)! : nil
        defer { free(chatTemplatePath) }

        let chatTemplateContent = config.chatTemplateContent != nil ? strdup(config.chatTemplateContent!)! : nil
        defer { free(chatTemplateContent) }

        let modelConfig = ml_ModelConfig(
            n_ctx: config.nCtx,
            n_threads: config.nThreads,
            n_threads_batch: config.nThreadsBatch,
            n_batch: config.nBatch,
            n_ubatch: config.nUbatch,
            n_seq_max: config.nSeqMax,
            n_gpu_layers: options.gpuLayers,
            chat_template_path: chatTemplateContent == nil ? chatTemplatePath : nil,
            chat_template_content: chatTemplateContent,
            enable_sampling: false,
            grammar_str: nil,
            max_tokens: 0,
            enable_thinking: true,
            verbose: false,
            qnn_model_folder_path: nil,
            qnn_lib_folder_path: nil
        )

        var input = ml_LlmCreateInput(
            model_name: nil,
            model_path: cPath,
            tokenizer_path: cTokenizerPath,
            config: modelConfig,
            plugin_id: plugin_id(),
            device_id: deviceId,
            license_id: nil,
            license_key: nil
        )

        ml_init()
        let result = ml_llm_create(&input, &handle)
        if result < 0 {
            ml_deinit()
            throw LLMError.modelLoadingFailed(result)
        }

        isLoaded = true
    }

    @NexaAIActor
    public func generate(prompt: String, config: GenerationConfig = .default) throws -> GenerateResult {
        return try generate(prompt: prompt, config: config) { [weak self] _ in
            return self?.continueStream ?? true
        }
    }

    @NexaAIActor
    public func generateAsyncStream(messages: [ChatMessage], options: GenerationOptions = .init()) throws -> AsyncThrowingStream<String, Error> {
        let prompt = try applyChatTemplate(messages: messages, options: options.templateOptions)
        return generateAsyncStream(prompt: prompt, config: options.config)
    }

    public func applyChatTemplate(messages: [ChatMessage], options: ChatTemplateOptions = .init()) throws -> String {
        var cStrings: [UnsafePointer<CChar>] = []
        let mlMessages = UnsafeMutablePointer<ml_LlmChatMessage>.allocate(capacity: messages.count)
        for (index, message) in messages.enumerated() {
            let cRole = strdup(message.role.rawValue)
            let cContent = strdup(message.content)
            cStrings.append(cRole!)
            cStrings.append(cContent!)
            mlMessages[index] = ml_LlmChatMessage(role: cRole, content: cContent)
        }

        let mlTools = options.tools.isEmpty ? nil : strdup(options.tools)
        defer { free(mlTools) }

        defer {
            for cString in cStrings {
                free(UnsafeMutablePointer(mutating: cString))
            }
            mlMessages.deallocate()
        }

        var input = ml_LlmApplyChatTemplateInput(
            messages: mlMessages,
            message_count: Int32(messages.count),
            tools: mlTools,
            enable_thinking: options.enableThinking
        )

        var output = ml_LlmApplyChatTemplateOutput()

        let result = ml_llm_apply_chat_template(handle, &input, &output)

        if result < 0 {
            throw LLMError.applyChatTemplateFailed(result)
        }

        guard let outTextPtr = output.formatted_text else {
            return ""
        }
        defer { ml_free(outTextPtr) }

        return String(cString: outTextPtr)
    }

    public func reset() {
        if isLoaded, let handle {
            ml_llm_reset(handle)
        }
    }

    public func saveKVCache(to path: String) throws {
        let pathPtr = strdup(path)!
        defer { free(pathPtr) }

        var input = ml_KvCacheSaveInput(path: pathPtr)
        var output = ml_KvCacheSaveOutput()
        let result = ml_llm_save_kv_cache(handle, &input, &output)

        if result < 0 {
            throw LLMError.kvCacheSaveFailed(result)
        }
    }

    public func loadKVCache(from path: String) throws {
        let pathPtr = strdup(path)!
        defer { free(pathPtr) }

        var input = ml_KvCacheLoadInput(path: pathPtr)
        var output = ml_KvCacheLoadOutput()
        let result = ml_llm_load_kv_cache(handle, &input, &output)

        if result < 0 {
            throw LLMError.kvCacheLoadFailed(result)
        }
    }

    public func stopStream() {
        continueStream = false
    }

    @NexaAIActor
    public func generateAsyncStream(prompt: String, config: GenerationConfig = .default) -> AsyncThrowingStream<String, Error> {
        return .init { continuation in
            Task {
                do {
                    lastProfileData = nil
                    continueStream = true
                    let result = try generate(prompt: prompt, config: config) { [weak self] token in
                        continuation.yield(token)
                        return self?.continueStream ?? true
                    }
                    lastProfileData = result.profileData
                    continuation.finish()
                } catch {
                    continuation.finish(throwing: error)
                }
            }
        }
    }

    @NexaAIActor
    @discardableResult
    public func generate(
        prompt: String,
        config: GenerationConfig,
        onToken: @escaping (String) -> Bool
    ) throws -> GenerateResult {
        let cPrompt = strdup(prompt)
        defer { free(cPrompt) }

        let holder = TokenCallbackWrapper(onToken)
        let userData = Unmanaged.passUnretained(holder).toOpaque()

        var output: ml_LlmGenerateOutput = ml_LlmGenerateOutput()
        let result = config.withUnsafePointerC { configPtr in
            var input = ml_LlmGenerateInput(prompt_utf8: cPrompt, config: configPtr, on_token: tokenCallback, user_data: userData)
            return ml_llm_generate(handle, &input, &output)
        }

        if result < 0 {
            throw LLMError.generateFailed(result)
        }

        guard let fullText = output.full_text else {
            throw LLMError.generateEmptyString
        }
        defer { ml_free(fullText) }

        return .init(response: String(cString: fullText), profileData: .init(from: output.profile_data))
    }

    deinit {
        if let handle {
            ml_llm_destroy(handle)
            ml_deinit()
        }
        handle = nil
    }
}

extension GenerationConfig {

    func withUnsafePointerC<T>(_ body: (UnsafePointer<ml_GenerationConfig>) throws -> T) rethrows -> T {
        let imageCount = Int32(imagePaths.count)
        let audioCount = Int32(audioPaths.count)
        return try imagePaths.withUnsafeMutableBufferPointerC { imagePaths in
            return try audioPaths.withUnsafeMutableBufferPointerC { audioPaths in
                let samplerConfig = samplerConfig
                let grammarPathPtr = samplerConfig.grammarPath == nil ? nil : strdup(samplerConfig.grammarPath)
                defer { free(grammarPathPtr) }
                let grammarPathStringPtr = samplerConfig.grammarString == nil ? nil : strdup(samplerConfig.grammarString)
                defer { free(grammarPathStringPtr) }

                var mlSamplerConfig = ml_SamplerConfig(
                    temperature: samplerConfig.temperature,
                    top_p: samplerConfig.topP,
                    top_k: samplerConfig.topK,
                    min_p: samplerConfig.minP,
                    repetition_penalty: samplerConfig.repetitionPenalty,
                    presence_penalty: samplerConfig.presencePenalty,
                    frequency_penalty: samplerConfig.frequencyPenalty,
                    seed: samplerConfig.seed,
                    grammar_path: grammarPathStringPtr == nil ? grammarPathPtr : nil,
                    grammar_string: grammarPathStringPtr
                )
                return try withUnsafeMutablePointer(to: &mlSamplerConfig) { sampleConfig in
                    let stopCount = Int32(stop.count)
                    return try stop.withUnsafeMutableBufferPointerC { buffer in
                       var mlGenerationConfig = ml_GenerationConfig(
                            max_tokens: maxTokens,
                            stop: buffer,
                            stop_count: stopCount,
                            n_past: nPast,
                            sampler_config: sampleConfig,
                            image_paths: imageCount == 0 ? nil : imagePaths,
                            image_count: imageCount,
                            audio_paths: audioCount == 0 ? nil : audioPaths,
                            audio_count: audioCount
                        )
                        return try withUnsafePointer(to: &mlGenerationConfig, body)
                    }
                }


            }
        }
    }
}

extension SamplerConfig {

    func withUnsafePointerC<T>(_ body: (UnsafePointer<ml_SamplerConfig>) throws -> T) rethrows -> T {
        let grammarPathPtr = grammarPath == nil ? nil : strdup(grammarPath)
        defer { free(grammarPathPtr) }

        let grammarPathStringPtr = grammarString == nil ? nil : strdup(grammarString)
        defer { free(grammarPathStringPtr) }

        var mlSamplerConfig = ml_SamplerConfig(
            temperature: temperature,
            top_p: topP,
            top_k: topK,
            min_p: minP,
            repetition_penalty: repetitionPenalty,
            presence_penalty: presencePenalty,
            frequency_penalty: frequencyPenalty,
            seed: seed,
            grammar_path: grammarPathStringPtr == nil ? grammarPathPtr : nil,
            grammar_string: grammarPathStringPtr
        )
        return try withUnsafePointer(to: &mlSamplerConfig, body)
    }
}


final class TokenCallbackWrapper {
    let callback: (String) -> Bool
    init(_ callback: @escaping (String) -> Bool) {
        self.callback = callback
    }
}

func tokenCallback(
    _ token: UnsafePointer<CChar>?,
    _ userData: UnsafeMutableRawPointer?
) -> Bool {
    guard let userData else {
        return false
    }
    let holder = Unmanaged<TokenCallbackWrapper>
        .fromOpaque(userData)
        .takeUnretainedValue()
    let tokenStr = token.flatMap { String(cString: $0) } ?? ""
    return holder.callback(tokenStr)
}
