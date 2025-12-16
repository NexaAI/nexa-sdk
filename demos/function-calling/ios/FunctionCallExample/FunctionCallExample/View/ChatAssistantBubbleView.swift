import SwiftUI

struct ChatAssistantBubbleView: View {
    let message: Message
    var body: some View {
        VStack(spacing: 12) {
            let rawStr = message.content.trimmingCharacters(in: .whitespacesAndNewlines)
            MarkdownView(content: rawStr)
                .frame(maxWidth: .infinity, alignment: .leading)
                .padding(.leading, 4)
                .padding(.trailing, 12)
        }
    }
}

#Preview {
    ScrollView {
        ChatAssistantBubbleView(message: .assistant("thiLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et doloreLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et doloreis a think message this is a think this is a think this is a think this is a think</think> orem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor "))
    }
}
