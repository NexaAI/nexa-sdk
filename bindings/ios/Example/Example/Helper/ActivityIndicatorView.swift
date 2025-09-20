
import SwiftUI

struct ActivityIndicatorView: View {
    private let color: [Color] = [.gray4, .gray6, .gray8]

    @State private var scale: [Bool] = [false, false, false]


    var body: some View {
        HStack(spacing: 4) {
            ForEach(0..<3) { i in
                Circle()
                    .fill(color[i])
                    .frame(width: 5, height: 5)
                    .opacity(scale[i] ? 1 : 0.3)
                    .animation(.easeInOut(duration: 0.6).repeatForever().delay(Double(i) * 0.2), value: scale[i])
            }
        }
        .onAppear {
            scale = [true, true, true]
        }
    }
}

#Preview {
    ActivityIndicatorView()
}

