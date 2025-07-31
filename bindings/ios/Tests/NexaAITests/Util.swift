import Foundation

func resoucePath(of name: String, ext: String = "gguf") throws -> String {
#if os(macOS)
    let currentFile = URL(fileURLWithPath: #file)
    print(currentFile.absoluteString)
    let modelPath = currentFile
        .deletingLastPathComponent()
        .appendingPathComponent("../../../../../modelfiles/\(name).\(ext)")
        .standardized.path
    return modelPath
#else
    guard let modelFileURL = Bundle.module.url(forResource: name, withExtension: ext) else {
        throw TestError.fileNotFound(name)
    }
    return modelFileURL.path()
#endif
}

enum TestError: LocalizedError {
    case fileNotFound(_ name: String)
    var errorDescription: String? {
        switch self {
        case .fileNotFound(let name):
            return "\"\(name)\" not found, please copy it to modelfiles dir"
        }
    }
}
