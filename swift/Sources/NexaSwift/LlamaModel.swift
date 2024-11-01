import Foundation
import llama

class LlamaModel {
    private let model: Model
    public var configuration: Configuration
    private let context: OpaquePointer
    private var sampler: UnsafeMutablePointer<llama_sampler>
    private var batch: Batch
    private var tokens: [Token]
    private var temporaryInvalidCChars: [CChar] = []
    private var generatedTokenAccount: Int32 = 0
    private var totalTokensProcessed: Int32 = 0
    private var ended = false
    private let n_ctx: Int32
    public var arch: String {
        return getModelDetails().arch
    }
    public var modelType: String {
        return getModelDetails().modelType
    }
    public var modelFtype: String {
        return getModelDetails().modelFtype
    }

    var shouldContinue: Bool {
        generatedTokenAccount < configuration.maxNewToken && !ended
    }
    
    public func reset() {
        generatedTokenAccount = 0
        ended = false
    }

    init(path: String, configuration: Configuration = .init()) throws {
        self.configuration = configuration
        llama_backend_init()
        llama_numa_init(GGML_NUMA_STRATEGY_DISABLED)
        var model_params = llama_model_default_params()
        #if os(iOS) || targetEnvironment(simulator)
            model_params.n_gpu_layers = 0
        #endif
        guard let model = llama_load_model_from_file(path, model_params) else {
            throw NexaSwiftError.others("Cannot load model at path \(path)")
        }
        
        self.model = model
        
        guard let context = llama_new_context_with_model(model, configuration.contextParameters) else {
            throw NexaSwiftError.others("Cannot load model context")
        }
        self.context = context
        self.n_ctx = Int32(llama_n_ctx(context))
        self.tokens = []
        self.sampler = llama_sampler_chain_init(llama_sampler_chain_default_params())
        self.batch = llama_batch_init(configuration.nTokens, 0, 1)
        try checkContextLength()
    }
    
    public func updateSampler() {
        self.sampler = llama_sampler_chain_init(llama_sampler_chain_default_params())
        llama_sampler_chain_add(sampler, llama_sampler_init_temp(configuration.temperature))
        llama_sampler_chain_add(sampler, llama_sampler_init_top_k(configuration.topK))
        llama_sampler_chain_add(sampler, llama_sampler_init_top_p(configuration.topP, 1))
        llama_sampler_chain_add(sampler, llama_sampler_init_softmax())
        llama_sampler_chain_add(sampler, llama_sampler_init_dist(configuration.seed))
    }

    private func checkContextLength() throws {
        let n_ctx_train = llama_n_ctx_train(model)
        if n_ctx > n_ctx_train {
            throw NexaSwiftError.others("Model was trained on \(n_ctx_train) context but tokens \(n_ctx) specified")
        }
    }
    
    private func getModelDetails() -> (arch: String, modelType: String, modelFtype: String) {
        let bufSize = 256
        var buf = [CChar](repeating: 0, count: bufSize)
        let result = llama_model_desc(model, &buf, bufSize)
        
        if result > 0 {
            let modelDesc = String(cString: buf)
            let components = modelDesc.components(separatedBy: " ")
            let arch = components[0] ?? "Unknown"
            let modelType = components[1] ?? "Unknown"
            let modelFtype = components[2] ?? "Unknown"
            return (arch, modelType, modelFtype)
        } else {
            return ("Unknown", "Unknown", "Unknown")
        }
    }

    func start(for prompt: String) throws {
//        print("arch: \(arch), modelType: \(modelType), modelFtype: \(modelFtype)")
        updateSampler()
        ended = false
        tokens = tokenize(text: prompt, addBos: true)

        // Check for token length
        if tokens.count > n_ctx {
            let originalCount = tokens.count
            tokens = Array(tokens.prefix(Int(n_ctx)))
            print("""
                WARNING: Input tokens (\(originalCount)) exceed context length (\(n_ctx)).
                Truncating to first \(n_ctx) tokens. Some content at the end will be ignored.
                Consider splitting your input into smaller chunks for better results.
                """)
        }

        temporaryInvalidCChars = []
        batch.clear()

        tokens.enumerated().forEach { index, token in
            batch.add(token: token, position: Int32(index), seqIDs: [0], logit: false)
        }
        batch.logits[Int(batch.n_tokens) - 1] = 1

        if llama_decode(context, batch) != 0 {
            throw NexaSwiftError.decodeError
        }
        generatedTokenAccount = 0
        totalTokensProcessed = batch.n_tokens
    }

    func `continue`() throws -> String {
        if totalTokensProcessed >= n_ctx {
            print("WARNING: Reached maximum context length (\(n_ctx)). Stopping generation.")
            temporaryInvalidCChars.removeAll()
            ended = true
            return ""
        }
        
        let newToken =  llama_sampler_sample(sampler, context, batch.n_tokens - 1)

        if llama_token_is_eog(model, newToken) {
            temporaryInvalidCChars.removeAll()
            ended = true
            return ""
        }


        let newTokenCChars = tokenToCChars(token: newToken)
        temporaryInvalidCChars.append(contentsOf: newTokenCChars)

        let newTokenStr: String
        if let validString = String(validating: temporaryInvalidCChars + [0], as: UTF8.self) {
            newTokenStr = validString
            temporaryInvalidCChars.removeAll()
        } else if let suffixIndex = temporaryInvalidCChars.firstIndex(where: { $0 != 0 }),
                  let validSuffix = String(validating: Array(temporaryInvalidCChars.suffix(from: suffixIndex)) + [0],
                                           as: UTF8.self) {
            newTokenStr = validSuffix
            temporaryInvalidCChars.removeAll()
        } else {
            newTokenStr = ""
        }

        batch.clear()
        batch.add(token: newToken, position: totalTokensProcessed, seqIDs: [0], logit: true)
        generatedTokenAccount += 1
        totalTokensProcessed += 1

        if llama_decode(context, batch) != 0 {
            throw NexaSwiftError.decodeError
        }
        return newTokenStr.filter { $0 != "\0" }
    }

    private func tokenToCChars(token: llama_token) -> [CChar] {
        var length: Int32 = 8
        var piece = Array<CChar>(repeating: 0, count: Int(length))

        let nTokens = llama_token_to_piece(model, token, &piece, length, 0, false)
        if nTokens >= 0 {
            return Array(piece.prefix(Int(nTokens)))
        } else {
            length = -nTokens
            piece = Array<CChar>(repeating: 0, count: Int(length))
            let nNewTokens = llama_token_to_piece(model, token, &piece, length, 0, false)
            return Array(piece.prefix(Int(nNewTokens)))
        }
    }

    private func tokenize(text: String, addBos: Bool) -> [Token] {
        let utf8Count = text.utf8.count
        let n_tokens = utf8Count + (addBos ? 1 : 0) + 1
        
        return Array(unsafeUninitializedCapacity: n_tokens) { buffer, initializedCount in
            initializedCount = Int(
                llama_tokenize(model, text, Int32(utf8Count), buffer.baseAddress, Int32(n_tokens), addBos, false)
            )
        }
    }

    func clear() {
        tokens.removeAll()
        temporaryInvalidCChars.removeAll()
        llama_kv_cache_clear(context)
    }

    deinit {
        llama_batch_free(batch)
        llama_free(context)
        llama_free_model(model)
        llama_backend_free()
    }
}
