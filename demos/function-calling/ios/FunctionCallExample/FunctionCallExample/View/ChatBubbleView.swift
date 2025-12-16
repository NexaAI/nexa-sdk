
import SwiftUI

struct ChatBubbleView: View {
    let item: ChatItem
    var body: some View {
        VStack(spacing: 0) {
            content
                .padding(.horizontal, 16)
        }
    }

    @ViewBuilder
    private var content: some View {
        switch item {
        case .user(let message):
            ChatUserBubbleView(message: message)
        case .assistant(let message):
            ChatAssistantBubbleView(message: message)
        case .waiting(let item):
            ChatWaitingView(text: item.content)
        case .system(let item):
            SystemView(content: item.content)
        case .calendar(let event):
            CalendarEventView(event: event)
        }
    }
}

struct SystemView: View {
    let content: String
    var body: some View {
        Text("\(content)")
            .textStyle(.caption1(textColor: Color.Text.tertiary))
            .frame(maxWidth: .infinity, alignment: .leading)
            .padding(4)
    }
}

#Preview {
    ScrollView {
        VStack(spacing: 16) {
            ChatBubbleView(item: .user(.user("This is a user message")))
        }
    }
}
