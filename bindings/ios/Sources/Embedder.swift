import Foundation
import NexaBridge

private typealias ml_Embedder = OpaquePointer

final public class Embedder {
    private let handle: ml_Embedder?
    public init(
        modelPath: String,
        tokenizerPath: String? = nil,
        device: String? = nil
    ) throws {
        let model = strdup(modelPath)!
        defer { free(model) }

        let tokenizer = tokenizerPath != nil ? strdup(tokenizerPath!)! : nil
        defer { free(tokenizer) }

        let dev = device != nil ? strdup(device!) : nil
        defer { free(dev) }

        handle = ml_embedder_create(model, tokenizerPath, dev)
        guard handle != nil else {
            throw EmbedderError.creationFailed
        }
    }

    public func embed(texts: [String], config: EmbeddingConfig) throws -> [Float] {
        guard let handle else {
            throw EmbedderError.creationFailed
        }

        let textCount = Int32(texts.count)
        var outPtr: UnsafeMutablePointer<Float>? = nil
        let result = texts.withUnsafeMutableBufferPointerC { buffer in
            return config.withUnsafePointerC { mConfig in
                return ml_embedder_embed(handle, buffer, textCount, mConfig, &outPtr)
            }
        }

        if result <= 0 {
            throw EmbedderError.embedFailed(result)
        }
        
        guard let out = outPtr else { return [] }
        defer { free(out) }

        let buffer = UnsafeBufferPointer(start: out, count: Int(result))
        return Array(buffer)
    }

    public func embeddingDim() throws -> Int32 {
        guard let handle else {
            throw EmbedderError.creationFailed
        }

        let result = ml_embedder_embedding_dim(handle)
        if result <= 0 {
            throw EmbedderError.embeddingDim(result)
        }
        return result
    }

    public func setloRA(_ loRAId: Int32) throws {
        guard let handle else {
            throw EmbedderError.creationFailed
        }
        ml_embedder_set_lora(handle, loRAId)
    }

    public func removeLoRA(by loRAId: Int32) throws {
        guard let handle else {
            throw EmbedderError.creationFailed
        }
        ml_embedder_remove_lora(handle, loRAId)
    }

    public func addLoRA(_ path: String) throws {
        guard let handle else {
            throw EmbedderError.creationFailed
        }
        let pathPtr = strdup(path)!
        defer {
            free(pathPtr)
        }
        let result = ml_embedder_add_lora(handle, pathPtr)
        if result < 0 {
            throw EmbedderError.addLoRAFailed(result)
        }
    }

    public func listLoRAs() throws -> [Int32] {
        guard let handle else {
            throw EmbedderError.creationFailed
        }

        var outPtr: UnsafeMutablePointer<Int32>? = nil
        let result = ml_embedder_list_loras(handle, &outPtr)
        if result < 0 {
            throw EmbedderError.listLoRAsFailed(result)
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
            ml_embedder_get_profiling_data(handle, &out)
            return .init(from: out)
        }
        return nil
    }

    deinit {
        if let handle {
            ml_embedder_destroy(handle)
        }
    }
}

extension Embedder {

    public enum EmbedderError: LocalizedError {
        case creationFailed
        case embedFailed(Int32)
        case embeddingDim(Int32)
        case addLoRAFailed(Int32)
        case listLoRAsFailed(Int32)

        public var errorDescription: String? {
            switch self {
            case .creationFailed:
                return "Model loading failed"
            case .embedFailed(let code),
                 .embeddingDim(let code),
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

fileprivate extension EmbeddingConfig {

    func withUnsafePointerC<T>(_ body: (UnsafePointer<ml_EmbeddingConfig>) throws -> T) rethrows -> T {
        let methodCString = strdup(normalizeMethod.rawValue)
        var cStruct = ml_EmbeddingConfig(
            batch_size: batchSize,
            normalize: normalize,
            normalize_method: methodCString
        )
        defer { free(methodCString) }
        return try withUnsafePointer(to: &cStruct, body)
    }

    func withUnsafeMutablePointerC<T>(_ body: (UnsafeMutablePointer<ml_EmbeddingConfig>) throws -> T) rethrows -> T {
        let methodCString = strdup(normalizeMethod.rawValue)
        var cStruct = ml_EmbeddingConfig(
            batch_size: batchSize,
            normalize: normalize,
            normalize_method: methodCString
        )
        defer { free(methodCString) }
        return try withUnsafeMutablePointer(to: &cStruct, body)
    }
}
