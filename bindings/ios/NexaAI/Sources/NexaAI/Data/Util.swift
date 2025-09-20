
import Foundation

extension Array where Element == String {

    // [String] -> const char **
    func withUnsafeMutableBufferPointerC<T>(_ body: (UnsafeMutablePointer<UnsafePointer<CChar>?>?) throws -> T) rethrows -> T {
        let cStringPointers: [UnsafeMutablePointer<CChar>?] = map { strdup($0) }
        var constCStringPointers: [UnsafePointer<CChar>?] = cStringPointers.map { UnsafePointer($0) }
        return try constCStringPointers.withUnsafeMutableBufferPointer { buffer in
            defer {
                for ptr in cStringPointers {
                    if let p = ptr { free(p) }
                }
            }
            return try body(buffer.baseAddress)
        }
    }

}

extension Array  {

    func withUnsafeBufferPointerC<T>(_ body: (UnsafePointer<UnsafePointer<Element>?>?) -> T) -> T {
        return withUnsafeBufferPointer { buffer in
            let ptrs = [buffer.baseAddress]
            return ptrs.withUnsafeBufferPointer { body($0.baseAddress) }
        }
    }

    mutating func withUnsafeMutableBufferPointerC<T>(_ body: (UnsafeMutablePointer<UnsafeMutablePointer<Element>?>?) -> T) -> T {
        return withUnsafeMutableBufferPointer { buffer in
            var ptrs = [buffer.baseAddress]
            return ptrs.withUnsafeMutableBufferPointer { body($0.baseAddress) }
        }
    }
}
