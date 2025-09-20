
import SwiftUI
import NexaAI

@Observable
@MainActor
class LogViewModel {
    var text = ""
    init(text: String = "") {
        self.text = text
    }

    func append(_ str: String, enableEnter: Bool = true) {
        text += str
        if enableEnter {
            text += "\n"
        }
    }
}

struct LogView: View {
    @State var vm: LogViewModel

    var body: some View {
        VStack(spacing: 16) {
            TextEditor(text: $vm.text)
                .scrollContentBackground(.hidden)
                .frame(maxWidth: .infinity, maxHeight: .infinity)
                .padding(.horizontal, 16)
        }
    }
}

@Observable
@MainActor
class EmbedderViewModel {
    var logVM: LogViewModel = .init(text: "\nClick ellipsis button to test\n")
    var embedder: Embedder?

    func load(_ modelPath: String)  {
        logVM.append("----------- Load Begin ----------\n")
        logVM.append("Load Model From: \(modelPath)")
        do {
            embedder = try Embedder(modelPath: modelPath)
        } catch {
            print(error)
            logVM.append("\(error.localizedDescription)")
        }
        logVM.append("----------- Load End -----------\n")
    }

    func dim() {
        guard let embedder else {
            return
        }
        do {
            logVM.append("----------- Embedding dimension ----------\n")
            let dim = try embedder.embeddingDim()
            logVM.append("Embedding dimension: \(dim)")
        } catch {
            print(error)
        }
    }

    func embed()  {
        guard let embedder else {
            return
        }

        logVM.append("----------- Begin Embed ----------\n")
        do {
            let texts = [
                "Hello, this is a test sentence.",
                "Another test sentence for embedding.",
                "Third sentence to test batch processing."
            ]
            logVM.append("")
            logVM.append("Embed Text: \(texts)")
            logVM.append("")

            let cfg = EmbeddingConfig(batchSize: Int32(texts.count), normalize: true, normalizeMethod: .l2)
            let result = try embedder.embed(texts: texts, config: cfg)

            for embedding in result.embeddings {
                logVM.append("Embeding result(prefix 20): \(embedding.prefix(20)), ...")

                logVM.append("Calculate and print stats")
                let count = Float(embedding.count)
                let mean = (embedding.reduce(0.0, +)) / count

                let variance =
                    (embedding.map {
                        let diff = mean - $0
                        return diff * diff
                    }
                    .reduce(0.0, +)) / count

                let std = sqrt(variance)
                logVM.append(
                    "Embedding stats: min=\(embedding.min()!), max=\(embedding.max()!), mean=\(mean), std=\(std)"
                )
            }

        } catch {
            print(error)
            logVM.append("Embed Error: \(error)")
        }

        logVM.append("----------- End Embed ----------\n")
    }

    func search() {
        guard let embedder else {
            return
        }
        logVM.append("----------- Begin Search ----------\n")
        do {
            let documents = ["The cat sat on the mat.",
                             "A dog barked at the mailman.",
                             "Quantum physics is a branch of science.",
                             "I love eating pizza on weekends.",
                             "Machine learning enables computers to learn from data."]
            let query = "Tell me about AI and computers"
            var searchEngine = EmbeddingSearch(embedder: embedder)
            try searchEngine.addDocuments(documents)

            let results = try searchEngine.search(query: query)
            logVM.append("Search Result: ")
            logVM.append("Query: \(query)")
            logVM.append("Top3: ")
            for (doc, score) in results {
                logVM.append("\(score), \(doc)")
            }
        } catch {
            logVM.append("Search Error: \(error)")
        }

        logVM.append("----------- End Search ----------\n")
    }
}

struct EmbedderView: View {

    @State var vm: EmbedderViewModel = .init()
    @State private var fileImporter: AnyFileImporter?
    @State var title: String = "Embedder Example"
    var body: some View {
        VStack(spacing: 16) {
            LogView(vm: vm.logVM)
        }
        .buttonStyle(.bordered)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Menu {
                    Button("load") { loadModelFile() }
                    Button("embed") { vm.embed() }
                    Button("dim") { vm.dim() }
                    Button("search") { vm.search() }
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
        .navigationTitle(title)
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
                    title = url.lastPathComponent
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
