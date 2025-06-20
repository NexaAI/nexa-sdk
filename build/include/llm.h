#include <stdint.h>

#include "common.h"

/* ------------------------------------------------------------------ */
/* ==================  LLM  ========================================= */
/* ------------------------------------------------------------------ */

extern "C" {

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
int32_t ml_llm_save_kv_cache(const ml_LLM *h, ml_Path path, int32_t tok_prefix);
int32_t ml_llm_load_kv_cache(ml_LLM *h, ml_Path path, int32_t /*tok_prefix*/);

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
int32_t ml_llm_apply_chat_template(ml_LLM *h, const char *role, const char *prompt_utf8, char **out);

/* ====================  Embedding  ================================= */
int32_t ml_llm_embed(ml_LLM *h, const char **texts_utf8, int32_t text_count, float **out);

/* ====================  Streaming Generation  =============================== */
int32_t ml_llm_generate_stream(ml_LLM *h, const char *prompt_utf8, const ml_GenerationConfig *cfg_in,
    ml_llm_token_callback on_token, void *user_data, char **out_full);

}  // extern "C"
