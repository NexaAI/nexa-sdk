// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

/*

/** Reranking configuration */
typedef struct {
    int32_t     batch_size;       /* Processing batch size */
    bool        normalize;        /* Whether to normalize scores */
    const char* normalize_method; /* Normalization: "softmax", "min-max", "none" */
} ml_RerankConfig;

*/
import Foundation

public struct RerankConfig {
    public var batchSize: Int32
    public var normalize: Bool
    public var normalizeMethod: NormalizeMethod
}

extension RerankConfig {
    public enum NormalizeMethod: String, CaseIterable {
        case softmax
        case minMax = "min-max"
        case none
    }
}
