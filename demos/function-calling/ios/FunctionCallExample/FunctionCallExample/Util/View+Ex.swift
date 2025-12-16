
import SwiftUI

extension View {

    func borderStyle(
        textColor: Color = Color.Text.primary,
        fillColor: Color = .clear,
        strokeColor: Color = .black0
    ) -> some View {
        self
            .tint(textColor)
            .background(
                RoundedRectangle(cornerRadius: 12)
                    .fill(fillColor)
                    .stroke(strokeColor, lineWidth: 1)
            )
    }
}
