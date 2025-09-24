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
            batchSize: 1,
            normalize: true,
            normalizeMethod: .l2
        )

        let result = try embedder.embed(texts: texts, config: cfg)
        let embeddings = result.embeddings
        #expect(embeddings.count == texts.count)

        let embedding = embeddings[0]

        let count = Float(embedding.count)
        let mean = (embedding.reduce(0.0, +)) / count

        let variance =
        (embedding.map {
            let diff = mean - $0
            return diff * diff
        }
            .reduce(0.0, +)) / count

        let std = sqrt(variance)
        print(
            "Embedding stats: min=\(embedding.min()!), max=\(embedding.max()!), mean=\(mean), std=\(std)"
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

    @Test func testEmbeddingSearch() async throws {
        let embedder = try await createEmbedder()
        let documents = ["The cat sat on the mat.",
                         "A dog barked at the mailman.",
                         "Quantum physics is a branch of science.",
                         "I love eating pizza on weekends.",
                         "Machine learning enables computers to learn from data."]
        let query = "Tell me about AI and computers"
        var searchEngine = EmbeddingSearch(embedder: embedder)
        try searchEngine.addDocuments(documents)

        let results = try searchEngine.search(query: query)
        print("--------------------------")
        print("query: ", query)
        print("Top3: ")
        for (doc, score) in results {
            print("\(score), \(doc)")
        }
        print("--------------------------")

    }
}

struct EmbeddingSearch {
    struct Document {
        let text: String
        let embedding: [Float]
    }

    private let embedder: Embedder
    private var documents: [Document] = []

    init(embedder: Embedder) {
        self.embedder = embedder
    }

    mutating func addDocuments(_ texts: [String]) throws {
        let config = EmbeddingConfig(batchSize: Int32(texts.count), normalize: true, normalizeMethod: .l2)
        let embeddings = try embedder.embed(texts: texts, config: config).embeddings
        for (idx, embedding) in embeddings.enumerated() {
            documents.append(Document(text: texts[idx], embedding: embedding))
        }
    }

    func search(query: String, topK: Int = 3) throws -> [(String, Float)] {
        let config = EmbeddingConfig(batchSize: 1, normalize: true, normalizeMethod: .l2)
        let queryEmbeddings = try embedder.embed(texts: [query], config: config).embeddings

        var results: [(String, Float)] = []
        for doc in documents {
            let score = cosineSimilarity(queryEmbeddings[0], doc.embedding)
            results.append((doc.text, score))
        }

        return results.sorted { $0.1 > $1.1 }.prefix(topK).map { $0 }
    }

    private func cosineSimilarity(_ a: [Float], _ b: [Float]) -> Float {
        precondition(a.count == b.count)
        let dot = zip(a, b).map(*).reduce(0, +)
        let normA = sqrt(a.map { $0 * $0 }.reduce(0, +))
        let normB = sqrt(b.map { $0 * $0 }.reduce(0, +))
        return dot / (normA * normB)
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
