#include "nexa/llm/llm.h"
#include "nexa/util/log.h"
#include <llama.h>

void go_log_wrap(char* msg);

static void llama_log_callback(ggml_log_level level, const char* msg, void* user_data) {
    go_log_wrap((char*)msg);
}

int ml_init() {
    llama_log_set(llama_log_callback, NULL);
    log_init();
    log_set_callback(go_log_wrap);
    return 0;
} 