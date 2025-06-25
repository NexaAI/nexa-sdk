#pragma once

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ========================================================================== */
/*                              CORE TYPES & UTILITIES                         */
/* ========================================================================== */

/** Path string type - plain char* for file paths */
typedef const char* ml_Path;

/** Logging callback function type */
typedef void (*ml_log_callback)(const char*);

/** Token callback for streaming generation */
typedef bool (*ml_llm_token_callback)(const char* token, void* user_data);

/* ====================  Core Initialization  ================================ */
void ml_init(void);                        /* Initialization */
void ml_deinit(void);                      /* Cleanup */
void ml_set_log(ml_log_callback callback); /* Set logging callback */
void ml_log(const char* message);          /* Log a message */
void ml_free(void* ptr);                   /* Free allocated memory */

/* ====================  Data Structures  ==================================== */

/** Image data structure */
typedef struct {
    float*  data;     /* Raw pixel data: width × height × channels */
    int32_t width;    /* Image width in pixels */
    int32_t height;   /* Image height in pixels */
    int32_t channels; /* Color channels: 3=RGB, 4=RGBA */
} ml_Image;

/** Audio data structure */
typedef struct {
    float*  data;        /* Audio samples: num_samples × channels */
    int32_t sample_rate; /* Sample rate in Hz */
    int32_t channels;    /* Audio channels: 1=mono, 2=stereo */
    int32_t num_samples; /* Number of samples per channel */
} ml_Audio;

/** Video data structure */
typedef struct {
    float*  data;       /* Frame data: width × height × channels × num_frames */
    int32_t width;      /* Frame width in pixels */
    int32_t height;     /* Frame height in pixels */
    int32_t channels;   /* Color channels per frame */
    int32_t num_frames; /* Number of video frames */
} ml_Video;

/* ========================================================================== */
/*                              LANGUAGE MODELS (LLM)                          */
/* ========================================================================== */

/** Text generation sampling parameters */
typedef struct {
    float   temperature;        /* Sampling temperature (0.0-2.0) */
    float   top_p;              /* Nucleus sampling parameter (0.0-1.0) */
    int32_t top_k;              /* Top-k sampling parameter */
    float   repetition_penalty; /* Penalty for repeated tokens */
    float   presence_penalty;   /* Penalty for token presence */
    float   frequency_penalty;  /* Penalty for token frequency */
    int32_t seed;               /* Random seed (-1 for random) */
    ml_Path grammar_path;       /* Optional grammar file path */
} ml_SamplerConfig;

/** Text generation configuration */
typedef struct {
    int32_t          max_tokens;         /* Maximum tokens to generate */
    float            temperature;        /* Sampling temperature */
    float            top_p;              /* Nucleus sampling parameter */
    int32_t          top_k;              /* Top-k sampling parameter */
    float            repetition_penalty; /* Repetition penalty */
    const char**     stop;               /* Array of stop sequences */
    int32_t          stop_count;         /* Number of stop sequences */
    ml_SamplerConfig sampler_config;     /* Advanced sampling config */
} ml_GenerationConfig;

/** Chat message structure */
typedef struct {
    const char* role;    /* Message role: "user", "assistant", "system" */
    const char* content; /* Message content in UTF-8 */
} ml_ChatMessage;

/* ====================  LLM Handle  ======================================== */
typedef struct ml_LLM ml_LLM; /* Opaque LLM handle */

/* ====================  Lifecycle Management  ============================== */
ml_LLM* ml_llm_create(ml_Path model_path, ml_Path tokenizer_path, int32_t context_length, const char* device);
void    ml_llm_destroy(ml_LLM* handle);
void    ml_llm_reset(ml_LLM* handle); /* Reset internal state */

/* ====================  Tokenization  ====================================== */
int32_t ml_llm_encode(const ml_LLM* handle, const char* text_utf8, int32_t** out_tokens);
int32_t ml_llm_decode(const ml_LLM* handle, const int32_t* token_ids, int32_t length, char** out_text);

/* ====================  KV-Cache Management  ============================== */
int32_t ml_llm_save_kv_cache(const ml_LLM* handle, ml_Path path);
int32_t ml_llm_load_kv_cache(ml_LLM* handle, ml_Path path);

/* ====================  LoRA Management  ================================== */
void    ml_llm_set_lora(ml_LLM* handle, int32_t lora_id);
int32_t ml_llm_add_lora(ml_LLM* handle, ml_Path lora_path);
void    ml_llm_remove_lora(ml_LLM* handle, int32_t lora_id);
int32_t ml_llm_list_loras(const ml_LLM* handle, int32_t** out_lora_ids);

