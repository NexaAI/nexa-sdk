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
extern "C"
{
#endif

    /* ========================================================================== */
    /*                              CORE TYPES & UTILITIES                         */
    /* ========================================================================== */

    /** Path string type - plain char* for file paths */
    typedef const char *ml_Path;

    /** Logging callback function type */
    typedef void (*ml_log_callback)(const char *);

    /** Token callback for streaming generation */
    typedef bool (*ml_llm_token_callback)(const char *token, void *user_data);

    /* ====================  Core Initialization  ================================ */

    /** Initialize the ML library - call before using any other functions */
    ML_API void ml_init(void);

    /** Clean up ML library resources - call when done */
    ML_API void ml_deinit(void);

    /** Set custom logging callback function */
    ML_API void ml_set_log(ml_log_callback callback);

    /** Log a message using the current logging mechanism */
    ML_API void ml_log(const char *message);

    /** Free memory allocated by ML library functions */
    ML_API void ml_free(void *ptr);

    /* ====================  Data Structures  ==================================== */

    /** Image data structure */
    typedef struct
    {
        float *data;      /* Raw pixel data: width × height × channels */
        int32_t width;    /* Image width in pixels */
        int32_t height;   /* Image height in pixels */
        int32_t channels; /* Color channels: 3=RGB, 4=RGBA */
    } ml_Image;

    /** Free image data structure and its pixel data */
    ML_API void ml_image_free(ml_Image *image);
    ML_API void ml_image_save(const ml_Image *image, const char *filename); /* Save image to file */

    /** Audio data structure */
    typedef struct
    {
        float *data;         /* Audio samples: num_samples × channels */
        int32_t sample_rate; /* Sample rate in Hz */
        int32_t channels;    /* Audio channels: 1=mono, 2=stereo */
        int32_t num_samples; /* Number of samples per channel */
    } ml_Audio;

    /** Free audio data structure and its sample data */
    ML_API void ml_audio_free(ml_Audio *audio);

    /** Video data structure */
    typedef struct
    {
        float *data;        /* Frame data: width × height × channels × num_frames */
        int32_t width;      /* Frame width in pixels */
        int32_t height;     /* Frame height in pixels */
        int32_t channels;   /* Color channels per frame */
        int32_t num_frames; /* Number of video frames */
    } ml_Video;

    /** Free video data structure and its frame data */
    ML_API void ml_video_free(ml_Video *video);

    /* ========================================================================== */
    /*                              LANGUAGE MODELS (LLM)                          */
    /* ========================================================================== */

    /** Text generation sampling parameters */
    typedef struct
    {
        float temperature;        /* Sampling temperature (0.0-2.0) */
        float top_p;              /* Nucleus sampling parameter (0.0-1.0) */
        int32_t top_k;            /* Top-k sampling parameter */
        float min_p;              /* Minimum probability for nucleus sampling */
        float repetition_penalty; /* Penalty for repeated tokens */
        float presence_penalty;   /* Penalty for token presence */
        float frequency_penalty;  /* Penalty for token frequency */
        int32_t seed;             /* Random seed (-1 for random) */
        ml_Path grammar_path;     /* Optional grammar file path */
    } ml_SamplerConfig;

    /** LLM / VLM generation configuration (IMPROVED: support multiple images and audios) */
    typedef struct
    {
        int32_t max_tokens;              /* Maximum tokens to generate */
        const char **stop;               /* Array of stop sequences */
        int32_t stop_count;              /* Number of stop sequences */
        int32_t n_past;                  /* Number of past tokens to consider */
        ml_SamplerConfig sampler_config; /* Advanced sampling config */
        // --- Improved multimodal support ---
        ml_Path *image_paths; /* Array of image paths for VLM (NULL if none) */
        int32_t image_count;  /* Number of images */
        ml_Path *audio_paths; /* Array of audio paths for VLM (NULL if none) */
        int32_t audio_count;  /* Number of audios */
    } ml_GenerationConfig;

    /** Chat message structure */
    typedef struct
    {
        const char *role;    /* Message role: "user", "assistant", "system" */
        const char *content; /* Message content in UTF-8 */
    } ml_ChatMessage;

    /* ====================  LLM Handle  ======================================== */
    typedef struct ml_LLM ml_LLM; /* Opaque LLM handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize an LLM instance from model files */
    ML_API ml_LLM *ml_llm_create(ml_Path model_path, ml_Path tokenizer_path, int32_t context_length, const char *device);

    /** Destroy LLM instance and free associated resources */
    ML_API void ml_llm_destroy(ml_LLM *handle);

    /** Reset LLM internal state (clear KV cache, reset sampling) */
    ML_API void ml_llm_reset(ml_LLM *handle);

    /* ====================  Tokenization  ====================================== */

    /** Encode UTF-8 text to token IDs. Returns token count, negative on error */
    ML_API int32_t ml_llm_encode(const ml_LLM *handle, const char *text_utf8, int32_t **out_tokens);

    /** Decode token IDs to UTF-8 text. Returns character count, negative on error */
    ML_API int32_t ml_llm_decode(const ml_LLM *handle, const int32_t *token_ids, int32_t length, char **out_text);

    /* ====================  KV-Cache Management  ============================== */

    /** Save current KV cache state to file. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_save_kv_cache(const ml_LLM *handle, ml_Path path);

    /** Load KV cache state from file. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_load_kv_cache(ml_LLM *handle, ml_Path path);

    /* ====================  LoRA Management  ================================== */

    /** Set active LoRA adapter by ID */
    ML_API void ml_llm_set_lora(ml_LLM *handle, int32_t lora_id);

    /** Add LoRA adapter from file. Returns LoRA ID on success, negative on error */
    ML_API int32_t ml_llm_add_lora(ml_LLM *handle, ml_Path lora_path);

    /** Remove LoRA adapter by ID */
    ML_API void ml_llm_remove_lora(ml_LLM *handle, int32_t lora_id);

    /** List all loaded LoRA adapter IDs. Returns count, negative on error */
    ML_API int32_t ml_llm_list_loras(const ml_LLM *handle, int32_t **out_lora_ids);

    /* ====================  Sampling Configuration  =========================== */

    /** Configure text generation sampling parameters */
    ML_API void ml_llm_set_sampler(ml_LLM *handle, const ml_SamplerConfig *config);

    /** Reset sampling parameters to defaults */
    ML_API void ml_llm_reset_sampler(ml_LLM *handle);

    /* ====================  Text Generation  ================================== */

    /** Generate text from prompt. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_generate(
        ml_LLM *handle, const char *prompt_utf8, const ml_GenerationConfig *config, char **out_text);

    /** Get chat template by name. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_get_chat_template(ml_LLM *handle, const char *template_name, const char **out_template);

    /** Apply chat template to messages. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_apply_chat_template(
        ml_LLM *handle, ml_ChatMessage *messages, int32_t message_count, char **out_text);

    /** Print detailed performance profile (sampler + context) */
    ML_API void ml_llm_print_profile(const ml_LLM *h);

    /* ====================  Embedding Generation  ============================= */

    /** Generate embeddings for input texts. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_embed(ml_LLM *handle, const char **texts_utf8, int32_t text_count, float **out_embeddings);

    /* ====================  Streaming Generation  ============================= */

    /** Generate text with streaming token callback. Returns 0 on success, negative on error */
    ML_API int32_t ml_llm_generate_stream(ml_LLM *handle, const char *prompt_utf8, const ml_GenerationConfig *config,
                                          ml_llm_token_callback on_token, void *user_data, char **out_full_text);

    /** Generate next token from input IDs. Returns token ID (>0), 0 for EOS, negative for error */
    ML_API int32_t ml_llm_generate_next_token(ml_LLM *handle, int32_t **input_ids, int32_t *input_length, int32_t *n_past);

    /* ========================================================================== */
    /*                              MULTIMODAL MODELS (VLM)                          */
    /* ========================================================================== */

    typedef struct ml_VLM ml_VLM; /* Opaque VLM handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize a VLM instance from model files */
    ML_API ml_VLM *ml_vlm_create(ml_Path model_path, ml_Path mmproj_path, int32_t context_length, const char *device);

    /** Destroy VLM instance and free associated resources */
    ML_API void ml_vlm_destroy(ml_VLM *handle);

    /** Reset VLM internal state (clear KV cache, reset sampling) */
    ML_API void ml_vlm_reset(ml_VLM *handle);

    /* ====================  Tokenization  ====================================== */

    /** Encode UTF-8 text to token IDs. Returns token count, negative on error */
    ML_API int32_t ml_vlm_encode(const ml_VLM *handle, const char *text_utf8, int32_t **out_tokens);

    /** Decode token IDs to UTF-8 text. Returns character count, negative on error */
    ML_API int32_t ml_vlm_decode(const ml_VLM *handle, const int32_t *token_ids, int32_t length, char **out_text);

    /* ====================  Sampling Configuration  =========================== */

    /** Configure text generation sampling parameters */
    ML_API void ml_vlm_set_sampler(ml_VLM *handle, const ml_SamplerConfig *config);

    /** Reset sampling parameters to defaults */
    ML_API void ml_vlm_reset_sampler(ml_VLM *handle);

    /* ====================  Text Generation  ================================== */

    /** Generate text from prompt with optional multimodal inputs. Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_generate(
        ml_VLM *handle, const char *prompt_utf8, const ml_GenerationConfig *config, char **out_text);

    /** Generate multimodal text with explicit image(s) and audio(s). Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_generate_multimodal(
        ml_VLM *handle, const char *prompt_utf8, ml_Path *image_paths, int32_t image_count,
        ml_Path *audio_paths, int32_t audio_count, const ml_GenerationConfig *config, char **out_text);

    /** Get chat template by name. Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_get_chat_template(ml_VLM *handle, const char *template_name, const char **out_template);

    /** Apply chat template to messages. Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_apply_chat_template(
        ml_VLM *handle, ml_ChatMessage *messages, int32_t message_count, char **out_text);

    /* ====================  Embedding Generation  ============================= */

    /** Generate embeddings for input texts. Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_embed(ml_VLM *handle, const char **texts_utf8, int32_t text_count, float **out_embeddings);

    /* ====================  Streaming Generation  ============================= */

    /** Generate text with streaming token callback and multimodal inputs. Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_generate_stream(
        ml_VLM *handle, const char *prompt_utf8, const ml_GenerationConfig *config,
        ml_llm_token_callback on_token, void *user_data, char **out_full_text);

    /** Generate multimodal text with streaming and explicit image(s) and audio(s). Returns 0 on success, negative on error */
    ML_API int32_t ml_vlm_generate_stream_multimodal(
        ml_VLM *handle, const char *prompt_utf8, ml_Path *image_paths, int32_t image_count,
        ml_Path *audio_paths, int32_t audio_count, const ml_GenerationConfig *config,
        ml_llm_token_callback on_token, void *user_data, char **out_full_text);

    /* ========================================================================== */
    /*                              EMBEDDING MODELS                               */
    /* ========================================================================== */

    /** Embedding generation configuration */
    typedef struct
    {
        int32_t batch_size;           /* Processing batch size */
        bool normalize;               /* Whether to normalize embeddings */
        const char *normalize_method; /* Normalization: "l2", "mean", "none" */
    } ml_EmbeddingConfig;

    typedef struct ml_Embedder ml_Embedder; /* Opaque embedder handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize an embedder instance from model files */
    ML_API ml_Embedder *ml_embedder_create(ml_Path model_path, ml_Path tokenizer_path, const char *device);

    /** Destroy embedder instance and free associated resources */
    ML_API void ml_embedder_destroy(ml_Embedder *handle);

    /* ====================  Embedding Generation  ============================= */

    /** Generate embeddings for input texts. Returns 0 on success, negative on error */
    ML_API int32_t ml_embedder_embed(
        ml_Embedder *handle, const char **texts, int32_t text_count, const ml_EmbeddingConfig *config, float **out);

    /* ====================  Model Information  ================================ */

    /** Get embedding dimension. Returns dimension size, negative on error */
    ML_API int32_t ml_embedder_embedding_dim(const ml_Embedder *handle);

    /* ====================  LoRA Management  ================================== */

    /** Set active LoRA adapter by ID */
    ML_API void ml_embedder_set_lora(ml_Embedder *handle, int32_t lora_id);

    /** Add LoRA adapter from file. Returns LoRA ID on success, negative on error */
    ML_API int32_t ml_embedder_add_lora(ml_Embedder *handle, ml_Path lora_path);

    /** Remove LoRA adapter by ID */
    ML_API void ml_embedder_remove_lora(ml_Embedder *handle, int32_t lora_id);

    /** List all loaded LoRA adapter IDs. Returns count, negative on error */
    ML_API int32_t ml_embedder_list_loras(const ml_Embedder *handle, int32_t **out);

    /* ========================================================================== */
    /*                              RERANKING MODELS                               */
    /* ========================================================================== */

    /** Reranking configuration */
    typedef struct
    {
        int32_t batch_size;           /* Processing batch size */
        bool normalize;               /* Whether to normalize scores */
        const char *normalize_method; /* Normalization: "softmax", "min-max", "none" */
    } ml_RerankConfig;

    typedef struct ml_Reranker ml_Reranker; /* Opaque reranker handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize a reranker instance from model files */
    ML_API ml_Reranker *ml_reranker_create(ml_Path model_path, ml_Path tokenizer_path, const char *device);

    /** Destroy reranker instance and free associated resources */
    ML_API void ml_reranker_destroy(ml_Reranker *handle);

    /* ====================  Reranking  ========================================= */

    /** Rerank documents against a query. Returns 0 on success, negative on error */
    ML_API int32_t ml_reranker_rerank(ml_Reranker *handle, const char *query, const char **documents,
                                      int32_t documents_count, const ml_RerankConfig *config, float **out);

    /* ========================================================================== */
    /*                              IMAGE GENERATION                               */
    /* ========================================================================== */

    /* ====================  Configuration Structures  =========================== */

    /** Image generation sampling parameters */
    typedef struct
    {
        const char *method;   /* Sampling method: "ddim", "ddpm", etc. */
        int32_t steps;        /* Number of denoising steps */
        float guidance_scale; /* Classifier-free guidance scale */
        float eta;            /* DDIM eta parameter */
        int32_t seed;         /* Random seed (-1 for random) */
    } ml_ImageSamplerConfig;

    /** Image generation configuration */
    typedef struct
    {
        const char **prompts;                 /* Required positive prompts */
        const char **negative_prompts;        /* Optional negative prompts */
        int32_t height;                       /* Output image height */
        int32_t width;                        /* Output image width */
        ml_ImageSamplerConfig sampler_config; /* Sampling parameters */
        int32_t lora_id;                      /* LoRA ID (-1 for none) */
        const ml_Image *init_image;           /* Initial image (NULL for txt2img) */
        float strength;                       /* Denoising strength for img2img */
    } ml_ImageGenerationConfig;

    /** Diffusion scheduler configuration */
    typedef struct
    {
        const char *type;               /* Scheduler type: "ddim", etc. */
        int32_t num_train_timesteps;    /* Training timesteps */
        float beta_start;               /* Beta schedule start */
        float beta_end;                 /* Beta schedule end */
        const char *beta_schedule;      /* Beta schedule: "scaled_linear" */
        const char *prediction_type;    /* Prediction type: "epsilon", "v_prediction" */
        const char *timestep_type;      /* Timestep type: "discrete", "continuous" */
        const char *timestep_spacing;   /* Timestep spacing: "linspace", "leading", "trailing" */
        const char *interpolation_type; /* Interpolation type: "linear", "exponential" */
        ml_Path config_path;            /* Optional config file path */
    } ml_SchedulerConfig;

    typedef struct ml_ImageGen ml_ImageGen; /* Opaque image generator handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize an image generator instance */
    ML_API ml_ImageGen *ml_imagegen_create(ml_Path model_path, ml_Path scheduler_config_path, const char *device);

    /** Destroy image generator instance and free associated resources */
    ML_API void ml_imagegen_destroy(ml_ImageGen *handle);

    /** Load model from path with optional extra configuration data */
    ML_API bool ml_imagegen_load_model(ml_ImageGen *handle, ml_Path model_path, const void *extra_data);

    /** Close and cleanup image generator resources */
    ML_API void ml_imagegen_close(ml_ImageGen *handle);

    /* ====================  Configuration  ===================================== */

    /** Configure diffusion scheduler parameters */
    ML_API void ml_imagegen_set_scheduler(ml_ImageGen *handle, const ml_SchedulerConfig *config);

    /** Configure image generation sampling parameters */
    ML_API void ml_imagegen_set_sampler(ml_ImageGen *handle, const ml_ImageSamplerConfig *config);

    /** Reset sampling parameters to defaults */
    ML_API void ml_imagegen_reset_sampler(ml_ImageGen *handle);

    /* ====================  Image Generation  ================================== */

    /** Generate image from text prompt */
    ML_API ml_Image ml_imagegen_txt2img(
        ml_ImageGen *handle, const char *prompt_utf8, const ml_ImageGenerationConfig *config);

    /** Generate image from initial image and prompt */
    ML_API ml_Image ml_imagegen_img2img(ml_ImageGen *handle, const ml_Image *init_image, const char *prompt_utf8,
                                        const ml_ImageGenerationConfig *config);

    /** Generate image using full configuration */
    ML_API ml_Image ml_imagegen_generate(ml_ImageGen *handle, const ml_ImageGenerationConfig *config);

    /* ====================  LoRA Management  ================================== */

    /** Set active LoRA adapter by ID */
    ML_API void ml_imagegen_set_lora(ml_ImageGen *handle, int32_t lora_id);

    /** Add LoRA adapter from file. Returns LoRA ID on success, negative on error */
    ML_API int32_t ml_imagegen_add_lora(ml_ImageGen *handle, ml_Path lora_path);

    /** Remove LoRA adapter by ID */
    ML_API void ml_imagegen_remove_lora(ml_ImageGen *handle, int32_t lora_id);

    /** List all loaded LoRA adapter IDs. Returns count, negative on error */
    ML_API int32_t ml_imagegen_list_loras(ml_ImageGen *handle, int32_t **out);

    /* ========================================================================== */
    /*                              COMPUTER VISION (CV)                           */
    /* ========================================================================== */

    /* ====================  OCR Configuration  ================================= */

    /** OCR pipeline configuration */
    typedef struct
    {
        ml_Path detector_model_path;   /* Text detection model path */
        ml_Path recognizer_model_path; /* Text recognition model path */
        const char *device;            /* Processing device ("cpu" default) */
    } ml_OCRPipelineConfig;

    /** OCR detection result */
    typedef struct
    {
        int32_t box[4]; /* Bounding box: [x_min, y_min, x_max, y_max] */
        char *text;     /* Detected text (UTF-8, caller must free) */
        float score;    /* Detection confidence score */
    } ml_OCRResult;

    /* ====================  Text Detection  ==================================== */
    typedef struct ml_TextDetector ml_TextDetector; /* Opaque detector handle */

    /** Create and initialize a text detector instance */
    ML_API ml_TextDetector *ml_textdetector_create(ml_Path model_path, const char *device);

    /** Destroy text detector instance and free associated resources */
    ML_API void ml_textdetector_destroy(ml_TextDetector *handle);

    /** Load text detection model from path */
    ML_API bool ml_textdetector_load_model(ml_TextDetector *handle, ml_Path model_path, const char *device);

    /** Close and cleanup text detector resources */
    ML_API void ml_textdetector_close(ml_TextDetector *handle);

    /** Detect text regions in a single image. Returns bounding boxes */
    ML_API int32_t *ml_textdetector_infer(const ml_TextDetector *handle, const ml_Image *image, int32_t *out_box_count);

    /** Detect text regions in multiple images. Returns bounding boxes for each image */
    ML_API int32_t *ml_textdetector_infer_batch(
        const ml_TextDetector *handle, const ml_Image *images, int32_t image_count, int32_t **out_counts);

    /* ====================  Text Recognition  ================================== */
    typedef struct ml_TextRecognizer ml_TextRecognizer; /* Opaque recognizer handle */

    /** Create and initialize a text recognizer instance */
    ML_API ml_TextRecognizer *ml_textrecognizer_create(ml_Path model_path, const char *device);

    /** Destroy text recognizer instance and free associated resources */
    ML_API void ml_textrecognizer_destroy(ml_TextRecognizer *handle);

    /** Load text recognition model from path */
    ML_API bool ml_textrecognizer_load_model(ml_TextRecognizer *handle, ml_Path model_path, const char *device);

    /** Close and cleanup text recognizer resources */
    ML_API void ml_textrecognizer_close(ml_TextRecognizer *handle);

    /** Recognize text in a single image region. Returns recognized text */
    ML_API char *ml_textrecognizer_infer(const ml_TextRecognizer *handle, const ml_Image *image);

    /** Recognize text in multiple image regions. Returns array of recognized texts */
    ML_API char **ml_textrecognizer_infer_batch(
        const ml_TextRecognizer *handle, const ml_Image *images, int32_t image_count);

    /* ====================  OCR Pipeline  ====================================== */
    typedef struct ml_OCR ml_OCR; /* Opaque OCR handle */

    /** Create and initialize a complete OCR pipeline */
    ML_API ml_OCR *ml_ocr_create(const ml_OCRPipelineConfig *config);

    /** Destroy OCR pipeline instance and free associated resources */
    ML_API void ml_ocr_destroy(ml_OCR *handle);

    /** Load OCR models from configuration */
    ML_API bool ml_ocr_load_model(ml_OCR *handle, const ml_OCRPipelineConfig *config);

    /** Close and cleanup OCR pipeline resources */
    ML_API void ml_ocr_close(ml_OCR *handle);

    /** Perform OCR on a single image. Returns detection and recognition results */
    ML_API ml_OCRResult *ml_ocr_infer(const ml_OCR *handle, const ml_Image *image, int32_t *out_count);

    /** Perform OCR on multiple images. Returns arrays of results for each image */
    ML_API ml_OCRResult **ml_ocr_infer_batch(
        const ml_OCR *handle, const ml_Image *images, int32_t image_count, int32_t **out_counts);

    /** Free OCR result structures and associated text data */
    ML_API void ml_ocr_free_results(ml_OCRResult *results, int32_t count);

    /* ========================================================================== */
    /*                              SPEECH RECOGNITION (ASR)                       */
    /* ========================================================================== */

    /* ====================  Configuration Structures  =========================== */

    /** ASR processing configuration */
    typedef struct
    {
        const char *timestamps; /* Timestamp mode: "none", "segment", "word" */
        int32_t beam_size;      /* Beam search size */
        bool stream;            /* Enable streaming mode */
    } ml_ASRConfig;

    /** ASR transcription result */
    typedef struct
    {
        char *transcript;         /* Transcribed text (UTF-8, caller must free) */
        float *confidence_scores; /* Confidence scores for each unit */
        int32_t confidence_count; /* Number of confidence scores */
        float *timestamps;        /* Timestamp pairs: [start, end] for each unit */
        int32_t timestamp_count;  /* Number of timestamp pairs */
    } ml_ASRResult;

    typedef struct ml_ASR ml_ASR; /* Opaque ASR handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize an ASR instance with language support */
    ML_API ml_ASR *ml_asr_create(ml_Path model_path, ml_Path tokenizer_path, /* tokenizer may be NULL */
                                 const char *language,                       /* ISO 639-1 or NULL */
                                 const char *device);

    /** Destroy ASR instance and free associated resources */
    ML_API void ml_asr_destroy(ml_ASR *handle);

    /** Load ASR model from path with optional extra configuration data */
    ML_API bool ml_asr_load_model(ml_ASR *handle, ml_Path model_path, const void *extra_data);

    /** Close and cleanup ASR resources */
    ML_API void ml_asr_close(ml_ASR *handle);

    /* ====================  Transcription  ===================================== */

    /** Transcribe audio samples to text */
    ML_API ml_ASRResult ml_asr_transcribe(
        ml_ASR *handle, const float *audio, int32_t num_samples, int32_t sample_rate, const ml_ASRConfig *config);

    /** Transcribe multiple audio samples in batch */
    ML_API ml_ASRResult *ml_asr_transcribe_batch(ml_ASR *handle, const float **audios, const int32_t *num_samples_array,
                                                 int32_t batch_size, int32_t sample_rate, const ml_ASRConfig *config);

    /** Transcribe audio chunk for streaming recognition */
    ML_API ml_ASRResult ml_asr_transcribe_step(
        ml_ASR *handle, const float *audio_chunk, int32_t num_samples, int32_t step, const ml_ASRConfig *config);

    /* ====================  Result Management  ================================ */

    /** Print ASR result to standard output for debugging */
    ML_API void ml_asr_print_result(const ml_ASRResult *result);

    /** Free ASR result structure and associated data */
    ML_API void ml_asr_free_result(ml_ASRResult *result);

    /* ====================  Language Management  ============================== */

    /** Get list of supported languages. Returns array of language codes */
    ML_API const char **ml_asr_list_supported_languages(const ml_ASR *handle, int32_t *out_count);

    /** Set recognition language by ISO 639-1 code */
    ML_API void ml_asr_set_language(ml_ASR *handle, const char *language);

    /* ========================================================================== */
    /*                              TEXT-TO-SPEECH (TTS)                         */
    /* ========================================================================== */

    /* ====================  Configuration Structures  =========================== */

    /** TTS synthesis configuration */
    typedef struct
    {
        const char *voice;   /* Voice identifier */
        float speed;         /* Speech speed (1.0 = normal) */
        int32_t seed;        /* Random seed (-1 for random) */
        int32_t sample_rate; /* Output sample rate in Hz */
    } ml_TTSConfig;

    /** TTS sampling parameters */
    typedef struct
    {
        float temperature;  /* Sampling temperature */
        float noise_scale;  /* Noise scale for voice variation */
        float length_scale; /* Length scale for speech duration */
    } ml_TTSSamplerConfig;

    /** TTS synthesis result */
    typedef struct
    {
        float *audio;           /* Audio samples: num_samples × channels */
        float duration_seconds; /* Audio duration in seconds */
        int32_t sample_rate;    /* Audio sample rate in Hz */
        int32_t channels;       /* Number of audio channels (default: 1) */
        int32_t num_samples;    /* Number of audio samples */
    } ml_TTSResult;

    typedef struct ml_TTS ml_TTS; /* Opaque TTS handle */

    /* ====================  Lifecycle Management  ============================== */

    /** Create and initialize a TTS instance with model and vocoder */
    ML_API ml_TTS *ml_tts_create(ml_Path model_path, ml_Path vocoder_path, const char *device);

    /** Destroy TTS instance and free associated resources */
    ML_API void ml_tts_destroy(ml_TTS *handle);

    /** Load TTS model from path with optional extra configuration data */
    ML_API bool ml_tts_load_model(ml_TTS *handle, ml_Path model_path, const void *extra_data);

    /** Close and cleanup TTS resources */
    ML_API void ml_tts_close(ml_TTS *handle);

    /* ====================  Configuration  ===================================== */

    /** Configure TTS sampling parameters */
    ML_API void ml_tts_set_sampler(ml_TTS *handle, const ml_TTSSamplerConfig *config);

    /** Reset sampling parameters to defaults */
    ML_API void ml_tts_reset_sampler(ml_TTS *handle);

    /* ====================  Speech Synthesis  ================================== */

    /** Synthesize speech from text */
    ML_API ml_TTSResult ml_tts_synthesize(ml_TTS *handle, const char *text_utf8, const ml_TTSConfig *config);

    /** Synthesize speech from multiple texts in batch */
    ML_API ml_TTSResult *ml_tts_synthesize_batch(
        ml_TTS *handle, const char **texts, int32_t text_count, const ml_TTSConfig *config);

    /** Synthesize speech chunk for streaming synthesis */
    ML_API ml_TTSResult ml_tts_synthesize_step(
        ml_TTS *handle, const char *text_utf8, int32_t step, const ml_TTSConfig *config);

    /* ====================  Cache Management  ================================== */

    /** Save TTS cache state to file */
    ML_API void ml_tts_save_cache(ml_TTS *handle, ml_Path path);

    /** Load TTS cache state from file */
    ML_API void ml_tts_load_cache(ml_TTS *handle, ml_Path path);

    /* ====================  Voice Management  ================================== */

    /** Get list of available voice identifiers */
    ML_API const char **ml_tts_list_available_voices(const ml_TTS *handle, int32_t *out_count);

    /* ====================  Result Management  ================================ */

    /** Print TTS result information to standard output for debugging */
    ML_API void ml_tts_print_result(const ml_TTSResult *result);

    /** Free TTS result structure and associated audio data */
    ML_API void ml_tts_free_result(ml_TTSResult *result);

#ifdef __cplusplus
} /* extern "C" */
#endif