
import SwiftUI
import MarkdownUI

struct MarkdownView: View {
    var content: String
    var scrollDisabled: Bool = true
    var fontSize: CGFloat = 16
    var fontColor: Color = Color.Text.primary
    var body: some View {
        ScrollView {
            Markdown(content)
        }
        .markdownTextStyle {
            MarkdownUI.ForegroundColor(fontColor)
            MarkdownUI.FontSize(fontSize)
        }
        .markdownTextStyle(\.strong) {
            MarkdownUI.FontSize(fontSize)
            MarkdownUI.ForegroundColor(fontColor)
        }
        .markdownBlockStyle(\.heading1){ configuration in
            configuration.label
                .markdownTextStyle {
                    FontWeight(.medium)
                }
        }
        .markdownBlockStyle(\.heading3) { configuration in
            configuration.label
                .markdownTextStyle {
                    FontWeight(.medium)
                    FontSize(20)
                }
        }
        .scrollDisabled(scrollDisabled)
    }
}
