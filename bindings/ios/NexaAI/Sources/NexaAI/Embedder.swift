import Foundation
import NexaBridge

private typealias ml_Embedder = OpaquePointer

public struct EmbedResult {
    public let embeddings: [[Float]]
    public let profileData: ProfileData
}

final public class Embedder {
    private var handle: ml_Embedder?
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

        var input = ml_EmbedderCreateInput(
            model_name: nil,
            model_path: model,
            tokenizer_path: tokenizer,
            plugin_id: plugin_id(),
            device_id: dev
        )

        ml_register_plugin(plugin_id, createLlamaPlugin)
        ml_init()
        let result = ml_embedder_create(&input, &handle)
        if result != ML_SUCCESS.rawValue {
            ml_deinit()
            throw EmbedderError.creationFailed(result)
        }
    }

    public func embed(inputIds: [[Int32]], config: EmbeddingConfig) throws -> EmbedResult {
        var output = ml_EmbedderEmbedOutput()
        let rowCount = inputIds.count
        let rowLengths = inputIds.map { Int32($0.count) }

        var cInputIds: [UnsafePointer<Int32>?] = []
        var buffers: [UnsafeBufferPointer<Int32>] = []

        for inputId in inputIds {
            inputId.withUnsafeBufferPointer { buf in
                buffers.append(buf)
                cInputIds.append(buf.baseAddress)
            }
        }

        let result = config.withUnsafePointerC { mConfig in
            return cInputIds.withUnsafeMutableBufferPointer { inputIdsBuffer in
                return rowLengths.withUnsafeBufferPointer { rowLengths in
                    var input = ml_EmbedderEmbedInput(texts: nil, text_count: 0, config: mConfig, input_ids_2d: inputIdsBuffer.baseAddress, input_ids_row_lengths: rowLengths.baseAddress, input_ids_row_count: Int32(rowCount))
                    return ml_embedder_embed(handle, &input, &output)
                }
            }
        }

        if result != ML_SUCCESS.rawValue {
            throw EmbedderError.embedFailed(result)
        }

        guard let out = output.embeddings else { return .init(embeddings: [], profileData: .init(from: output.profile_data)) }
        defer { free(out) }
        let dim = try Int(embeddingDim())
        let count = Int(output.embedding_count)
        var embeddings = [[Float]]()
        let embeddings1D = Array(UnsafeBufferPointer(start: out, count: dim * count))
        for i in 0..<count {
            let start = i * dim
            let end = start + dim - 1
            let sub = embeddings1D[start...end]
            embeddings.append(Array(sub))
        }
        return .init(embeddings: embeddings, profileData: ProfileData(from: output.profile_data))
    }

    public func embed(texts: [String], config: EmbeddingConfig) throws -> EmbedResult {
        var output = ml_EmbedderEmbedOutput()

        let textCount = Int32(texts.count)
        let result = texts.withUnsafeMutableBufferPointerC { buffer in
            return config.withUnsafePointerC { mConfig in
                var input = ml_EmbedderEmbedInput(texts: buffer, text_count: textCount, config: mConfig, input_ids_2d: nil, input_ids_row_lengths: nil, input_ids_row_count: 0)
                return ml_embedder_embed(handle, &input, &output)
            }
        }

        if result != ML_SUCCESS.rawValue {
            throw EmbedderError.embedFailed(result)
        }

        guard let out = output.embeddings else { return .init(embeddings: [], profileData: .init(from: output.profile_data)) }
        defer { free(out) }

        let dim = try Int(embeddingDim())
        let count = Int(output.embedding_count)
        var embeddings = [[Float]]()
        let embeddings1D = Array(UnsafeBufferPointer(start: out, count: dim * count))
        for i in 0..<count {
            let start = i * dim
            let end = start + dim - 1
            let sub = embeddings1D[start...end]
            embeddings.append(Array(sub))
        }
        return .init(embeddings: embeddings, profileData: ProfileData(from: output.profile_data))
    }

    public func embeddingDim() throws -> Int32 {
        var output = ml_EmbedderDimOutput()
        let result = ml_embedder_embedding_dim(handle, &output)
        if result != ML_SUCCESS.rawValue {
            throw EmbedderError.embeddingDim(result)
        }
        return output.dimension
    }

    deinit {
        if let handle {
            ml_embedder_destroy(handle)
            ml_deinit()
        }
        handle = nil
    }
}

extension Embedder {

    public enum EmbedderError: LocalizedError {
        case creationFailed(Int32)
        case embedFailed(Int32)
        case embeddingDim(Int32)

        public var errorDescription: String? {
            switch self {
            case .creationFailed(let code),
                 .embedFailed(let code),
                 .embeddingDim(let code):
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
