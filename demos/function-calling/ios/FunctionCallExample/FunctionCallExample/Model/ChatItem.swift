import Foundation

enum ChatItem: Equatable, Identifiable {
    struct Content: Identifiable, Equatable {
        let `id`: String
        let content: String

        init(content: String, id: String = UUID().uuidString) {
            self.content = content
            self.id = id
        }
    }

    case user(Message)
    case assistant(Message)
    case system(Content)
    case waiting(Content)
    case calendar(CalendarEventModel)

    var id: String {
        switch self {
        case .user(let message):
            return message.id
        case .assistant(let message):
            return message.id
        case .system(let content):
            return content.id
        case .waiting(let content):
            return content.id
        case .calendar(let event):
            return event.id
        }
    }
}
