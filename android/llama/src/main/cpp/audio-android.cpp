#include <android/log.h>
#include <jni.h>
#include <iomanip>
#include <math.h>
#include <string>
#include <unistd.h>
#include "omni.cpp"
#include <nlohmann/json.hpp>
#include <jni.h>
#include <string>
#include <iostream>
#include <thread>

#define TAG "audio-android.cpp"
#define LOGi(...) __android_log_print(ANDROID_LOG_INFO, TAG, __VA_ARGS__)
#define LOGe(...) __android_log_print(ANDROID_LOG_ERROR, TAG, __VA_ARGS__)

extern bool is_valid_utf8(const char* str);

extern std::string jstring2str(JNIEnv* env, jstring jstr);


// 用于捕获输出的函数
void redirect_output_to_logcat(const char* tag, int fd) {
    char buffer[1024];
    while (true) {
        ssize_t count = read(fd, buffer, sizeof(buffer) - 1);
        if (count <= 0) break;
        buffer[count] = '\0';
        __android_log_print(ANDROID_LOG_DEBUG, tag, "%s", buffer);
    }
}

// 初始化重定向
void setup_redirect_stdout_stderr() {
    int stdout_pipe[2];
    int stderr_pipe[2];

    pipe(stdout_pipe);
    pipe(stderr_pipe);

    // 重定向 stdout
    dup2(stdout_pipe[1], STDOUT_FILENO);
    close(stdout_pipe[1]);
    std::thread(redirect_output_to_logcat, "STDOUT", stdout_pipe[0]).detach();

    // 重定向 stderr
    dup2(stderr_pipe[1], STDERR_FILENO);
    close(stderr_pipe[1]);
    std::thread(redirect_output_to_logcat, "STDERR", stderr_pipe[0]).detach();
}

JNIEXPORT jint JNICALL JNI_OnLoad(JavaVM* vm, void* reserved) {
    setup_redirect_stdout_stderr();
    return JNI_VERSION_1_6;
}


extern "C" JNIEXPORT jlong JNICALL
Java_com_nexa_NexaAudioInference_init_1params(JNIEnv *env, jobject /* this */) {
    const char* argv = "-t 1";
    char* nc_argv = const_cast<char*>(argv);
    omni_context_params* ctx_params = new omni_context_params();
    omni_context_params_parse(argc, argv, ctx_params)

    return reinterpret_cast<jlong>(ctx_params);
}

extern "C" JNIEXPORT jlong JNICALL
Java_com_nexa_NexaAudioInference_init_1params(JNIEnv *env, jobject /* this */) {
    const char* argv = "-t 1";
    char* nc_argv = const_cast<char*>(argv);
    omni_context_params* ctx_params = new omni_context_params();
    omni_context_params_parse(argc, argv, ctx_params)

    return reinterpret_cast<jlong>(ctx_params);
}