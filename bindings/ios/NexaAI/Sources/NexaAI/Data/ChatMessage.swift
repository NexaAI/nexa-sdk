import Foundation

public struct ChatMessage {
    public var role: Role
    public var content: String
    public var images: [String]
    public var audios: [String]

    public init(
        role: Role,
        content: String,
        images: [String] = [],
        audios: [String] = []
    ) {
        self.role = role
        self.content = content
        self.images = images
        self.audios = audios
    }
}

public enum Role: String {
    case user
    case assistant
    case system
}
