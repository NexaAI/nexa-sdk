#include "ml.h"

extern "C" {

/* ====================  Lifecycle Management  ============================== */
ml_LLM* ml_llm_create(ml_Path model_path, ml_Path tokenizer_path, ml_ModelConfig config, const char* device) {
    return nullptr;
};
void ml_llm_destroy(ml_LLM* handle) {};
void ml_llm_reset(ml_LLM* handle) {}; /* Reset internal state */

/* ====================  Tokenization  ====================================== */
int32_t ml_llm_encode(const ml_LLM* handle, const char* text_utf8, int32_t** out_tokens) { return -255; };
int32_t ml_llm_decode(const ml_LLM* handle, const int32_t* token_ids, int32_t length, char** out_text) { return -255; };

/* ====================  KV-Cache Management  ============================== */
int32_t ml_llm_save_kv_cache(const ml_LLM* handle, ml_Path path) { return -255; };
int32_t ml_llm_load_kv_cache(ml_LLM* handle, ml_Path path) { return -255; };

/* ====================  LoRA Management  ================================== */
void    ml_llm_set_lora(ml_LLM* handle, int32_t lora_id) {};
int32_t ml_llm_add_lora(ml_LLM* handle, ml_Path lora_path) { return -255; };
void    ml_llm_remove_lora(ml_LLM* handle, int32_t lora_id) {};
int32_t ml_llm_list_loras(const ml_LLM* handle, int32_t** out_lora_ids) { return -255; };

/* ====================  Sampling Configuration  =========================== */
void ml_llm_set_sampler(ml_LLM* handle, const ml_SamplerConfig* config) {};
void ml_llm_reset_sampler(ml_LLM* handle) {};

/* ====================  Text Generation  ================================== */
int32_t ml_llm_generate(ml_LLM* handle, const char* prompt_utf8, const ml_GenerationConfig* config, char** out_text) {
    return -255;
};
int32_t ml_llm_get_chat_template(ml_LLM* handle, const char* template_name, const char** out_template) { return -255; };
int32_t ml_llm_apply_chat_template(ml_LLM* handle, ml_ChatMessage* messages, int32_t message_count, char** out_text) {
    return -255;
};

/* ====================  Embedding Generation  ============================= */
int32_t ml_llm_embed(ml_LLM* handle, const char** texts_utf8, int32_t text_count, float** out_embeddings) {
    return -255;
};

/* ====================  Streaming Generation  ============================= */
int32_t ml_llm_generate_stream(ml_LLM* handle, const char* prompt_utf8, const ml_GenerationConfig* config,
    ml_llm_token_callback on_token, void* user_data, char** out_full_text) {
    return -255;
};
}
