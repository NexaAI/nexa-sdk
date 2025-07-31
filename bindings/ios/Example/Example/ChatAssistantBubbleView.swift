import SwiftUI

struct ChatAssistantBubbleView: View {

    let message: Message
    var isResponding: Bool = false

    @State var isExpand: Bool = true

    var body: some View {
        VStack(spacing: 12) {
            thinkView

            if isResponding {
                ActivityIndicatorView()
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding(.leading, 12)
            } else {
                if !message.partation.other.isEmpty {
                    let rawStr = message.partation.other.trimmingCharacters(in: .whitespacesAndNewlines)
                    Text(rawStr)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .padding(.horizontal, 12)
                }
            }
        }
    }

    private var thinkView: some View {
        VStack(spacing: 12) {
            HStack {
                Circle()
                    .fill(Color.Thinkingbox.dotFront)
                    .frame(width: 14, height: 14)
                    .padding(3.5)
                    .background(
                        Circle()
                            .fill(Color.Thinkingbox.dotBack)
                    )
                Text(isResponding ? "Thinking ..." : "Complete")
                    .textStyle(.subtitle1(textColor: Color.Thinkingbox.font))
                Spacer()
                Image(.chevronDown)
                    .renderingMode(.template)
                    .scaledToFit()
                    .frame(width: 16, height: 16)
                    .rotationEffect(isExpand ? .degrees(0) : .degrees(-90))
                    .foregroundStyle(Color.Thinkingbox.icon)
            }
            .anyButton {
                isExpand.toggle()
            }
            if isExpand {
                let think = message.partation.think.trimmingCharacters(in: .whitespacesAndNewlines)
                Text(think)
                    .textStyle(.init(fontSize: 14, fontWeight: .medium, textColor:  Color.Thinkingbox.font, kerning: 0))
                    .frame(maxWidth: .infinity, alignment: .leading)

            }
        }
        .padding(16)
        .background(
            RoundedRectangle(cornerRadius: 12)
                .stroke(Color.Thinkingbox.border, lineWidth: 1)
        )
    }
}

#Preview {
    ScrollView {
        ChatAssistantBubbleView(message: .assistant("thiLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et doloreLorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et doloreis a think message this is a think this is a think this is a think this is a think</think> orem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor "), isExpand: false)
    }
}
