#include "ml.h"

extern "C" {

/* ====================  Lifecycle Management  ============================== */
ml_ASR* ml_asr_create(ml_Path model_path, ml_Path tokenizer_path, const char* language, const char* device) {
    return nullptr;
}

void ml_asr_destroy(ml_ASR* handle) {}

bool ml_asr_load_model(ml_ASR* handle, ml_Path model_path, const void* extra_data) {
    return false;
}

void ml_asr_close(ml_ASR* handle) {}

/* ====================  Transcription  ===================================== */
ml_ASRResult ml_asr_transcribe(ml_ASR* handle, const float* audio, int32_t num_samples, int32_t sample_rate, const ml_ASRConfig* config) {
    ml_ASRResult result = {0};
    result.transcript = nullptr;
    result.confidence_scores = nullptr;
    result.confidence_count = 0;
    result.timestamps = nullptr;
    result.timestamp_count = 0;
    return result;
}

ml_ASRResult* ml_asr_transcribe_batch(ml_ASR* handle, const float** audios, const int32_t* num_samples_array, int32_t batch_size, int32_t sample_rate, const ml_ASRConfig* config) {
    return nullptr;
}

ml_ASRResult ml_asr_transcribe_step(ml_ASR* handle, const float* audio_chunk, int32_t num_samples, int32_t step, const ml_ASRConfig* config) {
    ml_ASRResult result = {0};
    result.transcript = nullptr;
    result.confidence_scores = nullptr;
    result.confidence_count = 0;
    result.timestamps = nullptr;
    result.timestamp_count = 0;
    return result;
}

/* ====================  Result Management  ================================ */
void ml_asr_print_result(const ml_ASRResult* result) {}

void ml_asr_free_result(ml_ASRResult* result) {}

/* ====================  Language Management  ============================== */
const char** ml_asr_list_supported_languages(const ml_ASR* handle, int32_t* out_count) {
    if (out_count) {
        *out_count = 0;
    }
    return nullptr;
}

void ml_asr_set_language(ml_ASR* handle, const char* language) {}

}
