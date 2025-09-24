import Foundation
import NexaBridge

private typealias ml_Reranker = OpaquePointer

public struct RerankerResult {
    let scores: [Float]
    let profileData: ProfileData
}

public final class Reranker {
    private var handle: ml_Reranker?

    public init(
        modelPath: String,
        tokenizerPath: String? = nil,
        deviceId: String? = nil
    ) throws {
        let model = strdup(modelPath)!
        defer { free(model) }

        let tokenizer = tokenizerPath != nil ? strdup(tokenizerPath!)! : nil
        defer { free(tokenizer) }

        let dev = deviceId != nil ? strdup(deviceId!) : nil
        defer { free(dev) }

        var input = ml_RerankerCreateInput(
            model_name: nil,
            model_path: model,
            tokenizer_path: tokenizer,
            plugin_id: plugin_id(),
            device_id: dev
        )

        let result = ml_reranker_create(&input, &handle)
        if result < 0 || handle == nil {
            throw RerankerError.creationFailed(result)
        }
    }

    @NexaAIActor
    public func rerank(_ query: String, documents: [String], config: RerankConfig) throws -> RerankerResult {

        var output = ml_RerankerRerankOutput()
        let count = Int32(documents.count)
        let result = query.withCString { cQuery in
            return documents.withUnsafeMutableBufferPointerC { cDocuments in
                config.withUnsafePointerC { cConfig in
                    var input = ml_RerankerRerankInput(
                        query: cQuery,
                        documents: cDocuments,
                        documents_count: count,
                        config: cConfig
                    )
                    return ml_reranker_rerank(handle, &input, &output)
                }
            }
        }

        if result < 0 {
            throw RerankerError.rerankFailed(result)
        }

        guard let scores = output.scores else {
            return .init(scores: [], profileData: .init(from: output.profile_data))
        }
        defer { ml_free(scores) }

        let buffer = UnsafeBufferPointer(start: scores, count: Int(output.score_count))
        return .init(scores: Array(buffer), profileData: .init(from: output.profile_data))
    }

    deinit {
        if let handle {
            ml_reranker_destroy(handle)
        }
        handle = nil
    }
}

extension Reranker {
    public enum RerankerError: LocalizedError {
        case creationFailed(Int32)
        case rerankFailed(Int32)

        public var errorDescription: String? {
            switch self {
            case .creationFailed(let code),
                 .rerankFailed(let code):
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
