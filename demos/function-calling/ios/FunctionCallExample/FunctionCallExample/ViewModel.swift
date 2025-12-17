import PhotosUI
import SwiftUI
import Foundation
import UniformTypeIdentifiers

@Observable
@MainActor
class ViewModel {

    var query: String = ""
    var ipAddress: String = "https://falciform-seminormally-dorine.ngrok-free.dev"
    let modelName: String = "OmniNeural"
    
    private(set) var messages: [ChatItem] = .init()
    private(set) var currentGenerateItem: ChatItem?

    private(set) var isGenerating: Bool = false

    var error: String?

    var selectedItems: [PhotosPickerItem] = []
    private(set) var selectedImages: [URL] = []

    func clear() {
        messages = []
        currentGenerateItem = nil
        selectedImages = []
        selectedItems = []
    }

    func send() async {
        isGenerating = true
        defer {
            isGenerating = false
            currentGenerateItem = nil
        }
        do {
            let prompt = query.trimmingCharacters(in: .whitespacesAndNewlines)
            let message = Message.user(prompt, images: selectedImages)
            messages.append(.user(message))

            let imageUrl = selectedImages.first
            query = ""
            selectedItems = []
            selectedImages = []

            currentGenerateItem = .waiting(.init(content: "Analyzing..."))

            Task {
                try? await Task.sleep(for: .seconds(3.5))
                currentGenerateItem = .waiting(.init(content: "Interacting with MCP..."))
            }

            let (eventModel, responseText) = try await sendFunctionCallRequest(text: prompt, imageUrl: imageUrl)

            if let eventModel {
                messages.append(.assistant(.assistant("The event has been successfully added to your calendar.")))
                messages.append(.calendar(eventModel))
            }

            if !responseText.isEmpty {
                messages.append(.assistant(.assistant(responseText)))
            }

        } catch {
            self.error = error.localizedDescription
        }
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

    func sendFunctionCallRequest(text: String, imageUrl: URL?) async throws -> (event: CalendarEventModel?, responseText: String) {
        guard let url = URL(string: "\(ipAddress)/api/function-call") else {
            throw NSError(domain: "Bad url", code: NSURLErrorBadURL)
        }

        var image: String = ""
        if let imageUrl = imageUrl,
            let imageData = try? Data(contentsOf: imageUrl) {
            let base64String = imageData.base64EncodedString()
            image = "data:image/png;base64,\(base64String)"
        }

        let body: [String: String] = [
            "text": text,
            "image": image
        ]

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONSerialization.data(withJSONObject: body)
        request.timeoutInterval = 120

        let (data, response) = try await URLSession.shared.data(for: request)
        guard let httpResponse = response as? HTTPURLResponse else {
            throw VMError.invalidResponse
        }
        guard (200...299).contains(httpResponse.statusCode) else {
            throw VMError.statusCode(httpResponse.statusCode, data)
        }

        let functionCallResponse = try JSONDecoder().decode(FunctionCallResponse.self, from: data)
        let responseText = functionCallResponse.response_text ?? ""
        
        guard let innerData = functionCallResponse.func_result?.data(using: .utf8) else {
            throw VMError.invalidResponse
        }

        let funcResult = try JSONDecoder().decode(FuncResultContent.self, from: innerData)
        if let textItem = funcResult.content?.first?.text {
            let eventData = Data(textItem.utf8)
            let eventWrapper = try JSONDecoder().decode(EventWrapper.self, from: eventData)
            let event = eventWrapper.event


            var eventModel = CalendarEventModel()
            eventModel.id = event.id
            eventModel.eventName = event.summary ?? ""
            eventModel.description = event.description ?? ""
            eventModel.location = event.location ?? ""

            let dateFormatter = DateFormatter()
            dateFormatter.locale = Locale(identifier: "en_US")
            dateFormatter.dateFormat = "EEEE, MMMM d"

            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withInternetDateTime]
            if let startDate = event.start?.dateTime,
                !startDate.isEmpty,
                let date = formatter.date(from: startDate) {

                if let timeZoneID = event.start?.timeZone {
                    let timeZone = TimeZone(identifier: timeZoneID) ?? .current
                    dateFormatter.timeZone = timeZone
                }

                eventModel.startDate = dateFormatter.string(from: date)

                dateFormatter.dateFormat = "h:mm a"
                eventModel.startDateTime = dateFormatter.string(from: date)
            }

            if let endDate = event.end?.dateTime,
                !endDate.isEmpty,
                let date = formatter.date(from: endDate) {

                if let timeZoneID = event.end?.timeZone {
                    let timeZone = TimeZone(identifier: timeZoneID) ?? .current
                    dateFormatter.timeZone = timeZone
                }
                dateFormatter.dateFormat = "h:mm a"
                eventModel.endDateTime = dateFormatter.string(from: date)
            }

            return (eventModel, responseText)
        }

        return (nil, responseText)
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
    case invalidResponse
    case statusCode(Int, Data)
}


extension URL: @retroactive Identifiable {
    public var id: String { absoluteString }
}


struct FunctionCallResponse: Codable {
    var func_name: String?
    var func_result: String?
    var response_text: String?
}

struct FuncResultContent: Codable {
    let meta: String?
    let content: [ContentItem]?
    let structuredContent: String?
    let isError: Bool?
}

struct ContentItem: Codable {
    let type: String?
    let text: String?
}

struct EventWrapper: Codable {
    let event: Event
}

struct Event: Codable {
    let id: String
    let summary: String?
    let start: EventDate?
    let end: EventDate?
    let status: String?
    let htmlLink: String?
    let description: String?
    let location: String?
}

struct EventDate: Codable {
    let dateTime: String?
    let timeZone: String?
}
