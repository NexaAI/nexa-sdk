import Foundation
import NexaBridge

private typealias ml_Reranker = OpaquePointer

public final class Reranker {
    private let handle: ml_Reranker?

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

        handle = ml_reranker_create(model, tokenizerPath, dev)
        guard handle != nil else {
            throw RerankerError.creationFailed
        }
    }

    @NexaAIActor
    public func rerank(_ query: String, documents: [String], config: RerankConfig) throws -> [Float] {
        guard let handle else {
            throw RerankerError.creationFailed
        }

        var outPtr: UnsafeMutablePointer<Float>? = nil
        let result = query.withCString { cQuery in
            return documents.withUnsafeMutableBufferPointerC { cDocuments in
                config.withUnsafePointerC { cConfig in
                    return ml_reranker_rerank(handle, cQuery, cDocuments, Int32(documents.count), cConfig, &outPtr)
                }
            }
        }

        if result < 0 {
            throw RerankerError.rerankFailed(result)
        }

        guard let out = outPtr else {  return [] }
        defer { free(out) }

        let buffer = UnsafeBufferPointer(start: out, count: Int(result))
        return Array(buffer)
    }

    @discardableResult
    public func getProfileData() -> ProfilingData? {
        if let handle {
            var out = ml_ProfilingData()
            ml_reranker_get_profiling_data(handle, &out)
            return .init(from: out)
        }
        return nil
    }

    deinit {
        if let handle {
            ml_reranker_destroy(handle)
        }
    }
}

extension Reranker {

    public enum RerankerError: LocalizedError {
        case creationFailed
        case rerankFailed(Int32)

        public var errorDescription: String? {
            switch self {
            case .creationFailed:
                return "Model loading failed"
            case .rerankFailed(let code):
                if let errorMessage = ml_get_error_message(ml_ErrorCode(rawValue: code)) {
                    let result = String(cString: errorMessage)
                    return result
                }
                return "unknow error, code: \(code)"
            }
        }
    }

}

fileprivate extension RerankConfig {

    func withUnsafePointerC<T>(_ body: (UnsafePointer<ml_RerankConfig>) throws -> T) rethrows -> T {
        let methodCString = strdup(normalizeMethod.rawValue)
        var cStruct = ml_RerankConfig(
            batch_size: batchSize,
            normalize: normalize,
            normalize_method: methodCString
        )
        defer { free(methodCString) }
        return try withUnsafePointer(to: &cStruct, body)
    }
}
