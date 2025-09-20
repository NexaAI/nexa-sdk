
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
