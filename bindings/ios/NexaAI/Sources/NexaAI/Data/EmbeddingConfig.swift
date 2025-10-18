// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

import Foundation

public struct EmbeddingConfig {

    public var batchSize: Int32
    public var normalize: Bool
    public var normalizeMethod: NormalizeMethod

    public init(batchSize: Int32, normalize: Bool, normalizeMethod: NormalizeMethod) {
        self.batchSize = batchSize
        self.normalize = normalize
        self.normalizeMethod = normalizeMethod
    }
}

public extension EmbeddingConfig {

    enum NormalizeMethod: String {
        case l2
        case mean
        case none
    }
}
