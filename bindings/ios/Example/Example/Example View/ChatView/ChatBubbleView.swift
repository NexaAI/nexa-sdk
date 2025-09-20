
import SwiftUI


struct ChatBubbleView: View {

    let message: Message
    var isResponding: Bool = false
    var body: some View {
        VStack(spacing: 0) {
            content
                .padding(.horizontal, 16)
        }
    }

    @ViewBuilder
    private var content: some View {
        switch message.role {
        case .user:
            ChatUserBubbleView(message: message)
        case .assistant:
            ChatAssistantBubbleView(message: message, isResponding: isResponding)
            .padding(.vertical, 8)
        case .system:
            EmptyView()
        }
    }
}

#Preview {
    ScrollView {
        VStack(spacing: 16) {
            ChatBubbleView(message: .user("This is a user message"))
        }
    }
}
