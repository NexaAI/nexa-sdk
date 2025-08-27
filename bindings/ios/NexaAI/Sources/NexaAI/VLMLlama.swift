import Foundation
import NexaBridge

private typealias vlmHandle = OpaquePointer

final public class VLMLlama: Model {
    private var handle: vlmHandle?

    public private(set) var lastProfileData: ProfileData?
    public private(set) var isLoaded = false
    public let type: ModelType = .vlm

    private var continueStream = true
    
    public init() { }

    private func registerLlamaPlugin() {
        unregister()
        ml_register_plugin(plugin_id, createLlamaPlugin)
    }

    private func unregister() {
        if let handle {
            ml_vlm_destroy(handle)
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

        let cMmprojectPath = (options.mmprojPath != nil) ? strdup(options.mmprojPath!) : nil
        defer { free(cMmprojectPath) }

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
            system_library_path: nil,
            backend_library_path: nil,
            extension_library_path: nil,
            config_file_path: nil,
            embedded_tokens_path: nil,
            max_tokens: 0,
            enable_thinking: true,
            verbose: false
        )

        var input = ml_VlmCreateInput(
            model_path: cPath,
            mmproj_path: cMmprojectPath,
            config: modelConfig,
            plugin_id: plugin_id(),
            device_id: deviceId,
            tokenizer_path: cTokenizerPath,
            license_id: nil,
            license_key: nil
        )

        ml_init()
        let result = ml_vlm_create(&input, &handle)
        if result < 0 {
            ml_deinit()
            throw LLMError.modelLoadingFailed(result)
        }
        isLoaded = true
    }

    public func stopStream() {
        continueStream = false
    }

    public func generationStream(
        messages: [ChatMessage],
        options: GenerationOptions = .init()
    ) throws -> GenerateResult {
        let prompt = try applyChatTemplate(messages: messages, options: options.templeteOptions)
        return try generationStream(prompt: prompt, config: options.config) { [weak self] _ in
            return self?.continueStream ?? true
        }
    }

    public func generationAsyncStream(
        messages: [ChatMessage],
        options: GenerationOptions = .init()
    ) throws -> AsyncThrowingStream<String,Error> {
        let prompt = try applyChatTemplate(messages: messages, options: options.templeteOptions)
        return generationAsyncStream(prompt: prompt, config: options.config)
    }

    public func applyChatTemplate(
        messages: [ChatMessage],
        options: ChatTemplateOptions = .init()
    ) throws -> String {
        var cStrings: [UnsafePointer<CChar>] = []
        var cVlmContent: [UnsafeMutablePointer<ml_VlmContent>] = []
        let mlMessages = UnsafeMutablePointer<ml_VlmChatMessage>.allocate(capacity: messages.count)
        for (index, message) in messages.enumerated() {
            let cRole = strdup(message.role.rawValue)
            let cContent = strdup(message.content)
            cStrings.append(cRole!)
            cStrings.append(cContent!)

            var contentCount = 0
            if !message.content.isEmpty {
                contentCount += 1
            }
            contentCount += message.images.count
            contentCount += message.audios.count

            let vlmContent = UnsafeMutablePointer<ml_VlmContent>.allocate(capacity: contentCount)
            var currentIndex = 0
            if !message.content.isEmpty {
                let cType = strdup("text")
                cStrings.append(cType!)
                let text = ml_VlmContent(type: cType, text: cContent)
                vlmContent[currentIndex] = text
                currentIndex += 1
            }

            for image in message.images {
                let cImage = strdup(image)
                cStrings.append(cImage!)
                let cType = strdup("image")
                cStrings.append(cType!)
                let image = ml_VlmContent(type: cType, text: cImage)
                vlmContent[currentIndex] = image
                currentIndex += 1
            }

            for audio in message.audios {
                let cAudio = strdup(audio)
                cStrings.append(cAudio!)
                let cType = strdup("audio")
                cStrings.append(cType!)
                let audio = ml_VlmContent(type: cType, text: cAudio)
                vlmContent[currentIndex] = audio
                currentIndex += 1
            }
            cVlmContent.append(vlmContent)

            mlMessages[index] = ml_VlmChatMessage(role: cRole, contents: vlmContent, content_count: Int64(contentCount))
        }

        let mlTools = options.tools.isEmpty ? nil : strdup(options.tools)
        defer { free(mlTools) }

        defer {
            for cString in cStrings {
                free(UnsafeMutablePointer(mutating: cString))
            }
            for vlmContent in cVlmContent {
                vlmContent.deallocate()
            }
            mlMessages.deallocate()
        }

        var input = ml_VlmApplyChatTemplateInput(
            messages: mlMessages,
            message_count: Int32(messages.count),
            tools: mlTools,
            enable_thinking: options.enableThinking
        )

        var output = ml_VlmApplyChatTemplateOutput()
        let result = ml_vlm_apply_chat_template(handle, &input, &output)

        if result < 0 {
            throw VLMError.applyChatTemplateFailed(result)
        }

        guard let outTextPtr = output.formatted_text else {
            return ""
        }
        defer { ml_free(outTextPtr) }

        return String(cString: outTextPtr)
    }

    public func reset() {
        if isLoaded, let handle {
            ml_vlm_reset(handle)
        }
    }

    @NexaAIActor
    private func generationAsyncStream(prompt: String, config: GenerationConfig = .default) -> AsyncThrowingStream<String, Error> {
        return .init { continuation in
            Task {
                do {
                    lastProfileData = nil
                    continueStream = true
                    let result = try generationStream(prompt: prompt, config: config) { [weak self] token in
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
    public func generationStream(
        prompt: String,
        config: GenerationConfig = .default,
        onToken: @escaping (String) -> Bool
    ) throws -> GenerateResult {
        let cPrompt = strdup(prompt)
        defer { free(cPrompt) }

        let holder = TokenCallbackWrapper(onToken)
        let userData = Unmanaged.passUnretained(holder).toOpaque()

        var output: ml_VlmGenerateOutput = ml_VlmGenerateOutput()
        let result = config.withUnsafePointerC { configPtr in
            var input = ml_VlmGenerateInput(prompt_utf8: cPrompt, config: configPtr, on_token: tokenCallback, user_data: userData)
            return ml_vlm_generate(handle, &input, &output)
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
            ml_vlm_destroy(handle)
            ml_deinit()
        }
        handle = nil
    }
}
