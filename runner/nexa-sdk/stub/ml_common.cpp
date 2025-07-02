#include "ml.h"

extern "C" {
/* ====================  Core Initialization  ================================ */
void ml_init(void) {};                        /* Initialization */
void ml_deinit(void) {};                      /* Cleanup */
void ml_set_log(ml_log_callback callback) {}; /* Set logging callback */
void ml_log(const char* message) {};          /* Log a message */
void ml_free(void* ptr) {};                   /* Free allocated memory */
}