/* ====================  Sampling Configuration  =========================== */
void ml_llm_set_sampler(ml_LLM* handle, const ml_SamplerConfig* config);
void ml_llm_reset_sampler(ml_LLM* handle);

/* ====================  Text Generation  ================================== */
int32_t ml_llm_generate(ml_LLM* handle, const char* prompt_utf8, const ml_GenerationConfig* config, char** out_text);
int32_t ml_llm_get_chat_template(ml_LLM* handle, const char* template_name, const char** out_template);
int32_t ml_llm_apply_chat_template(ml_LLM* handle, ml_ChatMessage* messages, int32_t message_count, char** out_text);

/* ====================  Embedding Generation  ============================= */
int32_t ml_llm_embed(ml_LLM* handle, const char** texts_utf8, int32_t text_count, float** out_embeddings);

/* ====================  Streaming Generation  ============================= */
int32_t ml_llm_generate_stream(ml_LLM* handle, const char* prompt_utf8, const ml_GenerationConfig* config,
    ml_llm_token_callback on_token, void* user_data, char** out_full_text);

/* ========================================================================== */
/*                              MULTIMODAL MODELS (VLM)                          */
/* ========================================================================== */

typedef struct ml_VLM ml_VLM;  /* Opaque VLM handle */

/* ====================  Lifecycle Management  ============================== */
ml_VLM* ml_vlm_create(ml_Path model_path, ml_Path mmproj_path, int32_t context_length, const char* device);
void    ml_vlm_destroy(ml_VLM* handle);
void    ml_vlm_reset(ml_VLM* handle);  /* Reset internal state */

/* ====================  Tokenization  ====================================== */
int32_t ml_vlm_encode(const ml_VLM* handle, const char* text_utf8, int32_t** out_tokens);
int32_t ml_vlm_decode(const ml_VLM* handle, const int32_t* token_ids, 
                      int32_t length, char** out_text);

/* ====================  Sampling Configuration  =========================== */
void ml_vlm_set_sampler(ml_VLM* handle, const ml_SamplerConfig* config);
void ml_vlm_reset_sampler(ml_VLM* handle);

/* ====================  Text Generation  ================================== */
int32_t ml_vlm_generate(ml_VLM* handle, const char* prompt_utf8, 
                        const ml_GenerationConfig* config, char** out_text);
int32_t ml_vlm_get_chat_template(ml_VLM* handle, const char* template_name, 
                                 const char** out_template);
int32_t ml_vlm_apply_chat_template(ml_VLM* handle, ml_ChatMessage* messages, 
                                   int32_t message_count, char** out_text);

/* ====================  Embedding Generation  ============================= */
int32_t ml_vlm_embed(ml_VLM* handle, const char** texts_utf8, 
                     int32_t text_count, float** out_embeddings);

/* ====================  Streaming Generation  ============================= */
int32_t ml_vlm_generate_stream(ml_VLM* handle, const char* prompt_utf8,
                               const ml_GenerationConfig* config,
                               ml_llm_token_callback on_token, void* user_data,
                               char** out_full_text);

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
ml_Embedder* ml_embedder_create(ml_Path model_path, ml_Path tokenizer_path, const char* device);
void         ml_embedder_destroy(ml_Embedder* handle);

/* ====================  Embedding Generation  ============================= */
int32_t ml_embedder_embed(
    ml_Embedder* handle, const char** texts, int32_t text_count, const ml_EmbeddingConfig* config, float** out);

/* ====================  Model Information  ================================ */
int32_t ml_embedder_embedding_dim(const ml_Embedder* handle);

/* ====================  LoRA Management  ================================== */
void    ml_embedder_set_lora(ml_Embedder* handle, int32_t lora_id);
int32_t ml_embedder_add_lora(ml_Embedder* handle, ml_Path lora_path);
void    ml_embedder_remove_lora(ml_Embedder* handle, int32_t lora_id);
int32_t ml_embedder_list_loras(const ml_Embedder* handle, int32_t** out);

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
ml_Reranker* ml_reranker_create(ml_Path model_path, ml_Path tokenizer_path, const char* device);
void         ml_reranker_destroy(ml_Reranker* handle);

/* ====================  Reranking  ========================================= */
int32_t ml_reranker_rerank(ml_Reranker* handle, const char* query, const char** documents, int32_t documents_count,
    const ml_RerankConfig* config, float** out);

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

/** Image generation configuration */
typedef struct {
    const char*           prompt;          /* Required positive prompt */
    const char*           negative_prompt; /* Optional negative prompt */
    int32_t               height;          /* Output image height */
    int32_t               width;           /* Output image width */
    ml_ImageSamplerConfig sampler_config;  /* Sampling parameters */
    int32_t               lora_id;         /* LoRA ID (-1 for none) */
    const ml_Image*       init_image;      /* Initial image (NULL for txt2img) */
    float                 strength;        /* Denoising strength for img2img */
} ml_ImageGenerationConfig;

