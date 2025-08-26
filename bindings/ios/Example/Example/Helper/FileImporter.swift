
import Foundation
import SwiftUI
import UniformTypeIdentifiers

struct AnyFileImporter {
    var contentType: [UTType]
    var onCompletion: (Result<URL, any Error>) -> Void
}

extension View {

    func fileImpoter(_ importer: Binding<AnyFileImporter?>) -> some View {
        fileImporter(isPresented: .init(importer), allowedContentTypes: importer.wrappedValue?.contentType ?? [.item], onCompletion: importer.wrappedValue?.onCompletion ?? {_ in })
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

extension FileManager {
    static func copyToDocumentsDirectory(from sourceURL: URL) throws -> URL {
        let fileManager = FileManager.default
        let destinationURL = fileManager.urls(for: .documentDirectory, in: .userDomainMask)[0]
            .appendingPathComponent(sourceURL.lastPathComponent)

        if fileManager.fileExists(atPath: destinationURL.path) {
            try fileManager.removeItem(at: destinationURL)
        }

        try fileManager.copyItem(at: sourceURL, to: destinationURL)
        return destinationURL
    }
}
