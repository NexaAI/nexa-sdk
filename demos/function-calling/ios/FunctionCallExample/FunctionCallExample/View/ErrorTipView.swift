import SwiftUI

struct ErrorTipView: View {
    let error: String
    var body: some View {
        errorTipView(error: error)
    }

    @ViewBuilder
    func errorTipView(error: String) -> some View {
        VStack {
            Text(error)
                .multilineTextAlignment(.center)
                .lineLimit(3)
                .fontSize(14)

        }
        .padding(.horizontal, 12)
        .padding(.vertical, 12)
        .frame(maxWidth: .infinity)
        .background(Color.accent)
        .cornerRadius(8)
        .padding(.horizontal, 12)
        .transition(.move(edge: .bottom).combined(with: .opacity))
        .offset(y: -100)
    }
}


