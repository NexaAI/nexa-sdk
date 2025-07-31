import Foundation
import NexaBridge

private typealias ml_LLM = OpaquePointer

final public class LLM {
    private var handle: ml_LLM?
    private var modelPath: String
    private var tokenizerPath: String?
    private var config: ModelConfig?
    private var device: String?

    public init(
        modelPath: String,
        tokenizerPath: String? = nil,
        config: ModelConfig = .default,
        device: String? = nil
    ) {
        self.modelPath = modelPath
        self.tokenizerPath = tokenizerPath
        self.config = config
        self.device = device
    }

    public func loadModel() throws {
        unload()
        
        let model = strdup(modelPath)!
        defer { free(model) }

        let tokenizer = tokenizerPath != nil ? strdup(tokenizerPath!)! : nil
        defer { free(tokenizer) }

        let dev = device != nil ? strdup(device!) : nil
        defer { free(dev) }

        let chatTemplatePath = config?.chatTemplatePath != nil ? strdup(config?.chatTemplatePath!)! : nil
        defer { free(chatTemplatePath) }

        let chatTemplateContent = config?.chatTemplateContent != nil ? strdup(config?.chatTemplateContent!)! : nil
        defer { free(chatTemplateContent) }

        let modelConfig: ml_ModelConfig
        if let config {
            modelConfig = ml_ModelConfig(
                n_ctx: config.nCtx,
                n_threads: config.nThreads,
                n_threads_batch: config.nThreadsBatch,
                n_batch: config.nBatch,
                n_ubatch: config.nUbatch,
                n_seq_max: config.nSeqMax,
                chat_template_path: chatTemplateContent == nil ? chatTemplatePath : nil,
                chat_template_content: chatTemplateContent
            )
        } else {
            modelConfig = ml_model_config_default()
        }

        ml_init()
        handle = ml_llm_create(model, tokenizerPath, modelConfig, dev)
        guard handle != nil else {
            ml_deinit()
            throw LLMError.modelLoadingFailed
        }
    }

    public func reset() {
        guard let handle else {
            return
        }
        ml_llm_reset(handle)
    }

