import SwiftUI

struct TextStyleModifier: ViewModifier {

    struct TextStyle {
        let fontSize: CGFloat
        let fontWeight: Font.Weight
        let textColor: Color
        let kerning: CGFloat

        static func largeTitle(
            fontSize: CGFloat = 34,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func title1(
            fontSize: CGFloat = 28,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func title2(
            fontSize: CGFloat = 24,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func title3(
            fontSize: CGFloat = 20,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func headline(
            fontSize: CGFloat = 16,
            fontWeight: Font.Weight = .medium,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.15
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func subtitle1(
            fontSize: CGFloat = 16,
            fontWeight: Font.Weight = .medium,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.15
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func subtitle2(
            fontSize: CGFloat = 14,
            fontWeight: Font.Weight = .medium,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.1
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func body1(
            fontSize: CGFloat = 16,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.15
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func body2(
            fontSize: CGFloat = 14,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.15
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func caption1(
            fontSize: CGFloat = 12,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.25
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }

        static func caption2(
            fontSize: CGFloat = 10,
            fontWeight: Font.Weight = .regular,
            textColor: Color = Color.Text.primary,
            kerning: CGFloat = 0.5
        ) -> TextStyle {
            TextStyle(fontSize: fontSize, fontWeight: fontWeight, textColor: textColor, kerning: kerning)
        }
    }

    let style: TextStyle

    func body(content: Content) -> some View {
        content
            .fontSize(style.fontSize)
            .fontWeight(style.fontWeight)
            .foregroundStyle(style.textColor)
            .kerning(style.kerning)

    }
}

extension Text {

    func textStyle(_ style: TextStyleModifier.TextStyle) -> Text {
        self
            .fontSize(style.fontSize)
            .fontWeight(style.fontWeight)
            .foregroundStyle(style.textColor)
            .kerning(style.kerning)
    }
    
    func fontSize(_ size: CGFloat) -> Text {
        font(.system(size: size))
    }
}

extension View {

    func fontSize(_ size: CGFloat) -> some View {
        font(.system(size: size))
    }

    func textStyle(_ style: TextStyleModifier.TextStyle) -> some View {
        modifier(TextStyleModifier(style: style))
    }
}
