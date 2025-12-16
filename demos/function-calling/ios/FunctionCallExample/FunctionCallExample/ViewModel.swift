import PhotosUI
import SwiftUI
import Foundation
import UniformTypeIdentifiers

@Observable
@MainActor
class ViewModel {

    var query: String = ""
    var ipAddress: String = "192.168.1.107"
    let modelName: String = "OmniNeural"
    
    private(set) var messages: [ChatItem] = .init()
    private(set) var currentGenerateItem: ChatItem?

    private(set) var isLoading: Bool = false
    private(set) var isGenerating: Bool = false

    var error: String?

    private(set) var selectedImages: [URL] = []

    func clear() {
        messages = []
        currentGenerateItem = nil
        selectedImages = []
    }

    func send() async {
        isGenerating = true
        defer {
            isGenerating = false
            currentGenerateItem = nil
        }
        let prompt = query.trimmingCharacters(in: .whitespacesAndNewlines)
        let message = Message.user(prompt, images: selectedImages)
        messages.append(.user(message))

        query = ""
        selectedImages = []

        currentGenerateItem = .waiting(.init(content: "Analyzing..."))

        try? await Task.sleep(for: .seconds(2))

        messages.append(.assistant(.assistant("The event has been successfully added to your calendar.")))
        messages.append(.calendar(.mock))
    }

    func removeImage(url: URL) {
        selectedImages.removeAll { $0.absoluteString == url.absoluteString }
    }

    func handlePhotosImagePickerResult(_ photosPickerItems: [PhotosPickerItem]) {
        guard !photosPickerItems.isEmpty else {
            return
        }
        Task {
            for item in photosPickerItems {
                if let data = try? await item.loadTransferable(type: Data.self),
                   let url = try? saveToCaches(data, name: "\(item.itemIdentifier ?? UUID().uuidString).jpg") {
                    selectedImages.append(url)
                }
            }
        }
    }
}

private func saveToCaches(_ data: Data, name: String) throws -> URL {
    let fileManager = FileManager.default
    let downloadFolder = fileManager.urls(for: .cachesDirectory, in: .userDomainMask).first!
    let cacheFolder = downloadFolder.appendingPathComponent("data")
    if !fileManager.fileExists(atPath: cacheFolder.path) {
        try fileManager.createDirectory(at: cacheFolder, withIntermediateDirectories: true)
    }
    let fileURL = cacheFolder.appendingPathComponent(name)
    try data.write(to: fileURL)
    return fileURL
}

enum VMError: LocalizedError {
    case modelNotLoad

    var errorDescription: String? { description }
    var description: String {
        switch self {
        case .modelNotLoad:
            return "Model not loaded, please load model first"
        }
    }
}


extension URL: @retroactive Identifiable {
    public var id: String { absoluteString }
}
