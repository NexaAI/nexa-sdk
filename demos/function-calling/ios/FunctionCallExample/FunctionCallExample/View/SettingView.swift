
import SwiftUI

struct SettingView: View {

    @Binding var ipAddress: String

    @Environment(\.dismiss) var dismiss

    var body: some View {
        VStack {
            HStack {
                Text("Generation Settings")
                    .font(.system(size: 16, weight: .medium))
                Spacer()
                Image(.close)
                    .anyButton {
                        dismiss()
                    }
            }
            .padding(.vertical, 20)

            VStack(alignment: .leading, spacing: 16) {
                Text("Server IP Address")
                    .font(.system(size: 14, weight: .medium))
                    .foregroundStyle(Color.Text.primary)

                TextField("", text: $ipAddress)
                    .textStyle(.body2(textColor: Color.Input.Font.active))
                    .frame(maxWidth: .infinity)
                    .padding(.horizontal, 12)
                    .padding(.vertical, 8)
                    .background(
                        RoundedRectangle(cornerRadius: 8)
                            .fill(Color.Input.Bg.default)
                            .stroke(Color.Input.Border.default, lineWidth: 1)
                    )
            }

            Spacer()
        }
        .padding(.horizontal, 16)
    }
}


#Preview {
    SettingView(ipAddress: .constant("192.168.1.107"))
}
