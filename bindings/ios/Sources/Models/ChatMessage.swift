import Foundation

public struct ChatMessage {
    public var role: Role
    public var content: String
    public init(role: Role, content: String) {
        self.role = role
        self.content = content
    }
}

public enum Role: String {
    case user
    case assistant
    case system
}
