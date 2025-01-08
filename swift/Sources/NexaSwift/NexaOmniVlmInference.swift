import Foundation

public class NexaOmniVlmInference {
    private var libraryHandle: UnsafeMutableRawPointer?
    
    public init?(modelPath: String, projectorModelPath: String, version: String = "vlm-81-instruct") {
        guard let resourcePath = Bundle.module.resourcePath else {
           print("Resource path not found")
           return nil
       }
       
       let ggmlllamaPath = resourcePath + "/libggml_llama.dylib"
       guard let _ = dlopen(ggmlllamaPath, RTLD_LAZY | RTLD_GLOBAL) else {
           print("Failed to load ggml_llama: \(String(cString: dlerror()))")
           return nil
       }
       
       let llamaPath = resourcePath + "/libllama.dylib"
       guard let _ = dlopen(llamaPath, RTLD_LAZY | RTLD_GLOBAL) else {
           print("Failed to load llama: \(String(cString: dlerror()))")
           return nil
       }
       
       let dylibPath = resourcePath + "/libomni_vlm_wrapper_shared.dylib"
       guard let handle = dlopen(dylibPath, RTLD_LAZY) else {
           print("Failed to load main library: \(String(cString: dlerror()))")
           return nil
       }
       
       libraryHandle = handle
       print("Libraries loaded successfully")
       
       // 初始化模型
       typealias OmnivlmInitFunction = @convention(c) (UnsafePointer<CChar>, UnsafePointer<CChar>, UnsafePointer<CChar>) -> Void
       
       guard let initPointer = dlsym(handle, "omnivlm_init") else {
           print("Function 'omnivlm_init' not found")
           dlclose(handle)
           return nil
       }
       
       let omnivlmInit = unsafeBitCast(initPointer, to: OmnivlmInitFunction.self)
       omnivlmInit(
           (modelPath as NSString).utf8String!,
           (projectorModelPath as NSString).utf8String!,
           (version as NSString).utf8String!
       )
       print("Model initialized successfully")
    }
    
    deinit {
        if let handle = libraryHandle {
            dlclose(handle)
            print("Library unloaded")
        }
    }
    
    @NexaSwiftActor
    public func inferenceStreaming(prompt: String, imagePath: String) -> AsyncStream<String> {
        return AsyncStream { continuation in
            guard let handle = libraryHandle else {
                print("Library not loaded")
                continuation.finish()
                return
            }

            guard let streamingPtr = dlsym(handle, "omnivlm_inference_streaming"),
                  let samplePtr = dlsym(handle, "sample"),
                  let getStrPtr = dlsym(handle, "get_str") else {
                print("Failed to find required functions")
                continuation.finish()
                return
            }

            typealias StreamingFunction = @convention(c) (UnsafePointer<CChar>, UnsafePointer<CChar>) -> OpaquePointer
            typealias SampleFunction = @convention(c) (OpaquePointer) -> Int32
            typealias GetStrFunction = @convention(c) (OpaquePointer) -> UnsafePointer<CChar>

            let streamingFunc = unsafeBitCast(streamingPtr, to: StreamingFunction.self)
            let sampleFunc = unsafeBitCast(samplePtr, to: SampleFunction.self)
            let getStrFunc = unsafeBitCast(getStrPtr, to: GetStrFunction.self)

            Task {
                let sample = streamingFunc(
                    (prompt as NSString).utf8String!,
                    (imagePath as NSString).utf8String!
                )

                var res: Int32 = 0
                repeat {
                    res = sampleFunc(sample)

                    if let str = String(cString: getStrFunc(sample), encoding: .utf8) {
                        if !str.contains("<|im_start|>") && !str.contains("</s>") {
                            continuation.yield(str)
                        }
                    }
                } while res >= 0

                continuation.finish()
            }
        }
    }
}
