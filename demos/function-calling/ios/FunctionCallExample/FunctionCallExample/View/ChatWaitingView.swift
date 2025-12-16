
import SwiftUI

struct ChatWaitingView: View {
    var text: String
    var body: some View {
        HStack(spacing: 8) {
            ZStack {
                Circle()
                    .fill(Color.Thinkingbox.dotBack)
                    .frame(width: 14)
                Circle()
                    .fill(Color.Thinkingbox.dotFront)
                    .frame(width: 8)
            }
            Text(text)
                .fontSize(16)
                .fontWeight(.medium)
                .foregroundStyle(Color.Thinkingbox.font)
        }
        .frame(maxWidth: .infinity, alignment: .leading)
        .padding(.leading, 4)
        .padding(.trailing, 12)
    }
}

#Preview {
    ChatWaitingView(text: "Analyzing...")
}
