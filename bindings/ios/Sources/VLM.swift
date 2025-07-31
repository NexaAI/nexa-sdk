import Foundation
import NexaBridge

private typealias ml_VLM = OpaquePointer

final public class VLM {
    private var handle: ml_VLM?
    private var modelPath: String
    private var mmprojPath: String?
    private var contextLength: Int32
    private var device: String?

    public init(
        modelPath: String,
        mmprojPath: String? = nil,
        contextLength: Int32 = 1024,
        device: String? = nil
    ) {
        self.modelPath = modelPath
        self.mmprojPath = mmprojPath
        self.contextLength = contextLength
        self.device = device
    }

    public func loadModel() throws {
        unload()

        let model = strdup(modelPath)!
        defer { free(model) }

        let tokenizer = mmprojPath != nil ? strdup(mmprojPath!)! : nil
        defer { free(tokenizer) }

        let dev = device != nil ? strdup(device!) : nil
        defer { free(dev) }

        ml_init()
        handle = ml_vlm_create(model, mmprojPath, contextLength, dev)
        guard handle != nil else {
            ml_deinit()
            throw VLMError.modelLoadingFailed
        }
    }

    public func encode(text: String) throws -> [Int32] {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }

        let textPtr = strdup(text)!
        defer { free(textPtr) }

        var outTokensPtr: UnsafeMutablePointer<Int32>? = nil
        let result = ml_vlm_encode(handle, textPtr, &outTokensPtr)
        if result < 0 {
            throw VLMError.encodingFailed(result)
        }
        guard let outTokens = outTokensPtr else { return [] }
        let outTokensArray = UnsafeBufferPointer(start: outTokens, count: Int(result))
        return Array(outTokensArray)
    }

    public func decode(tokens: [Int32]) throws -> String {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }

        var outTextPtr: UnsafeMutablePointer<CChar>? = nil
        let result = ml_vlm_decode(handle, tokens, Int32(tokens.count), &outTextPtr)
        if result < 0 {
            throw VLMError.decodingFailed(result)
        }
        guard let outText = outTextPtr else { return "" }
        defer { free(outText) }

        return String(cString: outText)
    }

    public func setSampler(config: SamplerConfig) throws {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }
        config.withUnsafePointerC { configPtr in
            ml_vlm_set_sampler(handle, configPtr)
        }
    }

    public func resetSampler() throws {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }
        ml_vlm_reset_sampler(handle)
    }

    @NexaAIActor
    public func generate(prompt: String, config: GenerationConfig = .default) throws -> String {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }

        let promptPtr = strdup(prompt)!
        defer { free(promptPtr) }

        var outTokensPtr: UnsafeMutablePointer<CChar>? = nil
        let result = config.withUnsafePointerC { configPtr in
            return ml_vlm_generate(handle, promptPtr, configPtr, &outTokensPtr)
        }
        guard let outTokens = outTokensPtr else { return ""}
        defer { free(outTokens) }

        if result < 0 {
            throw VLMError.generateFailed(result)
        }
        return String(cString: outTokens)
    }

    public func getChatTemplate(name: String? = nil) throws -> String {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }

        let templateNamePtr = name == nil ? nil : strdup(name!)
        defer { free(templateNamePtr) }

        var outTemplatePtr: UnsafePointer<CChar>? = nil
        let result = ml_vlm_get_chat_template(handle, templateNamePtr, &outTemplatePtr)
        if result < 0 {
            throw VLMError.getTemplateFailed(result)
        }
        guard let outTemplate = outTemplatePtr else { return "" }
        return String(cString: outTemplate)
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
        config: GenerationConfig = .default,
        onToken: @escaping (String) -> Bool
    ) throws -> String {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }

        let promptPtr = strdup(prompt)!
        defer { free(promptPtr) }

        let holder = TokenCallbackWrapper(onToken)
        let userData = Unmanaged.passUnretained(holder).toOpaque()
        
        var fullTextPtr: UnsafeMutablePointer<CChar>? = nil
        let result = config.withUnsafePointerC { configPtr in
            return ml_vlm_generate_stream(
                handle,
                promptPtr,
                configPtr,
                tokenCallback,
                userData,
                &fullTextPtr
            )
        }
        if result < 0 {
            throw VLMError.generationStreamFailed(result)
        }

        guard let fullText = fullTextPtr else { return "" }
        defer { free(fullText) }
        return String(cString: fullText)
    }

    @NexaAIActor
    public func applyChatTemplate(messages: [ChatMessage]) throws -> String {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
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
        let result = ml_vlm_apply_chat_template(handle, mlMessages, count, &outTextPtr)

        if result < 0 {
            throw VLMError.applyChatTemplateFailed(result)
        }

        guard let outText = outTextPtr else { return "" }
        defer { free(outText) }

        return String(cString: outText)
    }

    public func embed(texts: [String]) throws -> [Float] {
        guard handle != nil else {
            throw VLMError.modelLoadingFailed
        }

        guard !texts.isEmpty else {
            return []
        }

        let textsMutablePtr = texts.map { strdup($0) }
        defer {
            for text in textsMutablePtr {
                free(text)
            }
        }

        var textsPtr = textsMutablePtr.map { UnsafePointer($0) }
        let textCount = Int32(texts.count)
        var outEmbeddingsPtr: UnsafeMutablePointer<Float>? = nil

        let result = textsPtr.withUnsafeMutableBufferPointer { buffer in
            return ml_vlm_embed(handle, buffer.baseAddress, textCount, &outEmbeddingsPtr)
        }

        if result < 0 {
            throw VLMError.embeddingFailed(result)
        }
        guard let outEmbeddings = outEmbeddingsPtr else { return [] }
        defer { free(outEmbeddings) }

        let buffer = UnsafeBufferPointer(start: outEmbeddings, count: Int(result))
        return Array(buffer)
    }

    public func reset() {
        guard let handle else {
            return
        }
        ml_vlm_reset(handle)
    }

    @discardableResult
    public func getProfileData() -> ProfilingData? {
        if let handle {
            var out = ml_ProfilingData()
            ml_vlm_get_profiling_data(handle, &out)
            return .init(from: out)
        }
        return nil
    }

    public func unload() {
        if let handle {
            ml_vlm_destroy(handle)
            ml_deinit()
        }
        handle = nil
    }

    deinit {
        if let handle {
            ml_vlm_destroy(handle)
            ml_deinit()
        }
    }
}

