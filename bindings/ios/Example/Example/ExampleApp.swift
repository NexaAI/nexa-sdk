import SwiftUI
import NexaAI

@main
struct ExampleApp: App {
    var body: some Scene {
        WindowGroup {
            ContentView()
                .onAppear {
                    NexaSdk.install([])
                }
        }
    }
}
