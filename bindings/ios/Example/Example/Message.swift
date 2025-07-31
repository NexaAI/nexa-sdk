
import Foundation

struct Message: Identifiable, Equatable {
    /// Unique identifier for the message
    var id: String

    /// The role of the message sender (user, assistant, or system)
    var role: Role

    /// The text content of the message
    var content: String

    /// Array of image URLs attached to the message
    var images: [URL]

    /// Array of audios URLs attached to the message
    var audios: [URL]

    /// Array of video URLs attached to the message
    var videos: [URL]

    /// Timestamp when the message was created
    var createAt: Date

    /// Creates a new message with the specified role, content, and optional media attachments
    /// - Parameters:
    ///   - role: The role of the message sender
    ///   - content: The text content of the message
    ///   - images: Optional array of image URLs
    ///   - videos: Optional array of video URLs
    init(id: String = UUID().uuidString, role: Role, content: String, images: [URL] = [], audios: [URL] = [],  videos: [URL] = [], createAt: Date = .now) {
        self.id = id
        self.role = role
        self.content = content
        self.images = images
        self.audios = audios
        self.videos = videos
        self.createAt = createAt
    }

    /// Defines the role of the message sender in the conversation
    enum Role: String, Codable {
        /// Message from the user
        case user
        /// Message from the AI assistant
        case assistant
        /// System message providing context or instructions
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

/// Convenience methods for creating different types of messages
extension Message {
    /// Creates a user message with optional media attachments
    /// - Parameters:
    ///   - content: The text content of the message
    ///   - images: Optional array of image URLs
    ///   - videos: Optional array of video URLs
    /// - Returns: A new Message instance with user role
    static func user(_ content: String, images: [URL] = [], audios: [URL] = [], videos: [URL] = []) -> Message {
        Message(role: .user, content: content, images: images, audios: audios, videos: videos)
    }

    /// Creates an assistant message
    /// - Parameter content: The text content of the message
    /// - Returns: A new Message instance with assistant role
    static func assistant(_ content: String) -> Message {
        Message(role: .assistant, content: content)
    }

    /// Creates a system message
    /// - Parameter content: The text content of the message
    /// - Returns: A new Message instance with system role
    static func system(_ content: String) -> Message {
        Message(role: .system, content: content)
    }
}
