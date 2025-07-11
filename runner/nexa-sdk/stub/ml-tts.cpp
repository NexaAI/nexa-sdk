#include "ml.h"

extern "C" {

/* ====================  Lifecycle Management  ============================== */
ml_TTS* ml_tts_create(ml_Path model_path, ml_Path vocoder_path, const char* device) {
    return nullptr;
}

void ml_tts_destroy(ml_TTS* handle) {}

bool ml_tts_load_model(ml_TTS* handle, ml_Path model_path, const void* extra_data) {
    return false;
}

void ml_tts_close(ml_TTS* handle) {}

/* ====================  Configuration  ===================================== */
void ml_tts_set_sampler(ml_TTS* handle, const ml_TTSSamplerConfig* config) {}

void ml_tts_reset_sampler(ml_TTS* handle) {}

/* ====================  Speech Synthesis  ================================== */
ml_TTSResult ml_tts_synthesize(ml_TTS* handle, const char* text_utf8, const ml_TTSConfig* config) {
    ml_TTSResult result = {0};
    result.audio = nullptr;
    result.duration_seconds = 0.0f;
    result.sample_rate = 0;
    result.channels = 0;
    result.num_samples = 0;
    return result;
}

ml_TTSResult* ml_tts_synthesize_batch(ml_TTS* handle, const char** texts, int32_t text_count, const ml_TTSConfig* config) {
    return nullptr;
}

ml_TTSResult ml_tts_synthesize_step(ml_TTS* handle, const char* text_utf8, int32_t step, const ml_TTSConfig* config) {
    ml_TTSResult result = {0};
    result.audio = nullptr;
    result.duration_seconds = 0.0f;
    result.sample_rate = 0;
    result.channels = 0;
    result.num_samples = 0;
    return result;
}

/* ====================  Cache Management  ================================== */
void ml_tts_save_cache(ml_TTS* handle, ml_Path path) {}

void ml_tts_load_cache(ml_TTS* handle, ml_Path path) {}

/* ====================  Voice Management  ================================== */
const char** ml_tts_list_available_voices(const ml_TTS* handle, int32_t* out_count) {
    if (out_count) {
        *out_count = 0;
    }
    return nullptr;
}

/* ====================  Result Management  ================================ */
void ml_tts_print_result(const ml_TTSResult* result) {}

void ml_tts_free_result(ml_TTSResult* result) {}

}
