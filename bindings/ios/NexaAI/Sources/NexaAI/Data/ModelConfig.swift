// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

import NexaBridge

/** LLM / VLM model configuration */
/*
 typedef struct {
    int32_t     n_ctx;                  // text context, 0 = from model
    int32_t     n_threads;              // number of threads to use for generation
    int32_t     n_threads_batch;        // number of threads to use for batch processing
    int32_t     n_batch;                // logical maximum batch size that can be submitted to llama_decode
    int32_t     n_ubatch;               // physical maximum batch size
    int32_t     n_seq_max;              // max number of sequences (i.e. distinct states for recurrent models)
    ml_Path     chat_template_path;     // path to chat template file, optional
    const char* chat_template_content;  // content of chat template file, optional
    // For QNN
    ml_Path system_library_path;    /* System library path */
    ml_Path backend_library_path;   /* Backend library path */
    ml_Path extension_library_path; /* Extension library path */
    ml_Path config_file_path;       /* Config file path */
    ml_Path embedded_tokens_path;   /* Embedded tokens path */
    int32_t max_tokens;             /* Maximum tokens */
    bool    enable_thinking;        /* Enable thinking */
    bool    verbose;                /* Verbose */
} ml_ModelConfig;
*/

public struct ModelConfig: Codable {
    public var nCtx: Int32
    public var nThreads: Int32
    public var nThreadsBatch: Int32
    public var nBatch: Int32
    public var nUbatch: Int32
    public var nSeqMax: Int32

    public var chatTemplatePath: String?
    public var chatTemplateContent: String?

    public static let `default`: ModelConfig = {
        return .init(
            nCtx: 2048,
            nThreads: 0,
            nThreadsBatch: 0,
            nBatch: 0,
            nUbatch: 0,
            nSeqMax: 0,
            chatTemplatePath: nil,
            chatTemplateContent: nil
        )
    }()

    public init(
        nCtx: Int32 = 2048,
        nThreads: Int32 = 0,
        nThreadsBatch: Int32 = 0,
        nBatch: Int32 = 0,
        nUbatch: Int32 = 0,
        nSeqMax: Int32 = 0,
        chatTemplatePath: String? = nil,
        chatTemplateContent: String? = nil
    ) {
        self.nCtx = nCtx
        self.nThreads = nThreads
        self.nThreadsBatch = nThreadsBatch
        self.nBatch = nBatch
        self.nUbatch = nUbatch
        self.nSeqMax = nSeqMax
        self.chatTemplatePath = chatTemplatePath
        self.chatTemplateContent = chatTemplateContent
    }
}