/** Diffusion scheduler configuration */
typedef struct {
    const char* type;                /* Scheduler type: "ddim", etc. */
    int32_t     num_train_timesteps; /* Training timesteps */
    float       beta_start;          /* Beta schedule start */
    float       beta_end;            /* Beta schedule end */
    const char* beta_schedule;       /* Beta schedule: "scaled_linear" */
    ml_Path     config_path;         /* Optional config file path */
} ml_SchedulerConfig;

typedef struct ml_ImageGen ml_ImageGen; /* Opaque image generator handle */

/* ====================  Lifecycle Management  ============================== */
ml_ImageGen* ml_imagegen_create(ml_Path model_path, ml_Path scheduler_config_path, const char* device);
void         ml_imagegen_destroy(ml_ImageGen* handle);
bool         ml_imagegen_load_model(ml_ImageGen* handle, ml_Path model_path, const void* extra_data);
void         ml_imagegen_close(ml_ImageGen* handle);

/* ====================  Configuration  ===================================== */
void ml_imagegen_set_scheduler(ml_ImageGen* handle, const ml_SchedulerConfig* config);
void ml_imagegen_set_sampler(ml_ImageGen* handle, const ml_ImageSamplerConfig* config);
void ml_imagegen_reset_sampler(ml_ImageGen* handle);

/* ====================  Image Generation  ================================== */
ml_Image ml_imagegen_txt2img(ml_ImageGen* handle, const char* prompt_utf8, const ml_ImageGenerationConfig* config);
ml_Image ml_imagegen_img2img(
    ml_ImageGen* handle, const ml_Image* init_image, const char* prompt_utf8, const ml_ImageGenerationConfig* config);
ml_Image ml_imagegen_generate(ml_ImageGen* handle, const ml_ImageGenerationConfig* config);

/* ====================  LoRA Management  ================================== */
void    ml_imagegen_set_lora(ml_ImageGen* handle, int32_t lora_id);
int32_t ml_imagegen_add_lora(ml_ImageGen* handle, ml_Path lora_path);
void    ml_imagegen_remove_lora(ml_ImageGen* handle, int32_t lora_id);
int32_t ml_imagegen_list_loras(ml_ImageGen* handle, int32_t** out);

/* ========================================================================== */
/*                              COMPUTER VISION (CV)                           */
/* ========================================================================== */

/* ====================  OCR Configuration  ================================= */

/** OCR pipeline configuration */
typedef struct {
    ml_Path     detector_model_path;   /* Text detection model path */
    ml_Path     recognizer_model_path; /* Text recognition model path */
    const char* device;                /* Processing device ("cpu" default) */
} ml_OCRPipelineConfig;

/** OCR detection result */
typedef struct {
    int32_t box[4]; /* Bounding box: [x_min, y_min, x_max, y_max] */
    char*   text;   /* Detected text (UTF-8, caller must free) */
    float   score;  /* Detection confidence score */
} ml_OCRResult;

/* ====================  Text Detection  ==================================== */
typedef struct ml_TextDetector ml_TextDetector; /* Opaque detector handle */

ml_TextDetector* ml_textdetector_create(ml_Path model_path, const char* device);
void             ml_textdetector_destroy(ml_TextDetector* handle);
bool             ml_textdetector_load_model(ml_TextDetector* handle, ml_Path model_path, const char* device);
void             ml_textdetector_close(ml_TextDetector* handle);

int32_t* ml_textdetector_infer(const ml_TextDetector* handle, const ml_Image* image, int32_t* out_box_count);
int32_t* ml_textdetector_infer_batch(
    const ml_TextDetector* handle, const ml_Image* images, int32_t image_count, int32_t** out_counts);

/* ====================  Text Recognition  ================================== */
typedef struct ml_TextRecognizer ml_TextRecognizer; /* Opaque recognizer handle */

ml_TextRecognizer* ml_textrecognizer_create(ml_Path model_path, const char* device);
void               ml_textrecognizer_destroy(ml_TextRecognizer* handle);
bool               ml_textrecognizer_load_model(ml_TextRecognizer* handle, ml_Path model_path, const char* device);
void               ml_textrecognizer_close(ml_TextRecognizer* handle);

char*  ml_textrecognizer_infer(const ml_TextRecognizer* handle, const ml_Image* image);
char** ml_textrecognizer_infer_batch(const ml_TextRecognizer* handle, const ml_Image* images, int32_t image_count);

/* ====================  OCR Pipeline  ====================================== */
typedef struct ml_OCR ml_OCR; /* Opaque OCR handle */

