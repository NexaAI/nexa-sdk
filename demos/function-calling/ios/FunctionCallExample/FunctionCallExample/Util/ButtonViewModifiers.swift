
import SwiftUI

struct BackgroundButtonStyle: ButtonStyle {

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .background(
                RoundedRectangle(cornerRadius: 12)
                    .fill(Color.Menu.Bg.active)
                    .opacity(configuration.isPressed ? 1 : 0)
            )
    }
}



struct DefaultButtonStyle: ButtonStyle {
    @Environment(\.isEnabled) var isEnabled

    var foregroundStyle: Color = Color.Button.Secondary.Text.default
    var foregroundStyleDisabled: Color = Color.Button.Secondary.Text.disabled
    var enableBackground: Color = .clear
    var disableBackground: Color = .clear
    var cornerRadius: CGFloat = 18
    var strokeColor: Color = Color.Button.Secondary.Border.default
    var strokeColorDisable: Color = Color.Button.Secondary.Border.disabled
    var lineWidth: CGFloat = 1
    var padding: EdgeInsets = EdgeInsets(top: 6, leading: 12, bottom: 6, trailing: 12)

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .padding(padding)
            .foregroundStyle(isEnabled ? (configuration.isPressed ? foregroundStyle.opacity(0.6) : foregroundStyle) : foregroundStyleDisabled)
            .background(
                RoundedRectangle(cornerRadius: cornerRadius)
                    .fill(isEnabled ? enableBackground : disableBackground)
                    .stroke(isEnabled ? (configuration.isPressed ? strokeColor.opacity(0.6) : strokeColor) : strokeColorDisable , lineWidth: lineWidth)
            )
            .animation(.easeInOut(duration: 0.2), value: configuration.isPressed)
    }
}

struct HighlightButtonStyle: ButtonStyle {

    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .overlay {
                configuration.isPressed ? Color.accentColor.opacity(0.4) : Color.accentColor.opacity(0)
            }
            .animation(.smooth, value: configuration.isPressed)
    }
}

struct PressableButtonStyle: ButtonStyle {
    
    func makeBody(configuration: Configuration) -> some View {
        configuration.label
            .scaleEffect(configuration.isPressed ? 0.95 : 1)
            .animation(.smooth, value: configuration.isPressed)
    }
}

enum ButtonStyleOption {
    case press, highlight, plain, background
}

extension View {

    @ViewBuilder
    func defaultButton(
        _ enableBackground: Color = .clear,
        disableBackground: Color = .clear,
        foregroundStyle: Color = Color.Button.Secondary.Text.default,
        foregroundStyleDisabled: Color = Color.Button.Secondary.Text.disabled,
        cornerRadius: CGFloat = 18,
        strokeColor: Color = Color.Button.Secondary.Border.default,
        strokeColorDisable: Color = Color.Button.Secondary.Border.disabled,
        lineWidth: CGFloat = 1
    ) -> some View {
        self.buttonStyle(
            DefaultButtonStyle(
                foregroundStyle: foregroundStyle,
                foregroundStyleDisabled: foregroundStyleDisabled,
                enableBackground: enableBackground,
                disableBackground: disableBackground,
                cornerRadius: cornerRadius,
                strokeColor: strokeColor,
                strokeColorDisable: strokeColorDisable,
                lineWidth: lineWidth
            )
        )
    }

    @ViewBuilder
    func anyButton(_ option: ButtonStyleOption = .plain, action: @escaping () -> Void) -> some View {
        switch option {
        case .press:
            self.pressableButton(action: action)
        case .highlight:
            self.highlightButton(action: action)
        case .plain:
            self.plainButton(action: action)
        case .background:
            self.backgroundButton(action: action)
        }
    }

    private func backgroundButton(action: @escaping () -> Void) -> some View {
        Button {
            action()
        } label: {
            self
        }
        .buttonStyle(BackgroundButtonStyle())
    }

    private func plainButton(action: @escaping () -> Void) -> some View {
        Button {
            action()
        } label: {
            self
        }
        .buttonStyle(.borderless)
    }
    
    private func highlightButton(action: @escaping () -> Void) -> some View {
        Button {
            action()
        } label: {
            self
        }
        .buttonStyle(HighlightButtonStyle())
    }
    
    private func pressableButton(action: @escaping () -> Void) -> some View {
        Button {
            action()
        } label: {
            self
        }
        .buttonStyle(PressableButtonStyle())
    }
}

#Preview {
    VStack {
        Text("Hello, world!")
            .padding()
            .frame(maxWidth: .infinity)
            .anyButton(.highlight, action: {
                
            })
        .padding()
        
        Text("Hello, world!")
            .anyButton(.press, action: {
                
            })
            .padding()
        
        Text("Hello, world!")
            .anyButton(action: {
                
            })
            .padding()

        HStack {
            Button("DefaultButton") {

            }
            .buttonStyle(DefaultButtonStyle())

            Button("Disable") {

            }
            .buttonStyle(DefaultButtonStyle())
            .disabled(true)
        }

    }
}
