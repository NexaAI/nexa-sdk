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

void redirect_output_to_logcat(const char* tag, int fd) {
    char buffer[1024];
    while (true) {
        ssize_t count = read(fd, buffer, sizeof(buffer) - 1);
        if (count <= 0) break;
        buffer[count] = '\0';
        __android_log_print(ANDROID_LOG_DEBUG, tag, "%s", buffer);
    }
}

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
Java_com_nexa_NexaAudioInference_init_1ctx_1params(JNIEnv *env, jobject /* this */, jstring jmodel, jstring jprojector, jstring jaudio) {
    const char* model = env->GetStringUTFChars(jmodel, nullptr);
    const char* projector = env->GetStringUTFChars(jprojector, nullptr);
    const char* audio = env->GetStringUTFChars(jaudio, nullptr);
    const char* argv[] = {"-t", "1"};
    int argc = 1;
    omni_context_params* ctx_params = new omni_context_params(omni_context_default_params());
    omni_context_params_parse(argc, const_cast<char**>(argv), *ctx_params);
    ctx_params->model = model;
    ctx_params->mmproj = projector;
    ctx_params->file = audio;

    return reinterpret_cast<jlong>(ctx_params);
}

extern "C" JNIEXPORT jlong JNICALL
Java_com_nexa_NexaAudioInference_init_1ctx(JNIEnv *env, jobject /* this */, jlong jctx_params) {
    auto* ctx_params = reinterpret_cast<omni_context_params *>(jctx_params);
    std::cout << ctx_params->n_gpu_layers << std::endl;
    std::cout << ctx_params->model << std::endl;
    std::cout << ctx_params->mmproj << std::endl;
    std::cout << ctx_params->file << std::endl;
    omni_context *ctx_omni = omni_init_context(*ctx_params);
    return reinterpret_cast<jlong>(ctx_omni);
}

extern "C" JNIEXPORT void JNICALL
Java_com_nexa_NexaAudioInference_free_1ctx(JNIEnv *env, jobject /* this */, jlong jctx_omni) {
    auto* ctx_omni = reinterpret_cast<omni_context *>(jctx_omni);
    omni_free(ctx_omni);
}

extern "C" JNIEXPORT jlong JNICALL
Java_com_nexa_NexaAudioInference_init_1npast(JNIEnv *env, jobject /* this */) {
    int* n_past = new int(0);
    return reinterpret_cast<jlong>(n_past);
}

extern "C" JNIEXPORT jlong JNICALL
Java_com_nexa_NexaAudioInference_init_1params(JNIEnv *env, jobject /* this */, jlong jctx_params) {
    auto* ctx_params = reinterpret_cast<omni_context_params *>(jctx_params);

    if (ctx_params == nullptr) {
        std::cerr << "Error: jctx_params is null!" << std::endl;
        return 0;  // Return 0 (null) if the context parameter is invalid.
    }

    // Step 2: Call the function to extract omni_params from ctx_params.
    omni_params extracted_params;
    try {
        extracted_params = get_omni_params_from_context_params(*ctx_params);
    } catch (const std::exception& e) {
        std::cerr << "Error in get_omni_params_from_context_params: " << e.what() << std::endl;
        return 0;  // Return 0 (null) if an exception is thrown during the extraction.
    }

    // Step 3: Allocate memory for omni_params and ensure it's successful.
    omni_params* all_params = nullptr;
    try {
        all_params = new omni_params(extracted_params);
    } catch (const std::bad_alloc& e) {
        std::cerr << "Error: Failed to allocate memory for omni_params: " << e.what() << std::endl;
        return 0;  // Return 0 (null) if memory allocation fails.
    }

    std::cout << " fname_inp size: " << all_params->whisper.fname_inp.size() << std::endl;

    // Step 4: Return the pointer to the newly allocated omni_params object.
    std::cout << "all_params address: " << all_params << std::endl;
    return reinterpret_cast<jlong>(all_params);
}


//val sampler = init_sampler(allParamsPointer, ctxParamsPointer, prompt, audiuo_path, npastPointer)
extern "C" JNIEXPORT jlong JNICALL
Java_com_nexa_NexaAudioInference_init_1sampler(JNIEnv *env, jobject /* this */, jlong jctx_omni, jlong jctx_params, jstring jprompt, jstring jaudio_path, jlong jnpast) {
    auto* n_past = reinterpret_cast<int*>(jnpast);
    if (n_past == nullptr) {
    std::cout << "n_past is null!" << std::endl;
    }
    const char* prompt = env->GetStringUTFChars(jprompt, nullptr);
    auto* all_params = reinterpret_cast<omni_params *>(jctx_params);
    auto* ctx_omni = reinterpret_cast<omni_context *>(jctx_omni);

    ggml_tensor *audio_embed = omni_process_audio(ctx_omni, *all_params);
    std::string system_prompt, user_prompt;
    system_prompt = "<start_of_turn>user\nAudio 1: <|audio_bos|>";
    user_prompt = "<|audio_eos|>\n" + std::string(prompt) + "<end_of_turn>\n<start_of_turn>model\n";

    eval_string(ctx_omni->ctx_llama, system_prompt.c_str(), all_params->gpt.n_batch, n_past, true);
    omni_eval_audio_embed(ctx_omni->ctx_llama, audio_embed, all_params->gpt.n_batch, n_past);
    eval_string(ctx_omni->ctx_llama, user_prompt.c_str(), all_params->gpt.n_batch, n_past, false);

    struct common_sampler * ctx_sampling = common_sampler_init(ctx_omni->model, all_params->gpt.sampling);

    return reinterpret_cast<jlong>(ctx_sampling);
}


extern "C" JNIEXPORT jstring JNICALL
Java_com_nexa_NexaAudioInference_sampler(JNIEnv *env, jobject /* this */, jlong jctx_omni, jlong jsampler, jlong jnpast) {
    auto* ctx_omni = reinterpret_cast<omni_context *>(jctx_omni);
    auto* sampler = reinterpret_cast<common_sampler *>(jsampler);
    auto* n_past = reinterpret_cast<int*>(jnpast);

    const char * tmp = sample(sampler, ctx_omni->ctx_llama, n_past);

    jstring new_token = nullptr;
    new_token = env->NewStringUTF(tmp);
    return new_token;
}


extern "C" JNIEXPORT jstring JNICALL
Java_com_nexa_NexaAudioInference_free_1sampler(JNIEnv *env, jobject /* this */, jlong jsampler) {
    auto* sampler = reinterpret_cast<common_sampler *>(jsampler);
    common_sampler_free(sampler);
}
