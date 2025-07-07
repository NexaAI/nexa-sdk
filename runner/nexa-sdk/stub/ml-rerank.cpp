#include "ml.h"

extern "C" {

/* ====================  Lifecycle Management  ============================== */
ml_Reranker *ml_reranker_create(ml_Path model_path, ml_Path tokenizer_path, const char *device) { return nullptr; };
void         ml_reranker_destroy(ml_Reranker *handle) {};

/* ====================  Reranking  ========================================= */
int32_t ml_reranker_rerank(ml_Reranker *handle, const char *query, const char **documents, int32_t documents_count,
    const ml_RerankConfig *config, float **out) {
    return -255;
};

/* ====================  Profiling Data  ================================ */
int32_t ml_reranker_get_profiling_data(const ml_Reranker *handle, ml_ProfilingData *out_data) { return -255; };
}
