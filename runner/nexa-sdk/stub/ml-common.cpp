#include "ml.h"

extern "C" {

const char* ml_get_error_message(ml_ErrorCode error_code) { return ""; }
/* ====================  Core Initialization  ============================== */
void ml_init(void) {};                        /* Initialization */
void ml_deinit(void) {};                      /* Cleanup */
void ml_set_log(ml_log_callback callback) {}; /* Set logging callback */
void ml_log(const char *message) {};          /* Log a message */
void ml_free(void *ptr) {};                   /* Free allocated memory */
const char* ml_version() { return "stub"; }

/* ====================  Data Structures  ================================== */
void ml_image_free(ml_Image *image) {}
void ml_image_save(const ml_Image *image, const char *filename) {}

void ml_audio_free(ml_Audio *audio) {}
void ml_audio_save(const ml_Audio *audio, const char *filename) {}

void ml_video_free(ml_Video *video) {}
void ml_video_save(const ml_Video *video, const char *filename) {}
}
