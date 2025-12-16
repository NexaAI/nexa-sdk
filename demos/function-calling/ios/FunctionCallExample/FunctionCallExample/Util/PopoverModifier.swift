import SwiftUI

final class WindowOverlayManager {
    static let shared = WindowOverlayManager()

    private var window: UIWindow?

    func show<Content: View>(backgroundColor: UIColor = .clear, @ViewBuilder content: @escaping () -> Content) {
        guard let scene = UIApplication.shared.connectedScenes.first as? UIWindowScene else { return }

        if window == nil {
            let newWindow = UIWindow(windowScene: scene)
            newWindow.windowLevel = .alert + 1
            newWindow.backgroundColor = .clear
            newWindow.rootViewController = UIHostingController(rootView: AnyView(EmptyView()))
            newWindow.rootViewController?.view.backgroundColor = .clear
            newWindow.isHidden = false
            window = newWindow
        }

        window?.rootViewController?.view.backgroundColor = backgroundColor
        window?.backgroundColor = backgroundColor
        if let hosting = window?.rootViewController as? UIHostingController<AnyView> {
            hosting.rootView = AnyView(content())
        }

        window?.makeKeyAndVisible()
    }

    func hide() {
        window?.isHidden = true
        window = nil
    }
}

extension UIApplication {
    var globalSafeAreaInsets: UIEdgeInsets {
        let windowScene = connectedScenes.first as? UIWindowScene
        let window = windowScene?.windows.first
        return window?.safeAreaInsets ?? .zero
    }
}

struct PopoverModifier<ContentView: View>: ViewModifier {

    @Binding var isPresented: Bool

    let position: Position
    let contentSize: CGSize
    let backgroundColor: UIColor
    let contentView: () -> ContentView

    enum Position {
        case top
        case topLeading
        case topTrailing
        case bottom
        case bottomLeading
        case bottomTrailing
        case center
        case centerLeading
        case centerTrailing
        case auto
    }

    @State private var sourceFrame: CGRect = .zero
    @State private var lastPosition: CGPoint = .zero

    func body(content: Content) -> some View {
        content
            .onFrameChange { _, frame in
                sourceFrame = frame
                let contentFrame = CGRect(origin: .zero, size: contentSize)
                lastPosition = caculatePosition(contentFrame)
            }
            .onChange(of: isPresented) { old, newValue in
                if newValue {
                    showAnimated()
                } else {
                    hideAnimated()
                }
            }
    }

    private func showAnimated() {
        WindowOverlayManager.shared.show(backgroundColor: backgroundColor) {
            ZStack(alignment: .topLeading) {
                Color.gray.opacity(0.01)
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
                    .onTapGesture {
                        isPresented = false
                    }
                contentView()
                    .onFrameChange { _ , newFrame in
                        DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
                            self.lastPosition = self.caculatePosition(newFrame)
                        }
                    }
                    .position(lastPosition)
            }
            .ignoresSafeArea()
        }
    }

    private func hideAnimated() {
        WindowOverlayManager.shared.hide()
    }

    private func caculatePosition(_ contentFrame: CGRect) -> CGPoint {
        var result = CGPoint.zero
        switch position {
        case .center:
            result = .init(x: sourceFrame.midX, y: sourceFrame.midY)
        case .centerLeading:
            result = .init(x: sourceFrame.minX + contentFrame.size.width / 2, y: sourceFrame.midY)
        case .centerTrailing:
            result = .init(x: sourceFrame.maxX - contentFrame.size.width / 2, y: sourceFrame.midY)
        case .top:
            result = .init(x: sourceFrame.midX, y: sourceFrame.midY - sourceFrame.size.height / 2 - contentFrame.height / 2)
        case .topLeading:
            result = .init(x: sourceFrame.minX + contentFrame.size.width / 2, y: sourceFrame.midY - sourceFrame.size.height / 2 - contentFrame.height / 2)
        case .topTrailing:
            result = .init(x: sourceFrame.maxX - contentFrame.size.width / 2, y: sourceFrame.midY - sourceFrame.size.height / 2 - contentFrame.height / 2)
        case .bottom:
            result = .init(x: sourceFrame.midX, y: sourceFrame.midY + sourceFrame.size.height / 2 + contentFrame.height / 2)
        case .bottomLeading:
            result = .init(x: sourceFrame.minX + contentFrame.size.width / 2, y: sourceFrame.midY + sourceFrame.size.height / 2 + contentFrame.height / 2)
        case .bottomTrailing:
            result = .init(x: sourceFrame.maxX - contentFrame.size.width / 2, y: sourceFrame.midY + sourceFrame.size.height / 2 + contentFrame.height / 2)
        case .auto:

            let contentHeight = contentFrame.size.height
            let safeTop = UIApplication.shared.globalSafeAreaInsets.top + 44
            let safeBottom = UIApplication.shared.globalSafeAreaInsets.bottom + 44
            result = .init(x: sourceFrame.minX + contentFrame.size.width / 2, y: sourceFrame.midY)

            if sourceFrame.minY < safeTop && sourceFrame.maxY > (UIScreen.main.bounds.height - safeBottom) {
                result.y = safeTop + (UIScreen.main.bounds.height - safeTop - safeBottom) / 2 - contentHeight / 2
            } else {
                let minY = result.y - contentHeight / 2.0
                let maxY = result.y + contentHeight / 2.0

                if minY < safeTop {
                    var screenBounds = UIScreen.main.bounds
                    screenBounds.origin.y = safeTop
                    screenBounds.size.height -= (safeTop + safeBottom)
                    let visiableHeight = sourceFrame.intersection(screenBounds).height
                    result.y = safeTop + visiableHeight / 2 + contentHeight / 2
                } else if maxY > (UIScreen.main.bounds.height - safeBottom) {
                    if sourceFrame.minY > safeTop + (UIScreen.main.bounds.height - safeTop - safeBottom) / 2  {
                        if sourceFrame.minY + contentHeight > (UIScreen.main.bounds.height - safeBottom) {
                            result.y = sourceFrame.minY - contentHeight / 2
                        } else {
                            result.y = sourceFrame.minY + contentHeight / 2
                        }
                    } else {
                        result.y = sourceFrame.minY + (UIScreen.main.bounds.height - safeBottom - sourceFrame.minY) / 2 - contentHeight / 2
                    }
                }
            }
        }
        return result
    }
}

extension View {
    func anyPopover<Content: View>(
        isPresented: Binding<Bool>,
        position: PopoverModifier<Content>.Position = .center,
        contentSize: CGSize,
        backgroundColor: UIColor = .clear,
        @ViewBuilder content: @escaping () -> Content
    ) -> some View {
        self.modifier(PopoverModifier(isPresented: isPresented, position: position, contentSize: contentSize, backgroundColor: backgroundColor, contentView: content))
    }
}


#Preview {
    @Previewable @State var showOverlay = false
    @Previewable @State var buttonFrame: CGRect = .zero

    NavigationStack {
        VStack {

            Spacer()
            Button("show top") {
                showOverlay = true
            }
            .anyPopover(isPresented: $showOverlay, position: .top, contentSize: .init(width: 200, height: 500)) {
                VStack(spacing: 16) {
                    Text("🎉 top window")
                        .font(.title2)
                        .foregroundColor(.white)
                    Button("close") {
                        showOverlay = false
                    }
                    .padding(.top)
                }
                .background(.brown.opacity(0.5))
                .cornerRadius(12)
                .offset(x: 0, y: -8)
            }

        }
        .navigationTitle("Hello world")
        .navigationBarTitleDisplayMode(.inline)

    }
}