extension VLM: Model {

    public func loadModel(modelPath: String, mmprojPath: String?, config: ModelConfig?) throws {
        self.modelPath = modelPath
        self.mmprojPath = mmprojPath
        self.contextLength = config?.nCtx ?? 256
        unload()
        try loadModel()
    }

    public var type: ModelType { .vlm }

}

extension VLM {

    public enum VLMError: LocalizedError {
        case modelLoadingFailed
        case encodingFailed(Int32)
        case decodingFailed(Int32)
        case generateFailed(Int32)
        case getTemplateFailed(Int32)
        case generationStreamFailed(Int32)
        case applyChatTemplateFailed(Int32)
        case embeddingFailed(Int32)

        public var errorDescription: String? {
            switch self {
            case .modelLoadingFailed:
                return "Model loading failed"
            case .encodingFailed(let code),
                 .decodingFailed(let code),
                 .generateFailed(let code),
                 .getTemplateFailed(let code),
                 .applyChatTemplateFailed(let code),
                 .embeddingFailed(let code),
                 .generationStreamFailed(let code):
                if let errorMessage = ml_get_error_message(ml_ErrorCode(rawValue: code)) {
                    let result = String(cString: errorMessage)
                    return result
                }
                return "unknow error, code: \(code)"
            }
        }
    }

}
