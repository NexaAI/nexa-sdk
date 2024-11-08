import SwiftUI

struct ContentView: View {
    @State private var viewModel = ViewModel()
    @State private var prompt = ""
    @FocusState private var isInputActive: Bool

    var body: some View {
        VStack {
            Text("Nexa Swift Demo").font(.title)
            
            Toggle(isOn: $viewModel.usingStream) {
                Text("Use Stream")
            }
            .padding(.bottom)
            
            TextField("Enter your message", text: $prompt, axis: .vertical)
                .textFieldStyle(.roundedBorder)
                .lineLimit(3...5)
                .padding(.bottom)
                .onSubmit {
                    guard !prompt.isEmpty else { return }
                    viewModel.run(for: prompt)
                }
                .focused($isInputActive)
            
            Button(action: {
                guard !prompt.isEmpty else { return }
                viewModel.run(for: prompt)
                isInputActive = false
            }) {
                Text("Send")
                    .frame(maxWidth: .infinity)
            }
            .buttonStyle(.borderedProminent)
            .padding(.bottom)
            
            ScrollView {
                Text(viewModel.result)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .textSelection(.enabled)
            }
            
            Spacer()
        }
        .padding()
    }
}

#Preview {
    ContentView()
}
