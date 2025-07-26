#include "ml.h"

extern "C" {

/* ====================  Lifecycle Management  ============================== */
ml_VLM *ml_vlm_create(ml_Path model_path, ml_Path mmproj_path, ml_ModelConfig config, const char *device) {
    return nullptr;
};
void ml_vlm_destroy(ml_VLM *handle) {};
void ml_vlm_reset(ml_VLM *handle) {}; /* Reset internal state */

/* ====================  Tokenization  ====================================== */
int32_t ml_vlm_encode(const ml_VLM *handle, const char *text_utf8, int32_t **out_tokens) { return -255; };
int32_t ml_vlm_decode(const ml_VLM *handle, const int32_t *token_ids, int32_t length, char **out_text) { return -255; };

/* ====================  Sampling Configuration  =========================== */
void ml_vlm_set_sampler(ml_VLM *handle, const ml_SamplerConfig *config) {};
void ml_vlm_reset_sampler(ml_VLM *handle) {};

/** Print detailed performance profile (sampler + context) */
void ml_vlm_print_profile(const ml_VLM *handle) {};

/* ====================  Text Generation  ================================== */
int32_t ml_vlm_generate(ml_VLM *handle, const char *prompt_utf8, const ml_GenerationConfig *config, char **out_text) {
    return -255;
};
int32_t ml_vlm_get_chat_template(ml_VLM *handle, const char *template_name, const char **out_template) { return -255; };
int32_t ml_vlm_apply_chat_template(ml_VLM *handle, ml_VlmChatMessage *messages, int32_t message_count, ml_Tool *tools, int32_t tool_count, char **out_text) {
    return -255;
};

/* ====================  Profiling Data  ================================ */
int32_t ml_vlm_get_profiling_data(const ml_VLM *handle, ml_ProfilingData *out_data) { return -255; };

/* ====================  Streaming Generation  ============================= */
int32_t ml_vlm_generate_stream(ml_VLM *handle, const char *prompt_utf8, const ml_GenerationConfig *config,
    ml_llm_token_callback on_token, void *user_data, char **out_full_text) {
    return -255;
};
}
