
import SwiftUI
import NexaAI

class EmbedderViewModel {
    var embedder: Embedder?

    func load(_ modelPath: String)  {
        do {
            embedder = try Embedder(modelPath: modelPath)
        } catch {
            print(error)
        }
    }
    func dim() {
        guard let embedder else {
            return
        }
        do {
            print("===> Test embedding dimension")
            let dim = try embedder.embeddingDim()
            print("Embedding dimension: \(dim)")
        } catch {
            print(error)
        }
    }

    func embed()  {
        guard let embedder else {
            return
        }
        do {
            let texts = [
                "Hello, this is a test sentence.",
                "Another test sentence for embedding.",
                "Third sentence to test batch processing."
            ]
            let cfg = EmbeddingConfig(batchSize: 2, normalize: true, normalizeMethod: .l2)
            let result = try embedder.embed(texts: texts, config: cfg)

            let embeddings = result.embeddings
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
        } catch {
            print(error)
        }
    }
}

struct EmbedderView: View {

    @State var vm: EmbedderViewModel = .init()
    @State private var fileImporter: AnyFileImporter?

    var body: some View {
        VStack(spacing: 16) {
            Button("load") {
                loadModelFile()
            }

            Button("embed") {
                vm.embed()
            }

            Button("dim") {
                vm.dim()
            }
        }
        .buttonStyle(.bordered)
        .fileImpoter($fileImporter)
    }

    func loadModelFile() {
        fileImporter = .init(contentType: [.item], onCompletion: { result in
            if let url = try? result.get() {
                do {
                    _ = url.startAccessingSecurityScopedResource()
                    let modelPath = try FileManager.copyToDocumentsDirectory(from: url).path()
                    url.stopAccessingSecurityScopedResource()
                    vm.load(modelPath)
                } catch {
                    print(error)
                }

            }
        })
    }
}

#Preview {
    EmbedderView()
}
