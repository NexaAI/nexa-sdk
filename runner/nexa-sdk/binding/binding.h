#include "llama.h"

#ifdef __cplusplus
extern "C" {
#endif

void init();

struct LLMPipeline;
struct LLMPipeline *llm_pipeline_new();
void llm_pipeline_free(struct LLMPipeline *);
bool llm_pipeline_load_model(struct LLMPipeline *, char *);
void llm_pipeline_close(struct LLMPipeline *);
// === Tokenizer ===
// === KV Cache ===
// === Lora ===
// === Sampler ===
// === Generation ===
int32_t llm_pipeline_generate(struct LLMPipeline *, char *, char *);
bool llm_pipeline_generate_send(struct LLMPipeline *, char *);
int32_t llm_pipeline_generate_next_token(struct LLMPipeline *, char *);

#ifdef __cplusplus
}  // extern "C"
#endif
