#include "ml.h"

extern "C" {

/* ====================  Lifecycle Management  ============================== */
ml_Embedder* ml_embedder_create(ml_Path model_path, ml_Path tokenizer_path, const char* device) { return nullptr; };
void         ml_embedder_destroy(ml_Embedder* handle) {};

/* ====================  Embedding Generation  ============================= */
int32_t ml_embedder_embed(
    ml_Embedder* handle, const char** texts, int32_t text_count, const ml_EmbeddingConfig* config, float** out) {
    return -255;
};

/* ====================  Model Information  ================================ */
int32_t ml_embedder_embedding_dim(const ml_Embedder* handle) { return -255; };

/* ====================  LoRA Management  ================================== */
void    ml_embedder_set_lora(ml_Embedder* handle, int32_t lora_id) {};
int32_t ml_embedder_add_lora(ml_Embedder* handle, ml_Path lora_path) { return -255; };
void    ml_embedder_remove_lora(ml_Embedder* handle, int32_t lora_id) {};
int32_t ml_embedder_list_loras(const ml_Embedder* handle, int32_t** out) { return -255; };

}
