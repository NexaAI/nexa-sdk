import SwiftUI
import NexaAI

@MainActor
@Observable
class LLMViewModel {
    var logVM: LogViewModel = .init(text: "\nClick ellipsis button to test\n")

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
            await llmLlama.reset()
            logVM.append("\n-------- Begin Generate(multi-round) ------\n")

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
                    let stream = try await llmLlama.generateAsyncStream(messages: messages)
                    logVM.append("-----------------------------")
                    logVM.append("User: \(userMsg)")
                    logVM.append("AI:", enableEnter: false)
                    var response = ""
                    for try await token in stream {
                        logVM.append("\(token)", enableEnter: false)
                        response += token
                    }
                    logVM.append("\n")
                    messages.append(.init(role: .assistant, content: response))
                    logVM.append("Profile Data: ")
                    logVM.append(await llmLlama.lastProfileData?.description ?? "")
                } catch {
                    print(error)
                }
            }

            logVM.append("\n-------- End Generate ------\n")
        }
    }

    func reset() {
        Task {
           await llmLlama?.reset()
        }
    }

    func load() {
        Task {
            logVM.append("\n-------- Begin Load ------\n")
            logVM.append(options.modelPath)
            do {
                llmLlama = LLMLlama()
                try await llmLlama?.load(options)
            } catch {
                print(error)
            }
            logVM.append("\n-------- End Load ------\n")
        }
    }
}

struct LLMView: View {

    @State var vm: LLMViewModel = .init()
    @State private var fileImporter: AnyFileImporter?

    var body: some View {
        VStack(spacing: 16) {
            LogView(vm: vm.logVM)
        }
        .buttonStyle(.bordered)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Menu {
                    Button("load") { loadModelFile() }
                    Button("unload") { vm.unload() }
                    Button("chat") { vm.generate() }
                    Button("reset") { vm.reset() }
                } label: {
                    Image(systemName: "ellipsis.circle")
                }
            }
        }
        .fileImpoter($fileImporter)
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
