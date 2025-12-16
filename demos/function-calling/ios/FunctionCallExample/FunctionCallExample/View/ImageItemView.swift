import SwiftUI

struct ImageItemView: View {
    let url: URL
    var onDeleteButtonPressed: (() -> Void)?
    var body: some View {
        ZStack(alignment: .topTrailing) {
            Image(uiImage: UIImage(contentsOfFile: url.path()) ?? .init())
                .resizable()
                .aspectRatio(contentMode: .fill)
                .frame(width: 100, height: 100)
                .clipShape(RoundedRectangle(cornerRadius: 16))
        }
        .background(
            RoundedRectangle(cornerRadius: 16)
                .fill(.clear)
                .stroke(Color.ImageBox.border, lineWidth: 1)
        )
        .overlay(alignment: .topTrailing) {
            Image(.x)
                .renderingMode(.template)
                .foregroundStyle(Color.Button.Default.Icon.default)
                .frame(width: 20, height: 20)
                .background(
                    Circle().fill(Color.Button.Default.Bg.default)
                )
                .offset(x: -4, y: 4)
                .anyButton {
                    withAnimation(.smooth) {
                        onDeleteButtonPressed?()
                    }
                }
        }
    }
}