    public func setSampler(config: SamplerConfig) throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }
        config.withUnsafePointerC { configPtr in
            ml_llm_set_sampler(handle, configPtr)
        }
    }

    public func resetSampler() throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }
        ml_llm_reset_sampler(handle)
    }

    public func encode(text: String) throws -> [Int32] {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        let textPtr = strdup(text)!
        defer { free(textPtr) }

        var outTokensPtr: UnsafeMutablePointer<Int32>? = nil
        let result = ml_llm_encode(handle, textPtr, &outTokensPtr)
        if result < 0 {
            throw LLMError.encodingFailed(result)
        }
        guard let outTokens = outTokensPtr else { return [] }
        let outTokensArray = UnsafeBufferPointer(start: outTokens, count: Int(result))
        return Array(outTokensArray)
    }

    public func decode(tokens: [Int32]) throws -> String {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        var outTextPtr: UnsafeMutablePointer<CChar>? = nil
        let result = ml_llm_decode(handle, tokens, Int32(tokens.count), &outTextPtr)
        if result < 0 {
            throw LLMError.decodingFailed(result)
        }
        guard let outText = outTextPtr else { return "" }
        defer { free(outText) }

        return String(cString: outText)
    }

    public func saveKVCache(to path: String) throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        let pathPtr = strdup(path)!
        defer { free(pathPtr) }

        let result = ml_llm_save_kv_cache(handle, pathPtr)
        if result < 0 {
            throw LLMError.kvCacheSaveFailed(result)
        }
    }

    public func loadKVCache(from path: String) throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        let pathPtr = strdup(path)!
        defer { free(pathPtr) }

        let result = ml_llm_load_kv_cache(handle, pathPtr)
        if result < 0 {
            throw LLMError.kvCacheLoadFailed(result)
        }
    }

    @NexaAIActor
    public func generate(prompt: String, config: GenerationConfig = .default) throws -> String {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        let promptPtr = strdup(prompt)!
        defer { free(promptPtr) }

        var outTokensPtr: UnsafeMutablePointer<CChar>? = nil
        let result = config.withUnsafePointerC { configPtr in
            return ml_llm_generate(handle, promptPtr, configPtr, &outTokensPtr)
        }

        if result < 0 {
            throw LLMError.generateFailed(result)
        }
        guard let outTokens = outTokensPtr else { return "" }
        defer { free(outTokens) }

        return String(cString: outTokens)
    }

    public func getChatTemplate(name: String? = nil) throws -> String {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        let templateNamePtr = name == nil ? nil : strdup(name!)
        defer { free(templateNamePtr) }

        var outTemplatePtr: UnsafePointer<CChar>? = nil
        let result = ml_llm_get_chat_template(handle, templateNamePtr, &outTemplatePtr)
        if result < 0 {
            throw LLMError.getTemplateFailed(result)
        }
        guard let outTemplate = outTemplatePtr else { return "" }
        return String(cString: outTemplate)
    }

    @NexaAIActor
    public func applyChatTemplate(messages: [ChatMessage]) throws -> String {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        guard !messages.isEmpty else { return "" }

        var cStrings: [UnsafePointer<CChar>] = []
        let mlMessages = UnsafeMutablePointer<ml_ChatMessage>.allocate(capacity: messages.count)
        for (index, message) in messages.enumerated() {
            let cRole = strdup(message.role.rawValue)
            let cContent = strdup(message.content)
            cStrings.append(cRole!)
            cStrings.append(cContent!)
            mlMessages[index] = ml_ChatMessage(role: cRole, content: cContent)
        }
        defer {
            for cString in cStrings {
                free(UnsafeMutablePointer(mutating: cString))
            }
            mlMessages.deallocate()
        }

        let count = Int32(messages.count)
        var outTextPtr: UnsafeMutablePointer<CChar>? = nil
        let result = ml_llm_apply_chat_template(handle, mlMessages, count, &outTextPtr)

        if result < 0 {
            throw LLMError.applyChatTemplateFailed(result)
        }

        guard let outText = outTextPtr else { return "" }
        defer { free(outText) }

        return String(cString: outText)
    }

    public func embed(texts: [String]) throws -> [Float] {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        var outEmbeddingsPtr: UnsafeMutablePointer<Float>? = nil
        let textCount = Int32(texts.count)
        let result = texts.withUnsafeMutableBufferPointerC { textPtr in
            return ml_llm_embed(handle, textPtr, textCount, &outEmbeddingsPtr)
        }

        if result < 0 {
            throw LLMError.embeddingFailed(result)
        }
        guard let outEmbeddings = outEmbeddingsPtr else { return [] }
        defer { free(outEmbeddings) }

        let buffer = UnsafeBufferPointer(start: outEmbeddings, count: Int(result))
        return Array(buffer)
    }

    @NexaAIActor
    public func generationAsyncStream(prompt: String, config: GenerationConfig = .default) -> AsyncThrowingStream<String, Error> {
        return .init { continuation in
            Task {
                _ = try generationStream(prompt: prompt, config: config) { token in
                    continuation.yield(token)
                    return true
                }
                continuation.finish()
            }
        }
    }

    @NexaAIActor
    public func generationStream(
        prompt: String,
        config: GenerationConfig,
        onToken: @escaping (String) -> Bool
    ) throws -> String {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        let promptPtr = strdup(prompt)!
        defer { free(promptPtr) }

        let holder = TokenCallbackWrapper(onToken)
        let userData = Unmanaged.passUnretained(holder).toOpaque()

        var fullTextPtr: UnsafeMutablePointer<CChar>? = nil
        let result = config.withUnsafePointerC { configPtr in
            return ml_llm_generate_stream(
                handle,
                promptPtr,
                configPtr,
                tokenCallback,
                userData,
                &fullTextPtr
            )
        }

        if result < 0 {
            throw LLMError.generationStreamFailed(result)
        }

        guard let fullText = fullTextPtr else { return "" }
        defer { free(fullText) }

        return String(cString: fullText)
    }

    public func setloRA(_ loRAId: Int32) throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }
        ml_llm_set_lora(handle, loRAId)
    }

    public func removeLoRA(by loRAId: Int32) throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }
        ml_llm_remove_lora(handle, loRAId)
    }

    public func addLoRA(_ path: String) throws {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }
        let pathPtr = strdup(path)!
        defer {
            free(pathPtr)
        }
        let result = ml_llm_add_lora(handle, pathPtr)
        if result < 0 {
            throw LLMError.addLoRAFailed(result)
        }
    }

    public func listLoRAs() throws -> [Int32] {
        guard let handle else {
            throw LLMError.modelLoadingFailed
        }

        var outPtr: UnsafeMutablePointer<Int32>? = nil
        let result = ml_llm_list_loras(handle, &outPtr)
        if result < 0 {
            throw LLMError.listLoRAsFailed(result)
        }

        guard let out = outPtr else { return [] }
        defer { free(out) }

        let buffer = UnsafeBufferPointer(start: out, count: Int(result))
        return Array(buffer)
    }

    @discardableResult
    public func getProfileData() -> ProfilingData? {
        if let handle {
            var out = ml_ProfilingData()
            ml_llm_get_profiling_data(handle, &out)
            return .init(from: out)
        }
        return nil
    }

    public func unload() {
        if let handle {
            ml_llm_destroy(handle)
            ml_deinit()
        }
        handle = nil
    }

    deinit {
        if let handle {
            ml_llm_destroy(handle)
            ml_deinit()
        }
    }
}

