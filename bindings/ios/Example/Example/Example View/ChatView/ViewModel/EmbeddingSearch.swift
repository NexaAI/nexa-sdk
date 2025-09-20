import Foundation
import NexaAI

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
