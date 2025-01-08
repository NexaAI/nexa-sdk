import XCTest
import Foundation
@testable import NexaSwift

final class OmniVLMWrapperTests: XCTestCase {
    
    func testInferenceStreaming() async throws {
        let modelPath = "/Users/liute/Downloads/model-q4_0.gguf"
        let projectorModelPath = "/Users/liute/Downloads/projector-q4_0.gguf"
        let omniVLMVersion = "vlm-81-instruct"
        let prompt = "describe the image."
        let imagePath = "/Users/liute/Downloads/cat.png"
        
        guard let inference = NexaOmniVlmInference(
            modelPath: modelPath,
            projectorModelPath: projectorModelPath,
            version: omniVLMVersion
        ) else {
            XCTFail("Failed to initialize NexaOmniVlmInference")
            return
        }
        
        var results: [String] = []
        
        for try await response in await inference.inferenceStreaming(prompt: prompt, imagePath: imagePath) {
            print(response, terminator: "")
            results.append(response)
        }
        
        XCTAssertFalse(results.isEmpty, "Streaming results should not be empty")
    }
}
