
import SwiftUI
import UniformTypeIdentifiers

struct AnyFileImporter {
    var contentType: [UTType]
    var allowsMultipleSelection: Bool = false
    var onCompletion: (Result<[URL], any Error>) -> Void
}

extension View {
    func fileImpoter(_ importer: Binding<AnyFileImporter?>) -> some View {
        fileImporter(
            isPresented: .init(importer),
            allowedContentTypes: importer.wrappedValue?.contentType ?? [.item],
            allowsMultipleSelection: importer.wrappedValue?.allowsMultipleSelection ?? false,
            onCompletion: importer.wrappedValue?.onCompletion ?? { _ in })
    }
}

extension Binding where Value == Bool {

    init<T: Sendable>(_ value: Binding<T?>) {
        self.init {
            value.wrappedValue != nil
        } set: { newValue in
            if newValue == false {
                value.wrappedValue = nil
            }
        }
    }
}
