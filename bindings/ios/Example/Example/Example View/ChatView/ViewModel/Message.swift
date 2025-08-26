
import Foundation

struct Message: Identifiable, Equatable {
    var id: String
    var role: Role
    var content: String
    var images: [URL]
    var audios: [URL]
    var videos: [URL]
    var createAt: Date

    init(id: String = UUID().uuidString, role: Role, content: String, images: [URL] = [], audios: [URL] = [],  videos: [URL] = [], createAt: Date = .now) {
        self.id = id
        self.role = role
        self.content = content
        self.images = images
        self.audios = audios
        self.videos = videos
        self.createAt = createAt
    }

    enum Role: String, Codable {
        case user
        case assistant
        case system
    }

    var isUser: Bool {
        role == .user
    }
    var isAssistant: Bool {
        role == .assistant
    }

    var partation: (think: String, other: String) {
        if !isAssistant {
            return ("", content)
        }
        let thinkPrefix = "<think>"
        let thinkSuffix = "</think>"

        if content.hasPrefix(thinkPrefix) {
            var c = content
            c.removeFirst(thinkPrefix.count)
            var parts = c.split(separator: thinkSuffix)
            if !parts.isEmpty {
                let think = String(parts.removeFirst())
                let other = String(parts.joined())
                return (think, other)
            }
        }
        return ("", content)
    }
}

extension Message {
    static func user(_ content: String, images: [URL] = [], audios: [URL] = [], videos: [URL] = []) -> Message {
        Message(role: .user, content: content, images: images, audios: audios, videos: videos)
    }
    static func assistant(_ content: String) -> Message {
        Message(role: .assistant, content: content)
    }
    
    static func system(_ content: String) -> Message {
        Message(role: .system, content: content)
    }
}
