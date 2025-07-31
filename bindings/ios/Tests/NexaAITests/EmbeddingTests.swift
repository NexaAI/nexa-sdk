import Foundation
import Testing

@testable import NexaAI

@Test func embeddingTest() async throws {
    print("===> Test Embedder Creation")

    let modelPath = try resoucePath(of: "jina-embeddings-v2-base-en-Q4_0")
    let embedder = try Embedder(modelPath: modelPath)

    print("===> Test embedding dimension")
    let dim = try embedder.embeddingDim()
    print("Embedding dimension: \(dim)")

    print("===> Test Single Text Embedding")
    let text = "Hello World!"
    let cfg = EmbeddingConfig(batchSize: 32, normalize: true, normalizeMethod: .l2)
    let embeddings = try embedder.embed(texts: [text], config: cfg)

    print("Embedding generated successfully")
    print("Embedding dimension: \(embeddings.count)")
    print(embeddings.prefix(20), " ...")

    print("Calculate and print stats")
    let count = Float(embeddings.count)
    let mean = (embeddings.reduce(0.0, +)) / count

    let variance =
        (embeddings.map {
            let diff = mean - $0
            return diff * diff
        }
        .reduce(0.0, +)) / count

    let std = sqrt(variance)
    print(
        "embedding stats: min=\(embeddings.min()!), max=\(embeddings.max()!), mean=\(mean), std=\(std)"
    )
}
