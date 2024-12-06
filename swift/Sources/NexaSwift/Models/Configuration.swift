import Foundation
import llama

public struct Configuration {
    public var nTokens:Int32
    public var embd: Int32
    public var nSeqMax: Int32
    public var seed: UInt32
    public var topK: Int32
    public var topP: Float
    public var nCTX: Int
    public var temperature: Float
    public var maxNewToken: Int
    public var batchSize: Int
    public var stopTokens: [String]

    public init(
        nTokens:Int32 = 2048,
        embd:Int32 = 512,
        nSeqMax:Int32 = 2,
        seed: UInt32 = 1234,
        topK: Int32 = 50,
        topP: Float = 1.0,
        nCTX: Int = 2048,
        temperature: Float = 0.7,
        batchSize: Int = 2048,
        stopSequence: String? = nil,
        maxNewToken: Int = 128,
        stopTokens: [String] = []) {
            self.nTokens = nTokens
            self.embd = embd
            self.nSeqMax = nSeqMax
            self.seed = seed
            self.topK = topK
            self.topP = topP
            self.nCTX = nCTX
            self.batchSize = batchSize
            self.temperature = temperature
            self.maxNewToken = maxNewToken
            self.stopTokens = stopTokens
    }
}

extension Configuration {
    var contextParameters: ContextParameters {
        var params = llama_context_default_params()
        let processorCount = max(1, min(16, ProcessInfo.processInfo.processorCount - 2))
        params.n_ctx = max(8, UInt32(self.nCTX)) // minimum context size is 8
        params.n_threads = Int32(processorCount)
        params.n_threads_batch = Int32(processorCount)
        return params
    }
}
