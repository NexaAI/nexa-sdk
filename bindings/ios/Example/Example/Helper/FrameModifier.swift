import SwiftUI

struct FrameModifier: ViewModifier {
    let coordinateSpace: CoordinateSpace
    let onFrameChange: ((_ oldValue: CGRect, _ newValue: CGRect) -> Void)
    func body(content: Content) -> some View {
        content.background(
            GeometryReader { proxy in
                Color.clear
                    .onAppear(perform: {
                        onFrameChange(.zero, proxy.frame(in: coordinateSpace))
                    })
                    .onChange(of:  proxy.frame(in: coordinateSpace), { oldValue, newValue in
                        onFrameChange(oldValue, newValue)
                    })
            }
        )
    }
}


extension View {
    func onFrameChange(in coordinateSpace: CoordinateSpace = .global, perform: @escaping ((CGRect, CGRect) -> Void)) -> some View {
        self.modifier(FrameModifier(coordinateSpace: coordinateSpace, onFrameChange: perform))
    }
}
