import SwiftUI
import NexaAI

struct ContentView: View {
    // MARK: - State Properties
    
    // UI State
    @State private var messages: [Message] = []
    @State private var inputText: String = ""
    @State private var isResponding = false
    @State private var errorAlert: AnyAlert?

    // Model State
    @State private var modelManager: ModelManager = .init()
    @State private var streamingTask: Task<Void, Never>?
    @State private var useStreaming: Bool = true
    @State private var fileImporter: AnyFileImporter?
    @State private var modelName: String?
    @State private var currentResponseMessageId: String?


    var body: some View {
        NavigationStack {
            ZStack {
                ScrollViewReader { proxy in
                    ScrollView {
                        VStack {
                            ForEach(messages) { message in
                                let isResponding = isResponding && message.id == currentResponseMessageId
                                ChatBubbleView(message: message, isResponding: isResponding)
                                    .id(message.id)
                            }
                        }
                    }
                    .scrollDismissesKeyboard(.immediately)
                    .scrollIndicators(.hidden)
                    .defaultScrollAnchor(.bottom)
                    .safeAreaPadding(.bottom, 60)
                    .onChange(of: messages.last?.content) {
                        if let lastMessage = messages.last {
                            withAnimation {
                                proxy.scrollTo(lastMessage.id, anchor: .bottom)
                            }
                        }
                    }
                    .safeAreaInset(edge: .bottom, content: {
                        inputField
                    })
                }
            }
            .navigationBarTitleDisplayMode(.inline)
            .toolbar { toolbarContent }
            .fileImpoter($fileImporter)
            .anyAlert(alert: $errorAlert)
        }
    }
    
    // MARK: - Subviews
    
    /// Floating input field with send/stop button
    private var inputField: some View {
        ZStack {
            TextField("Typing prompt...", text: $inputText, axis: .vertical)
                .textFieldStyle(.plain)
                .lineLimit(1...5)
                .frame(minHeight: 22)
                .disabled(isResponding)
                .onSubmit {
                    if !inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty {
                        handleSendOrStop()
                    }
                }
                .padding(16)
            
            HStack {
                Spacer()
                Button(action: handleSendOrStop) {
                    Image(systemName: isResponding ? "stop.circle.fill" : "arrow.up.circle.fill")
                        .font(.system(size: 30, weight: .bold))
                        .foregroundStyle(isSendButtonDisabled ? Color.gray.opacity(0.6) : .primary)
                }
                .disabled(isSendButtonDisabled)
                .animation(.easeInOut(duration: 0.2), value: isResponding)
                .animation(.easeInOut(duration: 0.2), value: isSendButtonDisabled)
                .padding(.trailing, 8)
            }
        }
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(Color.white)
                .stroke(Color.gray, lineWidth: 1)
        )
        .padding(.horizontal, 16)
        .padding(.bottom, 16)
    }
    
    private var isSendButtonDisabled: Bool {
        return inputText.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty && !isResponding
    }
    
    @ToolbarContentBuilder
    private var toolbarContent: some ToolbarContent {
        ToolbarItem(placement: .principal) {
            Button(action: loadModelFile) {
                if let modelName, !modelName.isEmpty {
                    Text(modelName)
                } else {
                    Text("Selected Model")
                }
            }
            .tint(.accentColor)
        }
    }
    
    // MARK: - Model Interaction

    func loadModelFile() {
        fileImporter = .init(contentType: [.item], onCompletion: { result in
            if let url = try? result.get() {
                modelName = url.lastPathComponent
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
            do {
                let llm = LLM(modelPath: modelPath)
                try await modelManager.loadModel(llm)
            } catch {
                print(error)
            }
        }
    }

    private func handleSendOrStop() {
        if isResponding {
            stopStreaming()
        } else {
            sendMessage()
        }
    }
    
    private func sendMessage() {
        if modelManager.model == nil {
            showError(message: "load model first")
            return
        }

        if modelManager.isLoadingModel {
            showError(message: "Model is loading, please wait...")
            return
        }
        isResponding = true

        let userMessage = Message.user(inputText)
        messages.append(userMessage)
        let lastMessages = messages
        inputText = ""

        let currentResponse = Message.assistant("")
        currentResponseMessageId = currentResponse.id
        messages.append(currentResponse)
        streamingTask = Task {
            do {
                if useStreaming {
                    let stream = try await modelManager.generationAsyncStream(from: lastMessages)
                    for try await partialResponse in stream {
                        updateLastMessage(with: partialResponse, isAppend: true)
                    }
                } else {
                    let response = try await modelManager.generate(from: lastMessages)
                    updateLastMessage(with: response)
                }
            } catch is CancellationError {
                // User cancelled generation
            } catch {
                showError(message: "An error occurred: \(error.localizedDescription)")
            }
            
            isResponding = false
            streamingTask = nil
        }
    }
    
    private func stopStreaming() {
        modelManager.stopGeneration()
        streamingTask?.cancel()
    }
    
    @MainActor
    private func updateLastMessage(with text: String, isAppend: Bool = false) {
        if isAppend {
            messages[messages.count - 1].content += text
        } else {
            messages[messages.count - 1].content = text
        }
    }

    
    @MainActor
    private func showError(message: String) {
        self.errorAlert = .init(title: message)
        self.isResponding = false
    }
}

#Preview {
    ContentView()
}
