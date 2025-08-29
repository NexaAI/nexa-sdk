#pragma once

/**
 * @file ml.h
 * @brief Unified C API for machine learning operations
 *
 * This header provides a comprehensive C interface for various ML tasks including:
 * - Language models (LLM) and multimodal models (VLM)
 * - Text embeddings and reranking
 * - Image generation and computer vision (OCR)
 * - Speech recognition (ASR) and text-to-speech (TTS)
 *
 * All functions return status codes where applicable, with negative values indicating errors.
 * Memory management follows RAII principles - use corresponding destroy/free functions.
 */

#include <stdbool.h>
#include <stdint.h>

#ifdef ML_SHARED
#if defined(_WIN32) && !defined(__MINGW32__)
#ifdef ML_BUILD
#define ML_API __declspec(dllexport)
#else
#define ML_API __declspec(dllimport)
#endif
#else
#define ML_API __attribute__((visibility("default")))
#endif
#else
#define ML_API
#endif

#ifdef __cplusplus
extern "C" {
#endif

/** Error code enumeration for ML operations */
typedef enum {
    /* Success */
    ML_SUCCESS = 0, /**< Operation completed successfully */

    /* ====================================================================== */
    /*                              COMMON ERRORS (100xxx)                     */
    /* ====================================================================== */

    /* General errors */
    ML_ERROR_COMMON_UNKNOWN           = -100000, /**< Unknown error */
    ML_ERROR_COMMON_INVALID_INPUT     = -100001, /**< Invalid input parameters or handle */
    ML_ERROR_COMMON_MEMORY_ALLOCATION = -100003, /**< Memory allocation failed */
    ML_ERROR_COMMON_FILE_NOT_FOUND    = -100004, /**< File not found or inaccessible */
    ML_ERROR_COMMON_NOT_INITIALIZED   = -100007, /**< Library not initialized */
    ML_ERROR_COMMON_NOT_SUPPORTED     = -100013, /**< Operation not supported */

    /* Model errors */
    ML_ERROR_COMMON_MODEL_LOAD    = -100201, /**< Model loading failed */
    ML_ERROR_COMMON_MODEL_INVALID = -100203, /**< Invalid model format */

    /* Embedding errors */
    ML_ERROR_COMMON_EMBEDDING_GENERATION = -100301, /**< Embedding generation failed */
    ML_ERROR_COMMON_EMBEDDING_DIMENSION  = -100302, /**< Invalid embedding dimension */

    /* Reranking errors */
    ML_ERROR_COMMON_RERANK_FAILED = -100401, /**< Reranking failed */
    ML_ERROR_COMMON_RERANK_INPUT  = -100402, /**< Invalid reranking input */

    /* Image generation errors */
    ML_ERROR_COMMON_IMG_GENERATION = -100501, /**< Image generation failed */
    ML_ERROR_COMMON_IMG_PROMPT     = -100502, /**< Invalid image prompt */
    ML_ERROR_COMMON_IMG_DIMENSION  = -100503, /**< Invalid image dimensions */

    /* Validation errors */
    ML_ERROR_COMMON_LICENSE_INVALID = -100601, /**< Invalid license */
    ML_ERROR_COMMON_LICENSE_EXPIRED = -100602, /**< License expired */

    /* ====================================================================== */
    /*                              LLM ERRORS (200xxx)                        */
    /* ====================================================================== */

    ML_ERROR_LLM_TOKENIZATION_FAILED         = -200001, /**< Tokenization failed */
    ML_ERROR_LLM_TOKENIZATION_CONTEXT_LENGTH = -200004, /**< Context length exceeded */

    ML_ERROR_LLM_GENERATION_FAILED          = -200101, /**< Text generation failed */
    ML_ERROR_LLM_GENERATION_PROMPT_TOO_LONG = -200103, /**< Input prompt too long */

    /* ====================================================================== */
    /*                              VLM ERRORS (300xxx)                        */
    /* ====================================================================== */

    /* Image/Audio processing */
    ML_ERROR_VLM_IMAGE_LOAD   = -300001, /**< Image loading failed */
    ML_ERROR_VLM_IMAGE_FORMAT = -300002, /**< Unsupported image format */
    ML_ERROR_VLM_AUDIO_LOAD   = -300101, /**< Audio loading failed */
    ML_ERROR_VLM_AUDIO_FORMAT = -300102, /**< Unsupported audio format */

    /* Generation */
    ML_ERROR_VLM_GENERATION_FAILED = -300201, /**< Multimodal generation failed */

    /* ====================================================================== */
    /*                              OCR ERRORS (400xxx)                        */
    /* ====================================================================== */

    ML_ERROR_OCR_DETECTION   = -400001, /**< OCR text detection failed */
    ML_ERROR_OCR_RECOGNITION = -400002, /**< OCR text recognition failed */
    ML_ERROR_OCR_MODEL       = -400003, /**< OCR model error */

    /* ====================================================================== */
    /*                              ASR ERRORS (500xxx)                        */
    /* ====================================================================== */

    ML_ERROR_ASR_TRANSCRIPTION = -500001, /**< ASR transcription failed */
    ML_ERROR_ASR_AUDIO_FORMAT  = -500002, /**< Unsupported ASR audio format */
    ML_ERROR_ASR_LANGUAGE      = -500003, /**< Unsupported ASR language */

    /* ====================================================================== */
    /*                              TTS ERRORS (600xxx)                        */
    /* ====================================================================== */

    ML_ERROR_TTS_SYNTHESIS    = -600001, /**< TTS synthesis failed */
    ML_ERROR_TTS_VOICE        = -600002, /**< TTS voice not found */
    ML_ERROR_TTS_AUDIO_FORMAT = -600003, /**< TTS audio format error */

} ml_ErrorCode;

/** Get error message string for error code */
ML_API const char* ml_get_error_message(const ml_ErrorCode error_code);

/* ========================================================================== */
/*                              CORE TYPES & UTILITIES                         */
/* ========================================================================== */

/** Plugin Id string type - plain char* for plugin id
 * support slash extention like "llama_cpp/Vulkan0" to specify device usage
 * @ref ml_get_device_list device must in list of plugin ids
 */
typedef const char* ml_PluginId;

/** Path string type - plain char* for file paths */
typedef const char* ml_Path;

typedef enum {
    ML_LOG_LEVEL_TRACE, /* Trace messages */
    ML_LOG_LEVEL_DEBUG, /* Debug messages */
    ML_LOG_LEVEL_INFO,  /* Informational messages */
    ML_LOG_LEVEL_WARN,  /* Warning messages */
    ML_LOG_LEVEL_ERROR  /* Error messages */
} ml_LogLevel;

/** Logging callback function type */
typedef void (*ml_log_callback)(ml_LogLevel, const char*);

/** Token callback for streaming generation */
typedef bool (*ml_token_callback)(const char* token, void* user_data);

/** Tool definition */
typedef struct {
    const char* name;            /** name of the function */
    const char* description;     /** description of the function in natural language */
    const char* parameters_json; /** JSON schema for the function parameters */
} ml_ToolFunction;

typedef struct {
    const char*            type;     /** always "function" */
    const ml_ToolFunction* function; /** pointer to ToolFunction */
} ml_Tool;

/** Input structure for saving KV cache */
typedef struct {
    ml_Path path; /** Path to save the KV cache */
} ml_KvCacheSaveInput;

/** Output structure for saving KV cache (empty for now) */
typedef struct {
    void* reserved; /** Reserved for future use, safe to set as NULL */
} ml_KvCacheSaveOutput;

/** Input structure for loading KV cache */
typedef struct {
    ml_Path path; /** Path to load the KV cache from */
} ml_KvCacheLoadInput;

/** Output structure for loading KV cache (empty for now) */
typedef struct {
    void* reserved; /** Reserved for future use, safe to set as NULL */
} ml_KvCacheLoadOutput;

/* ====================  Core Initialization  ================================ */

/**
 * @brief Initialize the ML C-Lib runtime, starting the life cycle of the library.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe.
 */
ML_API int32_t ml_init(void);

/** Plugin id create function type */
typedef ml_PluginId (*ml_plugin_id_func)();

/** Plugin instance create function type */
typedef void* (*ml_create_plugin_func)();

/**
 * @brief Register a custom plugin with the ML C-Lib runtime.
 *
 * @param plugin_id_func[in]: The pointer to plugin create_id function.
 * @param create_func[in]: The pointer to plugin create function.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Thread-safe.
 */
ML_API int32_t ml_register_plugin(ml_plugin_id_func plugin_id_func, ml_create_plugin_func create_func);

/**
 * @brief Deinitialize the ML C-Lib runtime, ending the life cycle of the library.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe.
 */
ML_API int32_t ml_deinit(void);

/**
 * @brief Set custom logging callback function, call before init
 *
 * @param callback[in]: The callback function to set.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Thread-safe
 */
ML_API int32_t ml_set_log(ml_log_callback callback);

/**
 * @brief Simple wrapper around free() to free memory allocated by ML library functions
 *
 * @param ptr[in]: The pointer to free.
 *
 * @thread_safety: Thread-safe if called for different pointers.
 */
ML_API void ml_free(void* ptr);

/**
 * @brief Get Library Version
 *
 * @param out_version[out]: Pointer to the library version.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Thread-safe.
 */
ML_API const char* ml_version(void);

/** Output structure containing the list of available plugins */
typedef struct {
    ml_PluginId* plugin_ids;   /**< Array of plugin IDs (UTF-8) (caller must free with ml_free) */
    int32_t      plugin_count; /**< Number of plugin IDs in the list */
} ml_GetPluginListOutput;

/**
 * @brief Query the list of available plugins.
 *
 * @param output[out] Pointer to plugin list and count. caller must free with `ml_free`.
 *
 * @return ml_ErrorCode ML_SUCCESS on success, negative value on failure.
 *
 * @thread_safety: Not thread-safe.
 *
 * @note The returned plugin_list TODO
 */
ML_API int32_t ml_get_plugin_list(ml_GetPluginListOutput* output);

/** Input structure for querying available devices for a plugin */
typedef struct {
    ml_PluginId plugin_id; /**< Plugin identifier */
} ml_GetDeviceListInput;

/** Output structure containing the list of available devices */
typedef struct {
    // example: Vlukan0, name is llama_cpp, device is cpu
    const char** device_ids;   /**< Array of device IDs  (caller must free with ml_free when not null) */
    const char** device_names; /**< Array of device names  (caller must free with ml_free when not null) */
    int32_t      device_count; /**< Number of device names in the list */
} ml_GetDeviceListOutput;

/**
 * @brief Query the list of available devices for a given plugin.
 *
 * @param input[in]   Pointer to input structure specifying the plugin.
 * @param output[out] Pointer to output structure to receive device list and count.
 *
 * @return ml_ErrorCode ML_SUCCESS on success, negative value on failure.
 *
 * @thread_safety: Not thread-safe.
 *
 * @note The returned device_list TODO
 */
ML_API int32_t ml_get_device_list(const ml_GetDeviceListInput* input, ml_GetDeviceListOutput* output);

/* ====================  Data Structures  ==================================== */

/** Image data structure */
typedef struct {
    float*  data;     /* Raw pixel data: width × height × channels */
    int32_t width;    /* Image width in pixels */
    int32_t height;   /* Image height in pixels */
    int32_t channels; /* Color channels: 3=RGB, 4=RGBA */
} ml_Image;

/** Free image data structure and its pixel data */
ML_API void ml_image_free(ml_Image* image);
ML_API void ml_image_save(const ml_Image* image, const char* filename); /* Save image to file */

/** Audio data structure */
typedef struct {
    float*  data;        /* Audio samples: num_samples × channels */
    int32_t sample_rate; /* Sample rate in Hz */
    int32_t channels;    /* Audio channels: 1=mono, 2=stereo */
    int32_t num_samples; /* Number of samples per channel */
} ml_Audio;

/** Free audio data structure and its sample data */
ML_API void ml_audio_free(ml_Audio* audio);

/** Video data structure */
typedef struct {
    float*  data;       /* Frame data: width × height × channels × num_frames */
    int32_t width;      /* Frame width in pixels */
    int32_t height;     /* Frame height in pixels */
    int32_t channels;   /* Color channels per frame */
    int32_t num_frames; /* Number of video frames */
} ml_Video;

/** Free video data structure and its frame data */
ML_API void ml_video_free(ml_Video* video);

/** Profile data structure for performance metrics */
typedef struct {
    int64_t ttft;        /* Time to first token (us) */
    int64_t prompt_time; /* Prompt processing time (us) */
    int64_t decode_time; /* Token generation time (us) */

    int64_t prompt_tokens;    /* Number of prompt tokens */
    int64_t generated_tokens; /* Number of generated tokens */
    int64_t audio_duration;   /* Audio duration (us) */

    double prefill_speed;    /* Prefill speed (tokens/sec) */
    double decoding_speed;   /* Decoding speed (tokens/sec) */
    double real_time_factor; /* Real-Time Factor(RTF) (1.0 = real-time, >1.0 = faster, <1.0 = slower) */

    const char* stop_reason; /* Stop reason: "eos", "length", "user", "stop_sequence" */
} ml_ProfileData;

/* ========================================================================== */
/*                              LANGUAGE MODELS (LLM)                          */
/* ========================================================================== */

/** Text generation sampling parameters */
typedef struct {
    float       temperature;        /* Sampling temperature (0.0-2.0) */
    float       top_p;              /* Nucleus sampling parameter (0.0-1.0) */
    int32_t     top_k;              /* Top-k sampling parameter */
    float       min_p;              /* Minimum probability for nucleus sampling */
    float       repetition_penalty; /* Penalty for repeated tokens */
    float       presence_penalty;   /* Penalty for token presence */
    float       frequency_penalty;  /* Penalty for token frequency */
    int32_t     seed;               /* Random seed (-1 for random) */
    ml_Path     grammar_path;       /* Optional grammar file path */
    const char* grammar_string;     /* Optional grammar string (BNF-like format) */
} ml_SamplerConfig;

/** LLM / VLM generation configuration (IMPROVED: support multiple images and audios) */
typedef struct {
    int32_t           max_tokens;     /* Maximum tokens to generate */
    const char**      stop;           /* Array of stop sequences */
    int32_t           stop_count;     /* Number of stop sequences */
    int32_t           n_past;         /* Number of past tokens to consider */
    ml_SamplerConfig* sampler_config; /* Advanced sampling config */
    // --- Improved multimodal support ---
    ml_Path* image_paths; /* Array of image paths for VLM (NULL if none) */
    int32_t  image_count; /* Number of images */
    ml_Path* audio_paths; /* Array of audio paths for VLM (NULL if none) */
    int32_t  audio_count; /* Number of audios */
} ml_GenerationConfig;

/** LLM / VLM model configuration */
typedef struct {
    int32_t     n_ctx;                  // text context, 0 = from model
    int32_t     n_threads;              // number of threads to use for generation
    int32_t     n_threads_batch;        // number of threads to use for batch processing
    int32_t     n_batch;                // logical maximum batch size that can be submitted to llama_decode
    int32_t     n_ubatch;               // physical maximum batch size
    int32_t     n_seq_max;              // max number of sequences (i.e. distinct states for recurrent models)
    int32_t     n_gpu_layers;           // number of layers to offload to GPU, 0 = all layers on CPU
    ml_Path     chat_template_path;     // path to chat template file, optional
    const char* chat_template_content;  // content of chat template file, optional
    bool        enable_sampling;       // enable sampling
    const char* grammar_str;            // grammar string
    // For QNN
    ml_Path model_path_1;             /* Model path */
    ml_Path system_library_path;    /* System library path */
    ml_Path backend_library_path;   /* Backend library path */
    ml_Path extension_library_path; /* Extension library path */
    ml_Path config_file_path;       /* Config file path */
    ml_Path embedded_tokens_path;   /* Embedded tokens path */
    int32_t max_tokens;             /* Maximum tokens */
    bool    enable_thinking;        /* Enable thinking */
    bool    verbose;                /* Verbose */
    // QNN Vision-specific paths
    ml_Path patch_embed_path;       /* Vision patch embedding model path */
    ml_Path vit_model_path;         /* Vision transformer model path */
    ml_Path vit_config_file_path;   /* Vision model config file path */
    // QNN Audio-specific paths
    ml_Path audio_encoder_helper0_path; /* Audio encoder helper 0 model path */
    ml_Path audio_encoder_helper1_path; /* Audio encoder helper 1 model path */
    ml_Path audio_encoder_model_path;   /* Audio encoder model path */
    ml_Path audio_encoder_config_file_path; /* Audio encoder config file path */
} ml_ModelConfig;

/* ====================  LLM Handle  ======================================== */
typedef struct ml_LLM ml_LLM; /* Opaque LLM handle */

/* ====================  Lifecycle Management  ============================== */
typedef struct {
    ml_Path        model_path;     /** Path to the model file */
    ml_Path        tokenizer_path; /** Path to the tokenizer file */
    ml_ModelConfig config;         /** Model configuration */
    ml_PluginId    plugin_id;      /** plugin to use for the model */
    const char*    device_id;      /** device to use for the model, NULL for default device */
    const char*    license_id;     /** licence id for loading NPU models, must be provided upon the first use of the license key. null terminated string */
    const char*    license_key;    /** licence key for loading NPU models, null terminated string */
} ml_LlmCreateInput;

/**
 * @brief Create and initialize an LLM instance from model files
 *
 * @param input[in]: Input parameters for the LLM creation
 * @param out_handle[out]: Pointer to the LLM handle. Must be freed with ml_llm_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_llm_create(const ml_LlmCreateInput* input, ml_LLM** out_handle);

/**
 * @brief Destroy LLM instance and free associated resources
 *
 * @param handle[in]: The LLM handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_llm_destroy(ml_LLM* handle);

/**
 * @brief Reset LLM internal state (clear KV cache, reset sampling)
 *
 * @param handle[in]: The LLM handle to reset.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_llm_reset(ml_LLM* handle);

/* ====================  KV-Cache Management  ============================== */

/**
 * @brief Save current KV cache state to file
 *
 * @param handle[in]: LLM handle
 * @param input[in]: Input parameters for saving KV cache
 * @param output[out]: Reserved struct for future use, safe to pass nullptr now
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_llm_save_kv_cache(ml_LLM* handle, const ml_KvCacheSaveInput* input, ml_KvCacheSaveOutput* output);

/**
 * @brief Load KV cache state from file
 *
 * @param handle[in]: LLM handle
 * @param input[in]: Input parameters for loading KV cache
 * @param output[out]: Reserved struct for future use, safe to pass nullptr now
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_llm_load_kv_cache(ml_LLM* handle, const ml_KvCacheLoadInput* input, ml_KvCacheLoadOutput* output);

/* ====================  Chat Template ================================== */

/** Chat message structure */
typedef struct {
    const char* role;    /* Message role: "user", "assistant", "system" */
    const char* content; /* Message content in UTF-8 */
} ml_LlmChatMessage;

/** Input structure for applying chat template */
typedef struct {
    ml_LlmChatMessage* messages;        /** Array of chat messages */
    int32_t            message_count;   /** Number of messages */
    ml_Tool*           tools;           /** Array of tools (optional, can be NULL) */
    int32_t            tool_count;      /** Number of tools */
    bool               enable_thinking; /** Enable thinking */
} ml_LlmApplyChatTemplateInput;

/** Output structure for applying chat template */
typedef struct {
    char* formatted_text; /** Formatted chat text (caller must free with ml_free) */
} ml_LlmApplyChatTemplateOutput;

/**
 * @brief Apply chat template to messages
 *
 * @param handle[in]: LLM handle
 * @param input[in]: Input parameters for applying chat template
 * @param output[out]: Output data containing the formatted text
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_llm_apply_chat_template(
    ml_LLM* handle, const ml_LlmApplyChatTemplateInput* input, ml_LlmApplyChatTemplateOutput* output);

/* ====================  Streaming Generation  ============================= */

/** Input structure for streaming text generation */
typedef struct {
    const char*                prompt_utf8; /** The full chat history as UTF-8 string */
    const ml_GenerationConfig* config;      /** Generation configuration (optional, can be nullptr) */
    ml_token_callback          on_token;    /** Token callback function for streaming */
    void*                      user_data;   /** User data passed to callback (optional, can be nullptr) */
} ml_LlmGenerateInput;

/** Output structure for streaming text generation */
typedef struct {
    char*          full_text;    /** Complete generated text (caller must free with ml_free) */
    ml_ProfileData profile_data; /** Profiling data for the generation */
} ml_LlmGenerateOutput;

/**
 * @brief Generate text with streaming token callback
 *
 * @param handle[in]: LLM handle
 * @param input[in]: Input parameters for streaming generation
 * @param output[out]: Output containing the complete generated text
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_llm_generate(ml_LLM* handle, const ml_LlmGenerateInput* input, ml_LlmGenerateOutput* output);

/* ========================================================================== */
/*                              MULTIMODAL MODELS (VLM)                          */
/* ========================================================================== */

typedef struct {
    const char* type;  // "text", "image", "audio", … (null-terminated UTF-8)
    const char* text;  // payload: the actual text, URL, or special token
} ml_VlmContent;

/* ---------- Message ---------- */
typedef struct {
    const char*    role;           // "user", "assistant", "system", …
    ml_VlmContent* contents;       // dynamically-allocated array (may be NULL)
    int64_t        content_count;  // number of elements in `contents`
} ml_VlmChatMessage;

typedef struct ml_VLM ml_VLM; /* Opaque VLM handle */

/* ====================  Lifecycle Management  ============================== */

typedef struct {
    ml_Path        model_path;     /** Path to the model file */
    ml_Path        mmproj_path;    /** Path to the mmproj file */
    ml_Path        tokenizer_path; /** Path to the tokenizer file */
    ml_ModelConfig config;         /** Model configuration */
    ml_PluginId    plugin_id;      /** Plugin to use for the model */
    const char*    device_id;      /** device to use for the model */
    const char*    license_id;     /** licence id for loading NPU models, must be provided upon the first use of the license key. null terminated string */
    const char*    license_key;    /** licence key for loading NPU models, null terminated string */
} ml_VlmCreateInput;

/**
 * @brief Create and initialize a VLM instance from model files
 *
 * @param input[in]: Input parameters for the VLM creation
 * @param out_handle[out]: Pointer to the VLM handle. Must be freed with ml_vlm_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_vlm_create(const ml_VlmCreateInput* input, ml_VLM** out_handle);

/**
 * @brief Destroy VLM instance and free associated resources
 *
 * @param handle[in]: The VLM handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_vlm_destroy(ml_VLM* handle);

/**
 * @brief Reset VLM internal state (clear KV cache, reset sampling)
 *
 * @param handle[in]: The VLM handle to reset.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_vlm_reset(ml_VLM* handle);

/* ====================  Text Generation  ================================== */

/** Input structure for applying VLM chat template */
typedef struct {
    ml_VlmChatMessage* messages;        /** Array of chat messages */
    int32_t            message_count;   /** Number of messages */
    ml_Tool*           tools;           /** Array of tools (optional, can be NULL) */
    int32_t            tool_count;      /** Number of tools */
    bool               enable_thinking; /** Enable thinking */
} ml_VlmApplyChatTemplateInput;

/** Output structure for applying VLM chat template */
typedef struct {
    char* formatted_text; /** Formatted chat text (caller must free with ml_free) */
} ml_VlmApplyChatTemplateOutput;

/**
 * @brief Apply chat template to messages
 *
 * @param handle[in]: VLM handle
 * @param input[in]: Input parameters for applying chat template
 * @param output[out]: Output data containing the formatted text
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_vlm_apply_chat_template(
    ml_VLM* handle, const ml_VlmApplyChatTemplateInput* input, ml_VlmApplyChatTemplateOutput* output);

/* ====================  Streaming Generation  ============================= */

/** Input structure for VLM streaming text generation */
typedef struct {
    const char*                prompt_utf8; /** The full chat history as UTF-8 string */
    const ml_GenerationConfig* config;      /** Generation configuration (optional, can be nullptr) */
    ml_token_callback          on_token;    /** Token callback function for streaming */
    void*                      user_data;   /** User data passed to callback (optional, can be nullptr) */
} ml_VlmGenerateInput;

/** Output structure for VLM streaming text generation */
typedef struct {
    char*          full_text;    /** Complete generated text (caller must free with ml_free) */
    ml_ProfileData profile_data; /** Profiling data for the generation */
} ml_VlmGenerateOutput;

/**
 * @brief Generate text with streaming token callback
 *
 * @param handle[in]: VLM handle
 * @param input[in]: Input parameters for streaming generation
 * @param output[out]: Output containing the complete generated text
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_vlm_generate(ml_VLM* handle, const ml_VlmGenerateInput* input, ml_VlmGenerateOutput* output);

/* ========================================================================== */
/*                              EMBEDDING MODELS                               */
/* ========================================================================== */

/** Embedding generation configuration */
typedef struct {
    int32_t     batch_size;       /* Processing batch size */
    bool        normalize;        /* Whether to normalize embeddings */
    const char* normalize_method; /* Normalization: "l2", "mean", "none" */
} ml_EmbeddingConfig;

typedef struct ml_Embedder ml_Embedder; /* Opaque embedder handle */

/* ====================  Lifecycle Management  ============================== */

/** Input structure for creating an embedder */
typedef struct {
    ml_Path     model_path;     /** Path to the model file */
    ml_Path     tokenizer_path; /** Path to the tokenizer file */
    ml_PluginId plugin_id;      /** Plugin to use for the model */
} ml_EmbedderCreateInput;

/**
 * @brief Create and initialize an embedder instance from model files
 *
 * @param input[in]: Input parameters for the embedder creation
 * @param out_handle[out]: Pointer to the embedder handle. Must be freed with ml_embedder_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_embedder_create(const ml_EmbedderCreateInput* input, ml_Embedder** out_handle);

/**
 * @brief Destroy embedder instance and free associated resources
 *
 * @param handle[in]: The embedder handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_embedder_destroy(ml_Embedder* handle);

/* ====================  Embedding Generation  ============================= */

/** Input structure for embedding generation */
typedef struct {
    const char**              texts;        /** Array of input texts in UTF-8 encoding */
    int32_t                   text_count;   /** Number of input texts */
    const ml_EmbeddingConfig* config;       /** Embedding configuration (optional, can be nullptr) */
    const int32_t**           input_ids_2d; /** 2D array of already tokenized raw input ids.
                                             * When passed in, texts will be ignored.
                                             * NOTE: this is supported for llama-cpp backend only.
                                             * Passing this param to other backends will be ignored */
    const int32_t* input_ids_row_lengths;   /** Array containing the length of each row in input_ids_2d */
    int32_t        input_ids_row_count;     /** Number of rows in input_ids_2d array */
} ml_EmbedderEmbedInput;

/** Output structure for embedding generation */
typedef struct {
    float*         embeddings;      /** Output embeddings array (caller must free with ml_free) */
    int32_t        embedding_count; /** Number of embeddings returned */
    ml_ProfileData profile_data;    /** Profiling data for the embedding generation */
} ml_EmbedderEmbedOutput;

/**
 * @brief Generate embeddings for input texts
 *
 * @param handle[in]: Embedder handle
 * @param input[in]: Input parameters for embedding generation
 * @param output[out]: Output data containing the generated embeddings
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_embedder_embed(
    ml_Embedder* handle, const ml_EmbedderEmbedInput* input, ml_EmbedderEmbedOutput* output);

/* ====================  Model Information  ================================ */

/** Output structure for getting embedding dimension */
typedef struct {
    int32_t dimension; /** The embedding dimension size */
} ml_EmbedderDimOutput;

/**
 * @brief Get embedding dimension from the model
 *
 * @param handle[in]: Embedder handle
 * @param output[out]: Output data containing the embedding dimension
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Thread-safe
 */
ML_API int32_t ml_embedder_embedding_dim(const ml_Embedder* handle, ml_EmbedderDimOutput* output);

/* ========================================================================== */
/*                              RERANKING MODELS                               */
/* ========================================================================== */

/** Reranking configuration */
typedef struct {
    int32_t     batch_size;       /* Processing batch size */
    bool        normalize;        /* Whether to normalize scores */
    const char* normalize_method; /* Normalization: "softmax", "min-max", "none" */
} ml_RerankConfig;

typedef struct ml_Reranker ml_Reranker; /* Opaque reranker handle */

/* ====================  Lifecycle Management  ============================== */

/** Input structure for creating a reranker */
typedef struct {
    ml_Path     model_path;     /** Path to the model file */
    ml_Path     tokenizer_path; /** Path to the tokenizer file */
    ml_PluginId plugin_id;      /** Plugin to use for the model */
} ml_RerankerCreateInput;

/**
 * @brief Create and initialize a reranker instance from model files
 *
 * @param input[in]: Input parameters for the reranker creation
 * @param out_handle[out]: Pointer to the reranker handle. Must be freed with ml_reranker_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_reranker_create(const ml_RerankerCreateInput* input, ml_Reranker** out_handle);

/**
 * @brief Destroy reranker instance and free associated resources
 *
 * @param handle[in]: The reranker handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_reranker_destroy(ml_Reranker* handle);

/* ====================  Reranking  ========================================= */

/** Input structure for reranking operation */
typedef struct {
    const char*            query;           /** Query text in UTF-8 encoding */
    const char**           documents;       /** Array of document texts in UTF-8 encoding */
    int32_t                documents_count; /** Number of documents */
    const ml_RerankConfig* config;          /** Reranking configuration (optional, can be nullptr) */
} ml_RerankerRerankInput;

/** Output structure for reranking operation */
typedef struct {
    float*         scores;       /** Output ranking scores array (caller must free with ml_free) */
    int32_t        score_count;  /** Number of scores returned */
    ml_ProfileData profile_data; /** Profiling data for the reranking operation */
} ml_RerankerRerankOutput;

/**
 * @brief Rerank documents against a query
 *
 * @param handle[in]: Reranker handle
 * @param input[in]: Input parameters for reranking operation
 * @param output[out]: Output data containing the ranking scores
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_reranker_rerank(
    ml_Reranker* handle, const ml_RerankerRerankInput* input, ml_RerankerRerankOutput* output);

/* ========================================================================== */
/*                              IMAGE GENERATION                               */
/* ========================================================================== */

/* ====================  Configuration Structures  =========================== */

/** Image generation sampling parameters */
typedef struct {
    const char* method;         /* Sampling method: "ddim", "ddpm", etc. */
    int32_t     steps;          /* Number of denoising steps */
    float       guidance_scale; /* Classifier-free guidance scale */
    float       eta;            /* DDIM eta parameter */
    int32_t     seed;           /* Random seed (-1 for random) */
} ml_ImageSamplerConfig;

/** Diffusion scheduler configuration */
typedef struct {
    const char* type;                /* Scheduler type: "ddim", etc. */
    int32_t     num_train_timesteps; /* Training timesteps */
    int32_t     steps_offset;        /* An offset added to the inference steps */
    float       beta_start;          /* Beta schedule start */
    float       beta_end;            /* Beta schedule end */
    const char* beta_schedule;       /* Beta schedule: "scaled_linear" */
    const char* prediction_type;     /* Prediction type: "epsilon", "v_prediction" */
    const char* timestep_type;       /* Timestep type: "discrete", "continuous" */
    const char* timestep_spacing;    /* Timestep spacing: "linspace", "leading", "trailing" */
    const char* interpolation_type;  /* Interpolation type: "linear", "exponential" */
    ml_Path     config_path;         /* Optional config file path */
} ml_SchedulerConfig;

/** Image generation configuration */
typedef struct {
    const char**          prompts;               /* Required positive prompts */
    int32_t               prompt_count;          /* Number of positive prompts */
    const char**          negative_prompts;      /* Optional negative prompts */
    int32_t               negative_prompt_count; /* Number of negative prompts */
    int32_t               height;                /* Output image height */
    int32_t               width;                 /* Output image width */
    ml_ImageSamplerConfig sampler_config;        /* Sampling parameters */
    ml_SchedulerConfig    scheduler_config;      /* Scheduler configuration */
    float                 strength;              /* Denoising strength for img2img */
} ml_ImageGenerationConfig;

typedef struct ml_ImageGen ml_ImageGen; /* Opaque image generator handle */

/* ====================  Lifecycle Management  ============================== */

/** Input structure for creating an image generator */
typedef struct {
    ml_Path     model_path;            /** Path to the model file */
    ml_Path     scheduler_config_path; /** Path to the scheduler config file */
    ml_PluginId plugin_id;             /** Plugin to use for the model */
    const char* device_id;             /** Device to use for the model, NULL for default device */
} ml_ImageGenCreateInput;

/**
 * @brief Create and initialize an image generator instance
 *
 * @param input[in]: Input parameters for the image generator creation
 * @param out_handle[out]: Pointer to the image generator handle. Must be freed with ml_imagegen_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_imagegen_create(const ml_ImageGenCreateInput* input, ml_ImageGen** out_handle);

/**
 * @brief Destroy image generator instance and free associated resources
 *
 * @param handle[in]: The image generator handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_imagegen_destroy(ml_ImageGen* handle);

/* ====================  Image Generation  ================================== */

/** Input structure for text-to-image generation */
typedef struct {
    const char*                     prompt_utf8; /** Text prompt in UTF-8 encoding */
    const ml_ImageGenerationConfig* config;      /** Image generation configuration */
    ml_Path                         output_path; /** Optional output file path (NULL for auto-generated) */
} ml_ImageGenTxt2ImgInput;

/** Input structure for image-to-image generation */
typedef struct {
    ml_Path                         init_image_path; /** Path to initial image file for img2img */
    const char*                     prompt_utf8;     /** Text prompt in UTF-8 encoding */
    const ml_ImageGenerationConfig* config;          /** Image generation configuration */
    ml_Path                         output_path;     /** Optional output file path (NULL for auto-generated) */
} ml_ImageGenImg2ImgInput;

/** Output structure for image generation */
typedef struct {
    ml_Path output_image_path; /** Path where the generated image will be saved (caller must free with ml_free) */
} ml_ImageGenOutput;

/**
 * @brief Generate image from text prompt and save to filesystem
 *
 * @param handle[in]: Image generator handle
 * @param input[in]: Input parameters for text-to-image generation
 * @param output[out]: Output data containing the path where the generated image is saved
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_imagegen_txt2img(
    ml_ImageGen* handle, const ml_ImageGenTxt2ImgInput* input, ml_ImageGenOutput* output);

/**
 * @brief Generate image from initial image file and prompt, save to filesystem
 *
 * @param handle[in]: Image generator handle
 * @param input[in]: Input parameters for image-to-image generation (includes initial image path)
 * @param output[out]: Output data containing the path where the generated image is saved
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_imagegen_img2img(
    ml_ImageGen* handle, const ml_ImageGenImg2ImgInput* input, ml_ImageGenOutput* output);

/* ========================================================================== */
/*                              SPEECH RECOGNITION (ASR)                       */
/* ========================================================================== */

/* ====================  Configuration Structures  =========================== */

/** ASR processing configuration */
typedef struct {
    const char* timestamps; /* Timestamp mode: "none", "segment", "word" */
    int32_t     beam_size;  /* Beam search size */
    bool        stream;     /* Enable streaming mode */
} ml_ASRConfig;

/** ASR transcription result */
typedef struct {
    char*   transcript;        /* Transcribed text (UTF-8, caller must free with ml_free) */
    float*  confidence_scores; /* Confidence scores for each unit (caller must free with ml_free) */
    int32_t confidence_count;  /* Number of confidence scores */
    float*  timestamps;        /* Timestamp pairs: [start, end] for each unit (caller must free with ml_free) */
    int32_t timestamp_count;   /* Number of timestamp pairs */
} ml_ASRResult;

typedef struct ml_ASR ml_ASR; /* Opaque ASR handle */

/* ====================  Lifecycle Management  ============================== */

/** Input structure for creating an ASR instance */
typedef struct {
    ml_Path     model_path;         /** Path to the model file */
    ml_Path     tokenizer_path;     /** Path to the tokenizer file (may be NULL) */
    const char* language;           /** Language code (ISO 639-1 or NULL) */
    ml_PluginId plugin_id;          /** Plugin to use for the model */
    const char* device_id;          /** Device to use for the model, NULL for default device */

    // QNN library paths (for QNN backend registration and loading)
    ml_Path encoder_model_path;
    ml_Path encoder_config_file_path;
    ml_Path decoder_model_path;
    ml_Path decoder_config_file_path;
    ml_Path jointer_model_path;
    ml_Path jointer_config_file_path;
    ml_Path system_library_path;
    ml_Path backend_library_path;
    ml_Path extension_library_path;
    ml_Path embed_weight_path;
    ml_Path pos_emb_path;
    ml_Path vocab_path;
    ml_Path audio_path;
    bool verbose;
} ml_AsrCreateInput;

/**
 * @brief Create and initialize an ASR instance with language support
 *
 * @param input[in]: Input parameters for the ASR creation
 * @param out_handle[out]: Pointer to the ASR handle. Must be freed with ml_asr_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_asr_create(const ml_AsrCreateInput* input, ml_ASR** out_handle);

/**
 * @brief Destroy ASR instance and free associated resources
 *
 * @param handle[in]: The ASR handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_asr_destroy(ml_ASR* handle);

/* ====================  Transcription  ===================================== */

/** Input structure for ASR transcription */
typedef struct {
    ml_Path             audio_path; /** Path to audio file */
    const char*         language;   /** Language code (ISO 639-1 or NULL for auto-detect) */
    const ml_ASRConfig* config;     /** ASR configuration (optional, can be nullptr) */
} ml_AsrTranscribeInput;

/** Output structure for ASR transcription */
typedef struct {
    ml_ASRResult   result;       /** Transcription result (caller must free with ml_free for text fields) */
    ml_ProfileData profile_data; /** Profiling data for the transcription operation */
} ml_AsrTranscribeOutput;

/**
 * @brief Transcribe audio file to text with specified language
 *
 * @param handle[in]: ASR handle
 * @param input[in]: Input parameters for transcription (includes audio file path and language)
 * @param output[out]: Output data containing the transcription result
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_asr_transcribe(ml_ASR* handle, const ml_AsrTranscribeInput* input, ml_AsrTranscribeOutput* output);

/* ====================  Language Management  ============================== */

/** Input structure for getting supported languages */
typedef struct {
    void* reserved; /** Reserved for future use, safe to set as NULL */
} ml_AsrListSupportedLanguagesInput;

/** Output structure for getting supported languages */
typedef struct {
    const char** language_codes; /** Array of supported language codes (caller must free with ml_free) */
    int32_t      language_count; /** Number of supported languages */
} ml_AsrListSupportedLanguagesOutput;

/**
 * @brief Get list of supported languages for ASR model
 *
 * @param handle[in]: ASR handle
 * @param input[in]: Input parameters for language list query
 * @param output[out]: Output data containing the supported languages array and count
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Thread-safe
 */
ML_API int32_t ml_asr_list_supported_languages(
    const ml_ASR* handle, const ml_AsrListSupportedLanguagesInput* input, ml_AsrListSupportedLanguagesOutput* output);

/* ========================================================================== */
/*                              TEXT-TO-SPEECH (TTS)                         */
/* ========================================================================== */

/* ====================  Configuration Structures  =========================== */

/** TTS synthesis configuration */
typedef struct {
    const char* voice;       /* Voice identifier */
    float       speed;       /* Speech speed (1.0 = normal) */
    int32_t     seed;        /* Random seed (-1 for random) */
    int32_t     sample_rate; /* Output sample rate in Hz */
} ml_TTSConfig;

/** TTS sampling parameters */
typedef struct {
    float temperature;  /* Sampling temperature */
    float noise_scale;  /* Noise scale for voice variation */
    float length_scale; /* Length scale for speech duration */
} ml_TTSSamplerConfig;

/** TTS synthesis result */
typedef struct {
    ml_Path audio_path;       /* Path where the synthesized audio is saved (caller must free with ml_free) */
    float   duration_seconds; /* Audio duration in seconds */
    int32_t sample_rate;      /* Audio sample rate in Hz */
    int32_t channels;         /* Number of audio channels (default: 1) */
    int32_t num_samples;      /* Number of audio samples */
} ml_TTSResult;

typedef struct ml_TTS ml_TTS; /* Opaque TTS handle */

/* ====================  Lifecycle Management  ============================== */

/** Input structure for creating a TTS instance */
typedef struct {
    ml_Path     model_path;   /** Path to the TTS model file */
    ml_Path     vocoder_path; /** Path to the vocoder file */
    ml_PluginId plugin_id;    /** Plugin to use for the model */
    const char* device_id;    /** Device to use for the model, NULL for default device */
} ml_TtsCreateInput;

/**
 * @brief Create and initialize a TTS instance with model and vocoder
 *
 * @param input[in]: Input parameters for the TTS creation
 * @param out_handle[out]: Pointer to the TTS handle. Must be freed with ml_tts_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_tts_create(const ml_TtsCreateInput* input, ml_TTS** out_handle);

/**
 * @brief Destroy TTS instance and free associated resources
 *
 * @param handle[in]: The TTS handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_tts_destroy(ml_TTS* handle);

/* ====================  Speech Synthesis  ================================== */

/** Input structure for TTS synthesis */
typedef struct {
    const char*         text_utf8;   /** Text to synthesize in UTF-8 encoding */
    const ml_TTSConfig* config;      /** TTS configuration (optional, can be nullptr) */
    ml_Path             output_path; /** Optional output file path (NULL for auto-generated) */
} ml_TtsSynthesizeInput;

/** Output structure for TTS synthesis */
typedef struct {
    ml_TTSResult   result;       /** Synthesis result with audio saved to filesystem */
    ml_ProfileData profile_data; /** Profiling data for the synthesis operation */
} ml_TtsSynthesizeOutput;

/**
 * @brief Synthesize speech from text and save to filesystem
 *
 * @param handle[in]: TTS handle
 * @param input[in]: Input parameters for speech synthesis
 * @param output[out]: Output data containing the path where synthesized audio is saved
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Not thread-safe
 */
ML_API int32_t ml_tts_synthesize(ml_TTS* handle, const ml_TtsSynthesizeInput* input, ml_TtsSynthesizeOutput* output);

/* ====================  Voice Management  ================================== */

/** Input structure for getting available voices */
typedef struct {
    void* reserved; /** Reserved for future use, safe to set as NULL */
} ml_TtsListAvailableVoicesInput;

/** Output structure for getting available voices */
typedef struct {
    const char** voice_ids;   /** Array of available voice identifiers (caller must free with ml_free) */
    int32_t      voice_count; /** Number of available voices */
} ml_TtsListAvailableVoicesOutput;

/**
 * @brief Get list of available voice identifiers
 *
 * @param handle[in]: TTS handle
 * @param input[in]: Input parameters for voice list query
 * @param output[out]: Output data containing the available voices array and count
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 *
 * @thread_safety: Thread-safe
 */
ML_API int32_t ml_tts_list_available_voices(
    const ml_TTS* handle, const ml_TtsListAvailableVoicesInput* input, ml_TtsListAvailableVoicesOutput* output);

/* ========================================================================== */
/*                              COMPUTER VISION (CV)                           */
/* ========================================================================== */

/* ====================  Generic CV Data Types  ============================= */
/** Generic bounding box structure */
typedef struct {
    float x;      /* X coordinate (normalized or pixel, depends on model) */
    float y;      /* Y coordinate (normalized or pixel, depends on model) */
    float width;  /* Width */
    float height; /* Height */
} ml_BoundingBox;

/** Generic detection/classification result */
typedef struct {
    ml_Path*       image_paths;   /* Output image paths (caller must free with ml_free) */
    int32_t        image_count;   /* Number of output images */
    int32_t        class_id;      /* Class ID (example: ConvNext) */
    float          confidence;    /* Confidence score [0.0-1.0] */
    ml_BoundingBox bbox;          /* Bounding box (example: YOLO) */
    const char*    text;          /* Text result (example: OCR) (caller must free with ml_free) */
    float*         embedding;     /* Feature embedding (example: CLIP embedding) (caller must free with ml_free) */
    int32_t        embedding_dim; /* Embedding dimension */
} ml_CVResult;

/** CV capabilities */
typedef enum {
    ML_CV_OCR            = 0, /* OCR */
    ML_CV_CLASSIFICATION = 1, /* Classification */
    ML_CV_SEGMENTATION   = 2, /* Segmentation */
    ML_CV_CUSTOM         = 3, /* Custom task */
} ml_CVCapabilities;

/** CV model preprocessing configuration */
typedef struct {
    ml_CVCapabilities capabilities; /* Capabilities */

    // MLX-OCR
    ml_Path det_model_path; /* detection model path */
    ml_Path rec_model_path; /* recognition model path */

    // QNN
    ml_Path model_path;             /* Model path */
    ml_Path system_library_path;    /* System library path */
    ml_Path backend_library_path;   /* Backend library path */
    ml_Path extension_library_path; /* Extension library path */
    ml_Path config_file_path;       /* Config file path */
    ml_Path char_dict_path;         /* Character dictionary path */
    ml_Path input_image_path;       /* Input image path */
} ml_CVModelConfig;

/* ====================  Generic CV Model  ================================== */

typedef struct ml_CV ml_CV; /* Opaque CV model handle */

typedef struct {
    ml_CVModelConfig config;
    ml_PluginId      plugin_id;
    const char*      device_id;
    const char*      license_id;     /** licence id for loading NPU models, must be provided upon the first use of the license key. null terminated string */
    const char*      license_key;    /** licence key for loading NPU models, null terminated string */
} ml_CVCreateInput;

/* ====================  Lifecycle Management  ============================== */

/**
 * @brief Create and initialize a CV model
 *
 * @param input[in]: Input parameters for the CV model creation
 * @param out_handle[out]: Pointer to the CV model handle. Must be freed with ml_cv_destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_cv_create(const ml_CVCreateInput* input, ml_CV** out_handle);

/**
 * @brief Destroy CV model instance and free associated resources
 *
 * @param handle[in]: The CV model handle to destroy.
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_cv_destroy(ml_CV* handle);

/* ====================  Generic Inference  ================================= */
/** Input structure for CV inference */
typedef struct {
    const char* input_image_path; /* Input image path */
} ml_CVInferInput;

/** Output structure for CV inference */
typedef struct {
    ml_CVResult* results;      /* Array of CV results (caller must free with ml_free) */
    int32_t      result_count; /* Number of CV results */
} ml_CVInferOutput;

/**
 * @brief Perform inference on a single image
 *
 * @param handle[in]: The CV model handle
 * @param input[in]: Input parameters for the inference
 * @param output[out]: Output data containing the inference results
 *
 * @return ml_ErrorCode: ML_SUCCESS on success, negative on failure.
 */
ML_API int32_t ml_cv_infer(const ml_CV* handle, const ml_CVInferInput* input, ml_CVInferOutput* output);

#ifdef __cplusplus
} /* extern "C" */
#endif