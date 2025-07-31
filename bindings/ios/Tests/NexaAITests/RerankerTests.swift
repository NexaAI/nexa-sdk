import Testing
import Foundation

@testable import NexaAI

private let modelFileName = "bge-reranker-v2-m3-Q4_0"
let query = "Today busy"
let documents = [
    "What is the weather today?",
    "What is the weather tomorrow?",
    "今天有点忙"
]

@Test func rerankerTest() async throws {
    print("Test create/init")
    let modelPath = try resoucePath(of: modelFileName)

    _ = try Reranker(modelPath: modelPath)
}

@Test func basicRerankingTest() async throws {
    
    print("===> Test Basic Reranking")
    let modelPath = try resoucePath(of: modelFileName)
    
    let config = RerankConfig(batchSize: 32, normalize: false, normalizeMethod: .none)
    
    let reranker = try Reranker(modelPath: modelPath)
    let scores = try await reranker.rerank(query, documents: documents, config: config)

    print("Basic reranking completed successfully")
    print("Query: \(query)")
    print("Documents and scores:")
    
    for (idx, doc) in documents.enumerated() {
        print(" Doc \(idx + 1): \"\(doc)\"")
        print("     Raw Score: \(scores[idx])")
    }
}

@Test
func testNormalizedReranking() async throws {
    try await _testNormalizedReranking(method: .softmax)
    try await _testNormalizedReranking(method: .minMax)
}

func _testNormalizedReranking(method: RerankConfig.NormalizeMethod) async throws {
    print("===> Test \(method.rawValue) Normalized Reranking")
    
    let modelPath = try resoucePath(of: modelFileName)
    let config = RerankConfig(batchSize: 32, normalize: true, normalizeMethod: method)

    let reranker = try Reranker(modelPath: modelPath)
    let scores = try await reranker.rerank(query, documents: documents, config: config)
    print("\(method.rawValue)", scores)
    print("\(method.rawValue) normalized reranking completed successfully")
    print("Query: \(query)")
    print("Documents and normalized scores:")
    
    var scoreIndices = [(score: Float, index: Int)]()
    for (idx, score) in scores.enumerated() {
        scoreIndices.append((score, idx))
    }
    let sorted = scoreIndices.sorted { $0.score > $1.score }
    print("Ranked results:")
    
    for (i, item) in sorted.enumerated() {
        let docIdx = item.index
        let score = item.score
        print(" Rank \(i + 1): Doc \(docIdx + 1) (score: \(score))")
        print("     \"\(documents[docIdx])\"")
    }
    
    print("Calculate and print stats")
    let count = Float(scores.count)
    let mean = (scores.reduce(0.0, +))/count
    
    let variance = (scores.map {
        let diff = mean - $0
        return diff * diff
    }
        .reduce(0.0, +)) / count
    
    let std = sqrt(variance)
    print("Score statistics: min=\(scores.min()!), max=\(scores.max()!), mean=\(mean), std=\(std)")
}

@Test
func batchSizes() async throws {
    print("===> Test Different Batch Sizes")
    let modelPath = try resoucePath(of: modelFileName)
    let reranker = try Reranker(modelPath: modelPath)
    
    let batchSizes = [1, 2, 8, 16, 32]
    for size in batchSizes {
        print("Testing batch size: \(size)")
        let config = RerankConfig(batchSize: Int32(size), normalize: true, normalizeMethod: .softmax)
        let scores = try await reranker.rerank(query, documents: documents, config: config)
        print("  Batch size \(size) completed successfully")
        print("  Top document score: \(scores.max()!)")
    }
}

@Test
func testEdgeCase() async throws {
    print("===> Test edge case")
    let modelPath = try resoucePath(of: modelFileName)
    var query = ""
    let doc = ["This is a test document."]

    let reranker = try Reranker(modelPath: modelPath)
    let config = RerankConfig(batchSize: 32, normalize: false, normalizeMethod: .none)
    
    var scores = try await reranker.rerank(query, documents: doc, config: config)
    if !scores.isEmpty {
        print("Empty query test: score =\(scores[0])")
    } else {
        print("Empty query test failed (expected behavior)")
    }
    
    query = "test query"
    scores = try await reranker.rerank(query, documents: [""], config: config)
    if !scores.isEmpty {
        print("Empty document test: score =\(scores[0])")
    } else {
        print("Empty document test failed")
    }

    scores = try await reranker.rerank(query, documents: ["Single document for testing."], config: config)
    if !scores.isEmpty {
        print("Single document test: score =\(scores[0])")
    } else {
        print("Single document test failed")
    }
}
