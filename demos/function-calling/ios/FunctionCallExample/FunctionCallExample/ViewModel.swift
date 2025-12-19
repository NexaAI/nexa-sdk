import PhotosUI
import SwiftUI
import Foundation
import UniformTypeIdentifiers

@Observable
@MainActor
class ViewModel {

    var query: String = ""
    var ipAddress: String = "http://8.137.53.212:10052"
    let modelName: String = "Function Call Demo"

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

//            Task {
//                try? await Task.sleep(for: .seconds(3.5))
//                currentGenerateItem = .waiting(.init(content: "Interacting with MCP..."))
//            }

            let (eventModel, responseText) = try await sendWithRetry(text: prompt, imageUrl: imageUrl)

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

    func sendWithRetry(
        text: String,
        imageUrl: URL?,
        maxRetries: Int = 3
    ) async throws -> (event: CalendarEventModel?, responseText: String) {

        var lastResponseText = ""
        for attempt in 1...maxRetries {
            let (event, responseText) = try await sendFunctionCallRequest(
                text: text,
                imageUrl: imageUrl
            )

            lastResponseText = responseText

            if event != nil {
                return (event, responseText)
            }
            if attempt < maxRetries {
                try Task.checkCancellation()
                try await Task.sleep(for: .milliseconds(300))
            }
        }
        return (nil, lastResponseText)
    }

    func sendFunctionCallRequest(text: String, imageUrl: URL?) async throws -> (event: CalendarEventModel?, responseText: String) {
        guard let url = URL(string: "\(ipAddress)/api/calendar/create") else {
            throw NSError(domain: "Bad url", code: NSURLErrorBadURL)
        }

        var image: String = ""
        if let imageUrl = imageUrl,
           let imageData = try? Data(contentsOf: imageUrl) {
            let base64String = imageData.base64EncodedString()
            image = base64String
        }

        let body: [String: String] = [
            "query": text,
            "base64": image
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

        print("result: ", String(data: data, encoding: .utf8) ?? "")
        let functionCallResponse = try JSONDecoder().decode(CreateCalendarResponse.self, from: data)
        guard functionCallResponse.success, let calendar = functionCallResponse.data else {
            return (nil, functionCallResponse.message)
        }

        var eventModel = CalendarEventModel()
        eventModel.id = calendar.calendarId ?? ""
        eventModel.eventName = calendar.summary ?? ""
        eventModel.description = calendar.description ?? ""
        eventModel.location = calendar.location ?? ""

        let dateFormatter = DateFormatter()
        dateFormatter.locale = Locale(identifier: "en_US")
        dateFormatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ss"

        if let startDate = calendar.start,
           !startDate.isEmpty,
           let date = dateFormatter.date(from: startDate) {

            if let timeZoneID = calendar.timeZone {
                let timeZone = TimeZone(identifier: timeZoneID) ?? .current
                dateFormatter.timeZone = timeZone
            }

            dateFormatter.dateFormat = "EEEE, MMMM d"
            eventModel.startDate = dateFormatter.string(from: date)

            dateFormatter.dateFormat = "h:mm a"
            eventModel.startDateTime = dateFormatter.string(from: date)
        }

        dateFormatter.dateFormat = "yyyy-MM-dd'T'HH:mm:ss"
        if let endDate = calendar.end,
           !endDate.isEmpty,
           let date = dateFormatter.date(from: endDate) {
            if let timeZoneID = calendar.timeZone {
                let timeZone = TimeZone(identifier: timeZoneID) ?? .current
                dateFormatter.timeZone = timeZone
            }
            dateFormatter.dateFormat = "h:mm a"
            eventModel.endDateTime = dateFormatter.string(from: date)
        }

        return (eventModel, functionCallResponse.message)
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


struct CreateCalendarResponse: Codable {
    let success: Bool
    let data: CalendarEventData?
    let message: String
}

struct CalendarEventData: Codable {
    let htmlLink: String?
    let calendarId: String?
    let summary: String?
    let start: String?
    let end: String?
    let timeZone: String?
    let location: String?
    let description: String?
}
