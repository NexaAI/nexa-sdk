import SwiftUI
import PhotosUI

struct ContentView: View {

    @State var vm: ViewModel = .init()
    @FocusState private var isFocused: Bool

    @State private var showPhotoPicker: Bool = false
    @State private var selectedItems: [PhotosPickerItem] = []

    @State private var showSettingView: Bool = false

    var body: some View {
        GeometryReader { geo in
            mainView
        }
    }

    @ViewBuilder
    var mainView: some View {
        VStack(spacing: 0) {
            headerView
            messagesSection
        }
        .safeAreaInset(edge: .bottom) {
            inputView
                .background(
                    RoundedRectangle(cornerRadius: 24)
                        .fill(Color.Background.primary)
                        .strokeBorder(Color.gray2)
                        .ignoresSafeArea(edges: .bottom)
                )
                .padding(12)
                .background(Color.Background.secondary)
                .overlay {
                    if let error = vm.error, !error.isEmpty {
                        errorTipView(error: error)
                    }
                }
        }
        .photosPicker(
            isPresented: $showPhotoPicker,
            selection: $selectedItems,
            maxSelectionCount: 1,
            matching: .images
        )
        .animation(.smooth, value: vm.messages.count)
        .onChange(of: selectedItems) { oldValue, newValue in
            vm.handlePhotosImagePickerResult(newValue)
        }
        .sheet(isPresented: $showSettingView) {
            SettingView(ipAddress: $vm.ipAddress)
                .presentationDetents([.height(200)])
                .presentationDragIndicator(.hidden)
                .presentationCornerRadius(16)
        }
    }

    @ViewBuilder
    var headerView: some View {
        VStack {
            HStack {
                Image(.menu)
                    .renderingMode(.template)
                    .foregroundStyle(Color.Icon.primary)
                    .contentShape(.rect)
                    .anyButton {
                        showSettingView.toggle()
                    }

                Spacer()
                Text(vm.modelName)
                    .font(.system(size: 14, weight: .medium))
                Spacer()

                Image(.addChat)
                    .renderingMode(.template)
                    .foregroundStyle(Color.Icon.primary)
                    .contentShape(.rect)
                    .anyButton {
                        vm.clear()
                    }
            }
            .foregroundStyle(Color.Text.primary)
            .padding(12)

            Divider()
                .frame(height: 0.5)
                .background(Color.Background.primary)
        }
        .frame(maxWidth: .infinity)
        .background(Color.Background.secondary)
    }

    @ViewBuilder
    var addButton: some View {
        Button {
            showPhotoPicker = true
        } label: {
            Image(.addIcon)
        }
        .disabled(vm.isGenerating)
    }

    @ViewBuilder
    var sendButton: some View {
        Button {
            isFocused = false
            Task { await vm.send() }
        } label: {
            Image(.send)
        }
        .disabled(disableButton || vm.query.isEmpty)
    }

    @ViewBuilder
    var inputView: some View {
        VStack {
            if !vm.selectedImages.isEmpty {
                imageItemsView
            }
            promptTextField
            HStack {
                addButton
                Spacer()
                sendButton
            }
        }
        .padding()
    }

    @ViewBuilder
    var imageItemsView: some View {
        ScrollView(.horizontal) {
            HStack(spacing: 4) {
                ForEach(vm.selectedImages) { url in
                    ImageItemView(url: url) {
                        vm.removeImage(url: url)
                    }
                }
                Spacer()
            }
            .padding(.leading, 4)
        }
        .scrollIndicators(.hidden)
    }

    @ViewBuilder
    func errorTipView(error: String) -> some View {
        ErrorTipView(error: error)
        .onAppear {
            DispatchQueue.main.asyncAfter(deadline: .now() + 3) {
                withAnimation {
                    vm.error = nil
                }
            }
        }
    }

    private var disableButton: Bool {
        vm.isLoading || vm.isGenerating
    }

    @ViewBuilder
    var promptTextField: some View {
        TextField(
            "Enter prompt...",
            text: $vm.query,
            prompt: Text("Enter prompt...").textStyle(.body1(textColor: Color.Text.inactive)),
            axis: .vertical
        )
        .textStyle(.body1(textColor: Color.Text.primary))
        .padding(.vertical, 12)
        .padding(.leading, 4)
        .lineLimit(5)
        .frame(minHeight: 46)
        .contentShape(Rectangle())
        .focused($isFocused)
    }

    @ViewBuilder
    private var messagesSection: some View {
        GeometryReader { geo in
            ScrollViewReader { proxy in
                ScrollView {
                    Text("").frame(height: 16)

                    if vm.messages.isEmpty {
                        Text("Please load the model or download the model before use.")
                            .multilineTextAlignment(.center)
                            .fontSize(14)
                            .foregroundStyle(Color.Text.secondary)
                            .padding(.top, 100)
                            .padding(.horizontal, 12)
                            .frame(maxWidth: .infinity)
                    }

                    VStack(spacing: 16) {
                        ForEach(Array(vm.messages.enumerated()), id: \.offset) { (idx, message) in
                            ChatBubbleView(item: message)
                                .id(message.id)
                        }
                        if let currentGenerateItem = vm.currentGenerateItem {
                            ChatBubbleView(item: currentGenerateItem)
                                .id(currentGenerateItem.id)
                        }
                    }
                }
                .safeAreaPadding(.bottom, 20)
                .scrollDismissesKeyboard(.immediately)
                .defaultScrollAnchor(.top)
                .scrollContentBackground(.hidden)
                .background(Color.Background.secondary)
            }
        }
    }

}

#Preview {
    ContentView()
}

