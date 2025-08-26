import SwiftUI

struct ChatAssistantBubbleView: View {

    let message: Message
    var isResponding: Bool = false

    var body: some View {
        VStack(spacing: 12) {
            if isResponding {
                ActivityIndicatorView()
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.leading, 12)
            } else {
                let rawStr = message.content.trimmingCharacters(in: .whitespacesAndNewlines)
                Text(rawStr)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.horizontal, 12)
            }
        }
    }
}

#Preview {
    ScrollView {
        ChatAssistantBubbleView(message: .assistant("thiLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et doloreLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et doloreis a think message this is a think this is a think this is a think this is a think</think> orem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor "))
    }
}
