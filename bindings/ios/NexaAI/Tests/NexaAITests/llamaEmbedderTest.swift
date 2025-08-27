import Testing
import Foundation
@testable import NexaAI

struct LLamaEmbedderTest {

    init() {
        NexaSdk.install()
    }

    func createEmbedder(name: String = "jina-embeddings-v2-small-en-Q4_K_M") async throws -> Embedder {
        let modelPath = try modelPath(of: name)
        return try Embedder(modelPath: modelPath)
    }

    @Test func testCreateEmbedder() async throws {
        let _ = try await createEmbedder()
    }

    @Test func testEmbedderCreation() async throws {
        let embedder = try await createEmbedder()
        let dim = try embedder.embeddingDim()
        print("Embedding dimension: \(dim)")
    }

    @Test func testEmbedderSingleText() async throws {
        let embedder = try await createEmbedder()
        let texts = ["🥳 🎂 Once upon a time"]

        let cfg = EmbeddingConfig(
            batchSize: 32,
            normalize: true,
            normalizeMethod: .l2
        )

        let result = try embedder.embed(texts: texts, config: cfg)
        let embeddings = result.embeddings
        #expect(embeddings.count == texts.count)

        let dim = try embedder.embeddingDim()
        let expectedTotalFloats = dim * Int32(embeddings.count)
        print(expectedTotalFloats)

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
            "Embedding stats: min=\(embeddings.min()!), max=\(embeddings.max()!), mean=\(mean), std=\(std)"
        )
        print("ProfileData: \n", result.profileData)
    }

    @Test func testEmbedderBatchProcessing() async throws {
        let embedder = try await createEmbedder()
        let texts = [
            "Hello world", "Good morning",
            "Machine learning is fascinating",
            "Natural language processing"
        ]
        let cfg = EmbeddingConfig(batchSize: 4, normalize: true, normalizeMethod: .l2)
        let result = try embedder.embed(texts: texts, config: cfg)
        let embeddings = result.embeddings
        print("Texts: ", texts)
        print("Total embeddings: \(embeddings.count)")
        print("Embeddings:", embeddings.prefix(20))
        print("ProfileData: \n", result.profileData)
    }
}


struct NexaSdkTest {

    @Test func testGetDeviceList() {
        print(NexaSdk.getLlamaDeviceList())
    }

    @Test func testGetVersion() {
        print(NexaSdk.version)
    }
}
