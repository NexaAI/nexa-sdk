import Foundation
import NexaBridge

public enum LLMError: LocalizedError, CustomStringConvertible {
    case createFailed
    case modelLoadingFailed(Int32)
    case kvCacheSaveFailed(Int32)
    case kvCacheLoadFailed(Int32)
    case generateFailed(Int32)
    case applyChatTemplateFailed(Int32)
    case generateEmptyString

    public var errorDescription: String? { desc }
    public var description: String { desc }

    var desc: String {
        switch self {
        case .createFailed:
            return "Create failed"
        case .generateEmptyString:
            return "Stream result is empty"
        case .modelLoadingFailed(let code),
             .kvCacheSaveFailed(let code),
             .kvCacheLoadFailed(let code),
             .generateFailed(let code),
             .applyChatTemplateFailed(let code):
            if let errorMessage = ml_get_error_message(ml_ErrorCode(rawValue: code)) {
                let result = String(cString: errorMessage)
                return result
            }
            return "unknow error, code: \(code)"
        }
    }
}
