import SwiftUI
import Foundation

struct ChatView: View {

    @State private var vm: ChatViewModel = .init()
    @FocusState private var isFocused: Bool
    @State private var generateViewHeight = 0.0
    @State private var lastUserMessageViewHeight = 0.0
    @State private var fileImporter: AnyFileImporter?

    var body: some View {
        ZStack(alignment: .top) {
            if vm.isLoadModelError {
                modelErrorLoadView
            } else {
                messagesSection
            }
        }
        .overlay(alignment: .top, content: {
            if vm.isLoadingModel {
                ProgressView(value: vm.modelLoadProgress)
                    .progressViewStyle(.linear)
                    .scaleEffect(y: 0.8)
                    .offset(y: -1)
            }
        })
        .safeAreaInset(edge: .bottom, content: {
            VStack(spacing: 16) {
                inputView
            }
        })
        .toolbarTitleDisplayMode(.inline)
        .toolbar { toolBar }
        .fileImpoter($fileImporter)
        .onDisappear {
            Task {
                await vm.modelManager.unload()
            }
        }
    }
    
    //MARK: - subviews
    private var messagesSection: some View {
        GeometryReader { geo in
            ScrollViewReader { proxy in
                ScrollView {
                    Text("").frame(height: 16)
                    VStack(spacing: 16) {
                        var messages = vm.messages
                        let totalCount = messages.count
                        let lastMessage = totalCount > 0 ? messages.removeLast() : nil
                        ForEach(Array(messages.enumerated()), id: \.offset) { (idx, message) in
                            ChatBubbleView(message: message)
                                .id(message.id)
                        }

                        if let lastMessage {
                            ChatBubbleView(message: lastMessage)
                                .id(lastMessage.id)
                                .onFrameChange { _, newValue in
                                    lastUserMessageViewHeight = newValue.height
                                }
                        }
                    }

                    if let currentGenerateMessge = vm.currentGenerateMessge {
                        ChatAssistantBubbleView(message: currentGenerateMessge)
                            .id(currentGenerateMessge.id)
                            .onFrameChange { _, newValue in
                                generateViewHeight = newValue.height
                            }
                            .padding(.horizontal, 16)
                    }

                    if let _ = vm.generationError {
                        Button("Error, try again") {
                            vm.regerationStream(from: vm.messages.count - 1)
                        }
                        .padding(.horizontal, 16)
                    }

                    Color.clear
                        .frame(height: bottomSpaceOfScrollView(geo))
                }
                .safeAreaPadding(.bottom, 20)
                .scrollDismissesKeyboard(.immediately)
                .defaultScrollAnchor(.top)
                .onChange(of: isFocused) { _ , newValue in
                    if newValue {
                        scrollToBottom(proxy)
                    }
                }
                .onChange(of: vm.scrollPosition) { _, newValue in
                    scrollToTop(proxy, newValue)
                }
            }
        }
    }

    private func bottomSpaceOfScrollView(_ geo: GeometryProxy) -> CGFloat {
        vm.isGenerating ? max(geo.size.height - generateViewHeight - lastUserMessageViewHeight - 16 - 8, 0) : max(geo.size.height - lastUserMessageViewHeight - 46 - 32 - 16 - 16,0)
    }

    private func scrollToTop(_ proxy: ScrollViewProxy, _ position: String?) {
        if position == nil {
            return
        }
        Task {
            try? await Task.sleep(for: .seconds(0.168))
            withAnimation {
                proxy.scrollTo(position, anchor: .top)
            }
        }
    }

    private func scrollToBottom(_ proxy: ScrollViewProxy) {
        Task {
            try? await Task.sleep(for: .seconds(0.1))
            withAnimation {
                proxy.scrollTo(vm.messages.last?.id, anchor: .bottom)
            }
        }
    }

    private var modelErrorLoadView: some View {
        GeometryReader { geo in
            ScrollView {
                Text("Model load error, please try again!")
                    .frame(height: geo.size.height)
                    .frame(maxWidth: .infinity)
            }
            .scrollDismissesKeyboard(.immediately)
            .scrollIndicators(.hidden)
            .scrollContentBackground(.hidden)
        }
    }

    @ToolbarContentBuilder
    private var toolBar: some ToolbarContent {
        ToolbarItem(placement: .principal) {
            modelSelectedButton
        }
    }

    func loadModelFile() {
        fileImporter = .init(contentType: [.item], onCompletion: { result in
            if let url = try? result.get() {
                vm.modelName = url.lastPathComponent
                do {
                    _ = url.startAccessingSecurityScopedResource()
                    let modelPath = try FileManager.copyToDocumentsDirectory(from: url).path()
                    url.stopAccessingSecurityScopedResource()
                    Task {
                        await vm.loadModel(from: modelPath)
                    }
                } catch {
                    print(error)
                }
            }
        })
    }

    private var modelSelectedButton: some View {
        HStack(alignment: .center, spacing: 4) {
            let title = vm.modelName.isEmpty ? "Selected Model" : vm.modelName
            Text(title)
                .textStyle(.subtitle2())
                .lineLimit(1)
        }
        .padding(.horizontal, 4)
        .contentShape(Rectangle())
        .padding(.vertical, 2)
        .padding(.horizontal, 8)
        .anyButton {
            loadModelFile()
        }
    }

    private var inputView: some View {
        HStack(alignment: .bottom) {
            promptTextField
                .padding(.leading, 8)
            Spacer()
            Button(action: onSendButtonPressed) {
                Image(systemName: (vm.isGenerating || vm.prompt.isEmpty) ? "stop.circle.fill" : "arrow.up.circle.fill")
                    .fontSize(25)
                    .foregroundStyle((vm.isGenerating || vm.prompt.isEmpty) ? Color.gray.opacity(0.6) : .primary)
                    .frame(width: 46, height: 46)
            }
            .disabled(vm.isGenerating || vm.prompt.isEmpty)
            .padding(.trailing, 8)
        }
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(.white)
                .stroke(Color.init(uiColor: .lightGray), lineWidth: 0.5)
        )
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 8)
                .fill(.white)
        )
    }

    private var promptTextField: some View {
        TextField(
            textFieldPlaceholder,
            text: $vm.prompt,
            prompt: Text(textFieldPlaceholder).textStyle(.body1(textColor: Color.Text.inactive)),
            axis: .vertical
        )
        .textStyle(.body1(textColor: Color.Text.primary))
        .padding(.vertical, 12)
        .padding(.leading, 4)
        .lineLimit(5)
        .disabled(!vm.modelManager.isLoaded)
        .focused($isFocused)
        .frame(minHeight: 46)
        .contentShape(Rectangle())
    }

    private var textFieldPlaceholder: String {
        if vm.isLoadingModel {
            return "Loading model, please wait..."
        }
        return vm.modelManager.isLoaded ? "Type prompt..." : "Model not loaded, Please initialize the model"
    }

    private func onSendButtonPressed() {
        isFocused = false
        vm.generationStream()
    }
}
