import SwiftUI
import NexaAI

enum FileType {
    case model
    case mmproj
    case image
    case audio
}

class VLMViewModel {
    var vlmLlama: VLMLlama?
    var options: ModelOptions = .init(modelPath: "", mmprojPath: "")
    var images = [String]()
    var audios = [String]()

    func setUrl(_ url: URL, fileType: FileType) {
        switch fileType {
        case .model:
            options.modelPath = url.path()
            if let mmprojPath = options.mmprojPath, !mmprojPath.isEmpty {
                load()
            }
        case .mmproj:
            options.mmprojPath = url.path()
            if !options.modelPath.isEmpty {
                load()
            }
        case .image:
            images.append(url.path())
        case .audio:
            audios.append(url.path())
        }
    }

    func unload() {
        vlmLlama = nil
    }
    
    func load() {
        Task {
            do {
                vlmLlama = VLMLlama()
                try await vlmLlama?.load(options)
            } catch {
                print(error)
            }
        }
    }

    func imageTest() {
        guard let vlmLlama else {
            return
        }
        Task {
            do {
                await vlmLlama.reset()
                var config = GenerationConfig.default
                config.imagePaths = images
                let message = ChatMessage(role: .user, content: "Describe this image", images: images)
                let stream = try await vlmLlama.generateAsyncStream(messages: [message], options: .init(config: config))
                for try await token in stream {
                    print(token, terminator: "")
                }
            } catch {
                print(error)
            }
        }
    }

    func generate()  {
        guard let vlmLlama else {
            return
        }
        Task {
            let userMsgs = [
                (prompt: "What do you see in this image?", images: images),
                (prompt: "Please repeat the number 42 three times.", images: nil),
                (prompt: "What colors do you see in this image? Please describe the color scheme.", images: images),
            ]
            var messages = [ChatMessage]()
            messages.append(.init(role: .system, content: "You are a helpful assistant that can see images."))
            for userMsg in userMsgs {
                do {
                    let user = ChatMessage(role: .user, content: userMsg.prompt, images: userMsg.images ?? [])
                    var config = GenerationConfig.default
                    config.maxTokens = 100
                    config.imagePaths = user.images
                    messages.append(user)
                    let stream = try await vlmLlama.generateAsyncStream(messages: messages, options: .init(config: config))
                    print("-----------------------------")
                    var response = ""
                    for try await token in stream {
                        print(token, terminator: "")
                        response += token
                    }
                    print("\n")
                    messages.append(.init(role: .assistant, content: response))
                    print(await vlmLlama.lastProfileData?.stopReason ?? "")
                } catch {
                    print(error)
                }
            }
        }
    }
}

struct VLMView: View {
    @State var vm: VLMViewModel = .init()
    @State private var fileImporter: AnyFileImporter?
    var body: some View {
        VStack(spacing: 16) {
            Button("load model") {
                openFile(fileType: .model)
            }

            Button("load mmproj") {
                openFile(fileType: .mmproj)
            }

            Button("unload") {
                vm.unload()
            }

            Button("load image") {
                openFile(fileType: .image)
            }

            Button("load audio") {
                openFile(fileType: .audio)
            }

            Button("reset resource") {
                vm.images = []
                vm.audios = []
            }

            Button("ImageTest") {
                vm.imageTest()
            }

            Button("chat") {
                vm.generate()
            }
        }
        .padding()
        .fileImpoter($fileImporter)
        .buttonStyle(.bordered)
    }

    func openFile(fileType: FileType) {
        fileImporter = .init(contentType: [.item], onCompletion: { result in
            if let url = try? result.get() {
                do {
                    _ = url.startAccessingSecurityScopedResource()
                    let modelPath = try FileManager.copyToDocumentsDirectory(from: url)
                    url.stopAccessingSecurityScopedResource()
                    vm.setUrl(modelPath, fileType: fileType)
                } catch {
                    print(error)
                }

            }
        })
    }
}
