import SwiftUI
import NexaAI

class LLMViewModel {
    var llmLlama: LLMLlama?
    var options: ModelOptions = .init(modelPath: "")

    func unload() {
        llmLlama = nil
    }

    func generate()  {
        guard let llmLlama else {
            return
        }
        Task {
            let system = "You are a helpful AI assistant "
            let userMsgs = [
                "What is 1+1?",
                "How about 2+2?",
                "What is n+n?",
                "Tell me a long stroy, about 100 words",
                "How to learn English"
            ]
            var messages = [ChatMessage]()
            messages.append(.init(role: .system, content: system))
            for userMsg in userMsgs {
                do {
                    messages.append(.init(role: .user, content: userMsg))
                    let stream = try await llmLlama.generationAsyncStream(messages: messages)
                    print("-----------------------------")
                    var response = ""
                    for try await token in stream {
                        print(token, terminator: "")
                        response += token
                    }
                    print("\n")
                    messages.append(.init(role: .assistant, content: response))
                    print(await llmLlama.lastProfileData?.stopReason ?? "")
                } catch {
                    print(error)
                }
            }
        }
    }

    func reset() {
        Task {
           await llmLlama?.reset()
        }
    }

    func load() {
        Task {
            do {
                llmLlama = LLMLlama()
                try await llmLlama?.load(options)
            } catch {
                print(error)
            }
        }
    }
}

struct LLMView: View {

    @State var vm: LLMViewModel = .init()
    @State private var fileImporter: AnyFileImporter?

    var body: some View {
        VStack(spacing: 16) {
            Button("load") {
                loadModelFile()
            }

            Button("unload") {
                vm.unload()
            }

            Button("chat") {
                vm.generate()
            }
            Button("reset") {
                vm.reset()
            }
        }
        .padding()
        .fileImpoter($fileImporter)
        .buttonStyle(.bordered)
    }

    func loadModelFile() {
        fileImporter = .init(contentType: [.item], onCompletion: { result in
            if let url = try? result.get() {
                do {
                    _ = url.startAccessingSecurityScopedResource()
                    let modelPath = try FileManager.copyToDocumentsDirectory(from: url).path()
                    url.stopAccessingSecurityScopedResource()
                    loadModel(modelPath)
                } catch {
                    print(error)
                }

            }
        })
    }

    func loadModel(_ modelPath: String) {
        Task {
            await MainActor.run {
                vm.options.modelPath = modelPath
                vm.load()
            }
        }
    }
}
