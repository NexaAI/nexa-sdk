import Testing
import Foundation
@testable import NexaAI

struct LLamaEmbedderTest {

    init() {
        NexaSdk.install()
    }

    @Test func testLLamaEmbedderLoad() async throws {
        let modelPath = try modelPath(of: "jina-embeddings-v2-small-en-Q4_K_M")
        let _ = try Embedder(modelPath: modelPath)
    }

    @Test func testLLamaEmbedderDim() async throws {
        let modelPath = try modelPath(of: "jina-embeddings-v2-small-en-Q4_K_M")
        let embedder = try Embedder(modelPath: modelPath)
        let dim = try embedder.embeddingDim()
        print("Embedding dimension: \(dim)")
    }

    @Test func testLLamaEmbedderEmbed() async throws {
        let modelPath = try modelPath(of: "jina-embeddings-v2-small-en-Q4_K_M")
        let embedder = try Embedder(modelPath: modelPath)
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

}


struct LlamaDeviceListTest {

    @Test func testGetDeviceList() {
        if let deviceList = NexaSdk.getLlamaDeviceList() {
            print(deviceList)
        }
    }

}
