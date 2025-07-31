
/*
/** Embedding generation configuration */
typedef struct {
    int32_t     batch_size;       /* Processing batch size */
    bool        normalize;        /* Whether to normalize embeddings */
    const char* normalize_method; /* Normalization: "l2", "mean", "none" */
} ml_EmbeddingConfig;

*/

import Foundation

public struct EmbeddingConfig {

    public var batchSize: Int32
    public var normalize: Bool
    public var normalizeMethod: NormalizeMethod
}

public extension EmbeddingConfig {

    enum NormalizeMethod: String {
        case l2
        case mean
        case none
    }
}