extension LLM: Model {
    public func loadModel(modelPath: String, mmprojPath: String?, config: ModelConfig?) throws {
        self.modelPath = modelPath
        self.tokenizerPath = mmprojPath
        self.config = config
        unload()
        try loadModel()
    }

    public var type: ModelType { .llm }
}

extension LLM {
    public enum LLMError: LocalizedError {
        case modelLoadingFailed
        case encodingFailed(Int32)
        case decodingFailed(Int32)
        case kvCacheSaveFailed(Int32)
        case kvCacheLoadFailed(Int32)
        case generateFailed(Int32)
        case getTemplateFailed(Int32)
        case applyChatTemplateFailed(Int32)
        case embeddingFailed(Int32)
        case generationStreamFailed(Int32)
        case addLoRAFailed(Int32)
        case listLoRAsFailed(Int32)

        public var errorDescription: String? {
            switch self {
            case .modelLoadingFailed:
                return "Model loading failed"
            case .encodingFailed(let code),
                 .decodingFailed(let code),
                 .kvCacheSaveFailed(let code),
                 .kvCacheLoadFailed(let code),
                 .generateFailed(let code),
                 .getTemplateFailed(let code),
                 .applyChatTemplateFailed(let code),
                 .embeddingFailed(let code),
                 .generationStreamFailed(let code),
                 .addLoRAFailed(let code),
                 .listLoRAsFailed(let code):
                if let errorMessage = ml_get_error_message(ml_ErrorCode(rawValue: code)) {
                    let result = String(cString: errorMessage)
                    return result
                }
                return "unknow error, code: \(code)"
            }
        }
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

                let mlSamplerConfig = ml_SamplerConfig(
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
                let stopCount = Int32(stop.count)
                return try stop.withUnsafeMutableBufferPointerC { buffer in
                   var mlGenerationConfig = ml_GenerationConfig(
                        max_tokens: maxTokens,
                        stop: buffer,
                        stop_count: stopCount,
                        n_past: nPast,
                        sampler_config: mlSamplerConfig,
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
