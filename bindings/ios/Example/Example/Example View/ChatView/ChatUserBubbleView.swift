import SwiftUI

struct ChatUserBubbleView: View {
    let message: Message
    var body: some View {
        VStack(alignment: .trailing, spacing: 16) {
            if !message.content.isEmpty {
                Text(message.content)
                    .textStyle(.body1(textColor: Color.Chatbox.font))
                    .multilineTextAlignment(.leading)
                    .padding(12)
                    .background(
                        UnevenRoundedRectangle(topLeadingRadius: 16, bottomLeadingRadius: 16, bottomTrailingRadius: 4, topTrailingRadius: 16, style: .continuous)
                            .fill(Color.Chatbox.bg)
                    )
                    .padding(.leading, 32)
                    .frame(maxWidth: .infinity, alignment: .trailing)
            }
        }
    }
}

#Preview {
    ChatUserBubbleView(message: .user("hello world"))
}
