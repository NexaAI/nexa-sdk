
/*
 typedef struct {
     int32_t n_ctx; // text context, 0 = from model
     int32_t n_threads; // number of threads to use for generation
     int32_t n_threads_batch; // number of threads to use for batch processing
     int32_t n_batch; // logical maximum batch size that can be submitted to llama_decode
     int32_t n_ubatch; // physical maximum batch size
     int32_t n_seq_max; // max number of sequences (i.e. distinct states for recurrent models)
     ml_Path chat_template_path; // path to chat template file, optional
     const char *chat_template_content; // content of chat template file, optional
 } ml_ModelConfig;

 */
import NexaBridge

public struct ModelConfig: Codable {
    public var nCtx: Int32
    public var nThreads: Int32
    public var nThreadsBatch: Int32
    public var nBatch: Int32
    public var nUbatch: Int32
    public var nSeqMax: Int32

    var chatTemplatePath: String?
    var chatTemplateContent: String?
    
    public static let `default`: ModelConfig = {
        let mlConfig = ml_model_config_default()
        return .init(
            nCtx: 2048,
            nThreads: mlConfig.n_threads,
            nThreadsBatch: mlConfig.n_threads_batch,
            nBatch: mlConfig.n_batch,
            nUbatch: mlConfig.n_ubatch,
            nSeqMax: mlConfig.n_seq_max,
            chatTemplatePath: nil,
            chatTemplateContent: nil
        )
    }()
}
