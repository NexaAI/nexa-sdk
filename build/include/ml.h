#pragma once

#include <stdbool.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

/* ------------------------------------------------------------------ */
/* ==================  Common ======================================= */
/* ------------------------------------------------------------------ */

typedef const char *ml_Path; /* path-strings are plain char* */

typedef void (*ml_log_callback)(const char *);
typedef bool (*ml_llm_token_callback)(const char *token, void *user_data);

void ml_init();
void ml_deinit();
void ml_set_log(ml_log_callback c);
void ml_log(const char *);

/* ------------------------------------------------------------------ */
/* ==================  LLM  ========================================= */
/* ------------------------------------------------------------------ */

typedef struct {
    float   temperature;
    float   top_p;
    int32_t top_k;
    float   repetition_penalty;
    float   presence_penalty;
    float   frequency_penalty;
    int32_t seed;         /* -1 ⇒ random */
    ml_Path grammar_path; /* optional */
} ml_SamplerConfig;

typedef struct {
    int32_t          max_tokens;
    float            temperature;
    float            top_p;
    int32_t          top_k;
    float            repetition_penalty;
    const char     **stop; /* array of stop words */
    int32_t          stop_count;
    ml_SamplerConfig sampler_config;  // TODO: not used
} ml_GenerationConfig;

typedef struct {
    const char *role;
    const char *content;
} ml_ChatMessage;

/* opaque handle forward-declared in the API */
typedef struct ml_LLM ml_LLM;

ml_LLM *ml_llm_create(ml_Path model_path, ml_Path /*tokenizer_path – unused*/, int32_t context_length,
    const char * /*device – llama.cpp handles via params*/);
void    ml_llm_destroy(ml_LLM *h);
void    ml_llm_reset(ml_LLM *h);

/* ====================  Tokeniser  ================================= */
int32_t ml_llm_encode(const ml_LLM *h, const char *text_utf8, int32_t **out);
int32_t ml_llm_decode(const ml_LLM *h, const int32_t *ids, int32_t len, char **out);

/* ====================  KV-cache save/load  ======================= */
// TODO
int32_t ml_llm_save_kv_cache(const ml_LLM *h, ml_Path path);
int32_t ml_llm_load_kv_cache(ml_LLM *h, ml_Path path);

/* ====================  (LoRA — minimal)  ========================= */
void    ml_llm_set_lora(ml_LLM    */*h*/, int32_t /*id*/);
int32_t ml_llm_add_lora(ml_LLM * /*h*/, ml_Path /*path*/);
void    ml_llm_remove_lora(ml_LLM    */*h*/, int32_t /*id*/);
int32_t ml_llm_list_loras(const ml_LLM * /*h*/, int32_t **out);

/* ====================  Sampler helpers  ========================== */
void ml_llm_set_sampler(ml_LLM *h, const ml_SamplerConfig *cfg);
void ml_llm_reset_sampler(ml_LLM *h);

/* ====================  Generation  =============================== */
int32_t ml_llm_generate(ml_LLM *h, const char *prompt_utf8, const ml_GenerationConfig *cfg_in, char **out);
int32_t ml_llm_get_chat_template(ml_LLM *h, const char *name, const char **out);
int32_t ml_llm_apply_chat_template(ml_LLM *h, ml_ChatMessage *messages, int32_t message_count, char **out);

/* ====================  Embedding  ================================= */
int32_t ml_llm_embed(ml_LLM *h, const char **texts_utf8, int32_t text_count, float **out);

/* ====================  Streaming Generation  =============================== */
int32_t ml_llm_generate_stream(ml_LLM *h, const char *prompt_utf8, const ml_GenerationConfig *cfg_in,
    ml_llm_token_callback on_token, void *user_data, char **out_full);

/* ------------------------------------------------------------------ */
/* ==================  Embedding  ================================== */
/* ------------------------------------------------------------------ */

typedef struct {
    int32_t     batch_size;
    bool        normalize;
    const char *normalize_method; /* "l2", "mean", "none" */
} ml_EmbeddingConfig;

typedef struct ml_Embedder ml_Embedder;

/* Constructor / destructor */
ml_Embedder *ml_embedder_create(ml_Path model_path, ml_Path tokenizer_path, const char *device);
void         ml_embedder_destroy(ml_Embedder *h);

/* Lifecycle */
bool ml_embedder_load_model(ml_Embedder *h, ml_Path model_path, const void *extra);
void ml_embedder_close(ml_Embedder *h);

/* Embedding API
   Returns a row-major matrix of size (text_count × dim).
   Caller owns the returned pointer (free with ml_free). */
float *ml_embedder_embed(
    ml_Embedder *h, const char **texts, int32_t text_count, const ml_EmbeddingConfig *cfg, int32_t *out_dim);

/* Model info */
int32_t ml_embedder_embedding_dim(const ml_Embedder *h);

/* LoRA management */
void     ml_embedder_set_lora(ml_Embedder *h, int32_t lora_id);
int32_t  ml_embedder_add_lora(ml_Embedder *h, ml_Path lora_path);
void     ml_embedder_remove_lora(ml_Embedder *h, int32_t lora_id);
int32_t *ml_embedder_list_loras(const ml_Embedder *h, int32_t *out_count);

/* ------------------------------------------------------------------ */
/* ==================  Rerank  ===================================== */
/* ------------------------------------------------------------------ */
typedef struct {
    int32_t     batch_size;
    bool        normalize;
    const char *normalize_method; /* "softmax", "min-max", "none" */
} ml_RerankConfig;

typedef struct ml_Reranker ml_Reranker;

ml_Reranker *ml_reranker_create(ml_Path model_path, ml_Path tokenizer_path, const char *device);
void         ml_reranker_destroy(ml_Reranker *h);

bool ml_reranker_load_model(ml_Reranker *h, ml_Path model_path, const void *extra);
void ml_reranker_close(ml_Reranker *h);

/* Returns score array of length documents_count (caller frees) */
float *ml_reranker_rerank(
    ml_Reranker *h, const char *query, const char **documents, int32_t documents_count, const ml_RerankConfig *cfg);

/* ------------------------------------------------------------------ */
/* ==================  Image generation  ============================ */
/* ------------------------------------------------------------------ */

/* ------------------------------------------------------------------ */
/* ==================  CV  ========================================= */
/* ------------------------------------------------------------------ */

#ifdef __cplusplus
} /* extern "C" */
#endif
