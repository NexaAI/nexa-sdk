import SwiftUI
import NexaAI

enum Item: String, Identifiable, Hashable, CaseIterable {
    case llm
    case vlm
    case embedder
    case chatbox
    var id: String {
        self.rawValue
    }
}

struct ContentView: View {
    @State private var selectedItem: Item?

    var body: some View {
        NavigationSplitView {
            List(selection: $selectedItem) {
                ForEach(Item.allCases) { item in
                    NavigationLink(item.rawValue, value: item)
                }
            }
        } detail: {
            switch selectedItem {
            case .llm:
                LLMView().navigationTitle("LLM")
            case .vlm:
                VLMView().navigationTitle("VLM")
            case .embedder:
                EmbedderView().navigationTitle("Embedder")
            case .chatbox:
                ChatView()
            case .none:
                EmptyView()
            }
        }
    }
}


#Preview {
    ContentView()
}
