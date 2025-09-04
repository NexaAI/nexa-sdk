#include "ml.h"
#include <stdint.h>
#include <string.h>

const char *ml_get_error_message(const ml_ErrorCode error_code) {
  return "unknown";
}

int32_t ml_init(void) { return 0; }
int32_t ml_register_plugin(ml_plugin_id_func plugin_id_func,
                           ml_create_plugin_func create_func) {
  return 0;
}
int32_t ml_deinit(void) { return 0; }
int32_t ml_set_log(ml_log_callback callback) { return 0; }
void ml_free(void *ptr) {}
const char *ml_version(void) { return "v1.0.3-qnn-rc4"; }

int32_t ml_get_plugin_list(ml_GetPluginListOutput *output) { return 0; }

int32_t ml_get_device_list(const ml_GetDeviceListInput *input,
                           ml_GetDeviceListOutput *output) {
  return 0;
}

void ml_image_free(ml_Image *image) {}
void ml_image_save(const ml_Image *image, const char *filename) {}

void ml_audio_free(ml_Audio *audio) {}

void ml_video_free(ml_Video *video) {}

int32_t ml_llm_apply_chat_template(ml_LLM *handle,
                                   const ml_LlmApplyChatTemplateInput *input,
                                   ml_LlmApplyChatTemplateOutput *output) {
  output->formatted_text =
      strdup(input->messages[input->message_count - 1].content);
  return strlen(output->formatted_text);
}

// int32_t ml_vlm_create(const ml_VlmCreateInput *input, ml_VLM **out_handle) {
//   return -1;
// }
//
// int32_t ml_vlm_destroy(ml_VLM *handle) { return -1; }
//
// int32_t ml_vlm_reset(ml_VLM *handle) { return -1; }
//
// int32_t ml_vlm_apply_chat_template(ml_VLM *handle,
//                                    const ml_VlmApplyChatTemplateInput *input,
//                                    ml_VlmApplyChatTemplateOutput *output) {
//   return -1;
// }
//
// int32_t ml_vlm_generate(ml_VLM *handle, const ml_VlmGenerateInput *input,
//                         ml_VlmGenerateOutput *output) {
//   return -1;
// }

int32_t ml_embedder_create(const ml_EmbedderCreateInput *input,
                           ml_Embedder **out_handle) {
  return -1;
}

int32_t ml_embedder_destroy(ml_Embedder *handle) { return -1; }

int32_t ml_embedder_embed(ml_Embedder *handle,
                          const ml_EmbedderEmbedInput *input,
                          ml_EmbedderEmbedOutput *output) {
  return -1;
}

int32_t ml_embedder_embedding_dim(const ml_Embedder *handle,
                                  ml_EmbedderDimOutput *output) {
  return -1;
}

int32_t ml_reranker_create(const ml_RerankerCreateInput *input,
                           ml_Reranker **out_handle) {
  return -1;
}

int32_t ml_reranker_destroy(ml_Reranker *handle) { return -1; }

int32_t ml_reranker_rerank(ml_Reranker *handle,
                           const ml_RerankerRerankInput *input,
                           ml_RerankerRerankOutput *output) {
  return -1;
}

int32_t ml_imagegen_create(const ml_ImageGenCreateInput *input,
                           ml_ImageGen **out_handle) {
  return -1;
}

int32_t ml_imagegen_destroy(ml_ImageGen *handle) { return -1; }

int32_t ml_imagegen_txt2img(ml_ImageGen *handle,
                            const ml_ImageGenTxt2ImgInput *input,
                            ml_ImageGenOutput *output) {
  return -1;
}

int32_t ml_imagegen_img2img(ml_ImageGen *handle,
                            const ml_ImageGenImg2ImgInput *input,
                            ml_ImageGenOutput *output) {
  return -1;
}

// int32_t ml_asr_create(const ml_AsrCreateInput *input, ml_ASR **out_handle) {
//   return -1;
// }

// int32_t ml_asr_destroy(ml_ASR *handle) { return -1; }

// int32_t ml_asr_transcribe(ml_ASR *handle, const ml_AsrTranscribeInput *input,
//                           ml_AsrTranscribeOutput *output) {
//   return -1;
// }

// int32_t
// ml_asr_list_supported_languages(const ml_ASR *handle,
//                                 const ml_AsrListSupportedLanguagesInput *input,
//                                 ml_AsrListSupportedLanguagesOutput *output) {
//   return -1;
// }

int32_t ml_tts_create(const ml_TtsCreateInput *input, ml_TTS **out_handle) {
  return -1;
}

int32_t ml_tts_destroy(ml_TTS *handle) { return -1; }

int32_t ml_tts_synthesize(ml_TTS *handle, const ml_TtsSynthesizeInput *input,
                          ml_TtsSynthesizeOutput *output) {
  return -1;
}

int32_t
ml_tts_list_available_voices(const ml_TTS *handle,
                             const ml_TtsListAvailableVoicesInput *input,
                             ml_TtsListAvailableVoicesOutput *output) {
  return -1;
}
