import NexaBridge
import Foundation

public struct LogLevel {
    let rawValue: UInt32

    init(rawValue: UInt32) {
        self.rawValue = rawValue
    }

    public static let trace = LogLevel(rawValue: ML_LOG_LEVEL_TRACE.rawValue)
    public static let debug = LogLevel(rawValue: ML_LOG_LEVEL_DEBUG.rawValue)
    public static let info  = LogLevel(rawValue: ML_LOG_LEVEL_INFO.rawValue)
    public static let warn  = LogLevel(rawValue: ML_LOG_LEVEL_WARN.rawValue)
    public static let error = LogLevel(rawValue: ML_LOG_LEVEL_ERROR.rawValue)

}

typealias MlLogCallback = @convention(c) (ml_LogLevel, UnsafePointer<CChar>?) -> Void

private var logLevel: [LogLevel] = [.error]
private let logCallback: MlLogCallback = { level, messagePtr in
    let contains = logLevel.isEmpty ? true : logLevel.contains { $0.rawValue == level.rawValue }
    guard contains else { return }

    let message = messagePtr.flatMap { String(cString: $0) } ?? ""
    switch level {
    case ML_LOG_LEVEL_TRACE:
        print("[NEXA_TRACE] \(message)")
    case ML_LOG_LEVEL_DEBUG:
        print("[NEXA_DEBUG] \(message)")
    case ML_LOG_LEVEL_INFO:
        print("[NEXA_INFO] \(message)")
    case ML_LOG_LEVEL_WARN:
        print("[NEXA_WARN] \(message)")
    case ML_LOG_LEVEL_ERROR:
        print("[NEXA_ERROR] \(message)")
    default:
        print("[NEXA_DEFAULT] \(message)")
    }
}

private func setLogLevelGlobal(_ options: [LogLevel]) {
    logLevel = options
    ml_set_log(logCallback)
}

public struct Device {
    public let id: String
    public let name: String
    init(id: String, name: String) {
        self.id = id
        self.name = name
    }
}

public class NexaSdk {
    private init() { }

    public class var version: String {
        String(cString: ml_version())
    }

    public class func install(_ logLevel: [LogLevel] = [.error]) {
        setLogLevel(logLevel)
    }

    public class func getLlamaDeviceList() -> [Device] {
        ml_init()

        var result = ml_register_plugin(plugin_id, createLlamaPlugin)
        if result < 0 {
            ml_deinit()
            return []
        }

        var input = ml_GetDeviceListInput(plugin_id: plugin_id())
        var output = ml_GetDeviceListOutput()
        result = ml_get_device_list(&input, &output)
        if result < 0 {
            ml_deinit()
            return []
        }

        var ids = [String]()
        for i in 0..<Int(output.device_count) {
            if let cStr = output.device_ids[i] {
                let deviceId = String(cString: cStr)
                ids.append(deviceId)
            }
        }

        var names = [String]()
        for i in 0..<Int(output.device_count) {
            if let cStr = output.device_names[i] {
                let deviceName = String(cString: cStr)
                names.append(deviceName)
            }
        }

        ml_free(output.device_names)
        ml_free(output.device_ids)
        ml_deinit()

        if ids.count != names.count {
            return []
        }

        var devices = [Device]()
        for (idx, id) in ids.enumerated() {
            let device = Device(id: id, name: names[idx])
            devices.append(device)
        }
        return devices
    }

    class func setLogLevel(_ logLevel: [LogLevel]) {
        setLogLevelGlobal(logLevel)
    }
}


@globalActor
public actor NexaAIActor {
    public static let shared = NexaAIActor()
}
