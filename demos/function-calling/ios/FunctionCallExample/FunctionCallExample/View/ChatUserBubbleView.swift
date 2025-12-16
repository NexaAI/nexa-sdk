import SwiftUI

struct ChatUserBubbleView: View {
    let message: Message
    var body: some View {
        VStack(alignment: .trailing, spacing: 16) {
            ScrollViewReader { proxy in
                ScrollView(.horizontal) {
                    LazyHStack (spacing: 4) {
                        ForEach(message.images, id: \.self) { url in
                            buildImageViewItem(url)
                                .id(url.absoluteString)
                        }
                    }
                }
                .scrollIndicators(.hidden)
                .scrollDisabled(message.images.count == 1)
                .defaultScrollAnchor(.trailing)
            }

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

    @ViewBuilder
    private func buildImageViewItem(_ url: URL) -> some View {
        if let image = ImageCache.loadImage(at: url) {
            GeometryReader {
                let size = $0.size
                Image(uiImage: image)
                    .resizable()
                    .aspectRatio(contentMode: .fill)
                    .frame(width: size.width, height: size.height)
                    .clipShape(
                        RoundedRectangle(cornerRadius: 16)
                    )
                    .overlay(
                        RoundedRectangle(cornerRadius: 16)
                            .stroke(Color.ImageBox.border, lineWidth: 1)
                    )
            }
            .frame(width: imageItemSize.width, height: imageItemSize.height)
        }
    }

    private var imageItemSize: CGSize {
        let width: CGFloat
        let height: CGFloat
        if message.images.count == 1 {
            if let url = message.images.first,
               let image = ImageCache.loadImage(at: url) {
                let iWidth = image.size.width
                let iHeight = image.size.height
                (width, height) = iHeight > iWidth ? (160.0, 240.0) : (240.0, 160.0)
            } else {
                (width, height) = (100, 100)
            }
        } else {
            (width, height) = (100, 100)
        }
        return .init(width: width, height: height)
    }
}

#Preview {
    ChatUserBubbleView(message: .user("hello world"))
}