ml_OCR* ml_ocr_create(const ml_OCRPipelineConfig* config);
void    ml_ocr_destroy(ml_OCR* handle);
bool    ml_ocr_load_model(ml_OCR* handle, const ml_OCRPipelineConfig* config);
void    ml_ocr_close(ml_OCR* handle);

ml_OCRResult*  ml_ocr_infer(const ml_OCR* handle, const ml_Image* image, int32_t* out_count);
ml_OCRResult** ml_ocr_infer_batch(
    const ml_OCR* handle, const ml_Image* images, int32_t image_count, int32_t** out_counts);

void ml_ocr_free_results(ml_OCRResult* results, int32_t count);

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
    char*   transcript;        /* Transcribed text (UTF-8, caller must free) */
    float*  confidence_scores; /* Confidence scores for each unit */
    int32_t confidence_count;  /* Number of confidence scores */
    float*  timestamps;        /* Timestamp pairs: [start, end] for each unit */
    int32_t timestamp_count;   /* Number of timestamp pairs */
} ml_ASRResult;

typedef struct ml_ASR ml_ASR; /* Opaque ASR handle */

/* ====================  Lifecycle Management  ============================== */
ml_ASR* ml_asr_create(ml_Path model_path, ml_Path tokenizer_path, /* tokenizer may be NULL */
    const char* language,                                         /* ISO 639-1 or NULL */
    const char* device);
void    ml_asr_destroy(ml_ASR* handle);
bool    ml_asr_load_model(ml_ASR* handle, ml_Path model_path, const void* extra_data);
void    ml_asr_close(ml_ASR* handle);

/* ====================  Transcription  ===================================== */
ml_ASRResult ml_asr_transcribe(
    ml_ASR* handle, const float* audio, int32_t num_samples, int32_t sample_rate, const ml_ASRConfig* config);

ml_ASRResult* ml_asr_transcribe_batch(ml_ASR* handle, const float** audios, const int32_t* num_samples_array,
    int32_t batch_size, int32_t sample_rate, const ml_ASRConfig* config);

ml_ASRResult ml_asr_transcribe_step(
    ml_ASR* handle, const float* audio_chunk, int32_t num_samples, int32_t step, const ml_ASRConfig* config);

/* ====================  Result Management  ================================ */
void ml_asr_print_result(const ml_ASRResult* result);
void ml_asr_free_result(ml_ASRResult* result);

/* ====================  Language Management  ============================== */
const char** ml_asr_list_supported_languages(const ml_ASR* handle, int32_t* out_count);
void         ml_asr_set_language(ml_ASR* handle, const char* language);

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
    float*  audio;            /* Audio samples: num_samples × channels */
    float   duration_seconds; /* Audio duration in seconds */
    int32_t sample_rate;      /* Audio sample rate in Hz */
    int32_t channels;         /* Number of audio channels (default: 1) */
    int32_t num_samples;      /* Number of audio samples */
} ml_TTSResult;

typedef struct ml_TTS ml_TTS; /* Opaque TTS handle */

/* ====================  Lifecycle Management  ============================== */
ml_TTS* ml_tts_create(ml_Path model_path, ml_Path vocoder_path, const char* device);
void    ml_tts_destroy(ml_TTS* handle);
bool    ml_tts_load_model(ml_TTS* handle, ml_Path model_path, const void* extra_data);
void    ml_tts_close(ml_TTS* handle);

/* ====================  Configuration  ===================================== */
void ml_tts_set_sampler(ml_TTS* handle, const ml_TTSSamplerConfig* config);
void ml_tts_reset_sampler(ml_TTS* handle);

/* ====================  Speech Synthesis  ================================== */
ml_TTSResult  ml_tts_synthesize(ml_TTS* handle, const char* text_utf8, const ml_TTSConfig* config);
ml_TTSResult* ml_tts_synthesize_batch(
    ml_TTS* handle, const char** texts, int32_t text_count, const ml_TTSConfig* config);
ml_TTSResult ml_tts_synthesize_step(ml_TTS* handle, const char* text_utf8, int32_t step, const ml_TTSConfig* config);

/* ====================  Cache Management  ================================== */
void ml_tts_save_cache(ml_TTS* handle, ml_Path path);
void ml_tts_load_cache(ml_TTS* handle, ml_Path path);

/* ====================  Voice Management  ================================== */
const char** ml_tts_list_available_voices(const ml_TTS* handle, int32_t* out_count);

/* ====================  Result Management  ================================ */
void ml_tts_print_result(const ml_TTSResult* result);
void ml_tts_free_result(ml_TTSResult* result);

#ifdef __cplusplus
} /* extern "C" */
#endif
