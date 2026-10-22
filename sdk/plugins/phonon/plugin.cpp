// SPDX-License-Identifier: BSD-3-Clause
#include "plugin/Plugin.h"

#include <algorithm>
#include <atomic>
#include <cstdlib>
#include <cstring>
#include <memory>
#include <mutex>

#include "phonon.hpp"
#include "plugin/ITts.h"

namespace {
const char* optional(const char* value) { return value ? value : ""; }

class PhononTts final : public geniex::ITts {
    std::unique_ptr<phonon::Phonon> model_;
    std::atomic<bool>               cancelled_{false};
    std::mutex                      mutex_;
    bool                            verify_ = false;
    phonon::Stats                   previous_;

   public:
    int32_t create(const geniex_TtsCreateInput* input) override {
        if (std::strcmp(input->device, "npu") != 0) return GENIEX_ERROR_COMMON_INVALID_DEVICE;
        if (!input->runtime_path || !*input->runtime_path || !input->text_library || !*input->text_library ||
            !input->soc_model || !*input->soc_model)
            return GENIEX_ERROR_COMMON_INVALID_INPUT;
        verify_             = (input->flags & GENIEX_TTS_VERIFY_NPU) != 0;
        const char* profile = std::getenv("PHONON_QNN_PROFILE");
        if (verify_ && (!profile || std::strcmp(profile, "1") != 0)) {
            GENIEX_LOG_ERROR("TTS verification requires PHONON_QNN_PROFILE=1 before loading the model");
            return GENIEX_ERROR_COMMON_INVALID_INPUT;
        }
        phonon::Options options;
        options.bundle_dir   = input->model_path;
        options.backend      = "htp";  // Strict QNN HTP: never retry with CPU or GPU.
        options.lib_dir      = input->runtime_path;
        options.text_library = input->text_library;
        options.lang         = optional(input->language);
        options.soc_model    = input->soc_model;
        model_               = std::make_unique<phonon::Phonon>(options);
        return GENIEX_SUCCESS;
    }

    int32_t cancel() override {
        cancelled_.store(true);
        return GENIEX_SUCCESS;
    }

    int32_t synthesize(const geniex_TtsSynthesizeInput* input, geniex_TtsSynthesizeOutput* output) override {
        std::unique_lock<std::mutex> lock(mutex_, std::try_to_lock);
        if (!lock.owns_lock()) return GENIEX_ERROR_COMMON_ALREADY_INITIALIZED;
        if (!model_) return GENIEX_ERROR_COMMON_NOT_INITIALIZED;
        cancelled_.store(false);
        phonon::Stats stats;
        uint64_t      delivered = 0;
        std::string   voice     = optional(input->voice);
        if (voice.empty()) voice = model_->default_voice();
        auto voices = model_->voices();
        if (std::find(voices.begin(), voices.end(), voice) == voices.end()) return GENIEX_ERROR_COMMON_INVALID_INPUT;
        try {
            model_->synthesize(
                input->text,
                voice,
                [&](const float* pcm, size_t samples) {
                    if (cancelled_.load()) return false;
                    delivered += samples;
                    bool more =
                        input->on_audio(pcm, static_cast<uint32_t>(samples), model_->sample_rate(), input->user_data);
                    if (!more) cancelled_.store(true);
                    return more && !cancelled_.load();
                },
                &stats,
                [&] { return cancelled_.load(); },
                false);
        } catch (...) {
            // A failed graph may have changed its cache/profile counters. Fail
            // closed rather than reuse an ambiguous session on the next call.
            model_.reset();
            throw;
        }
        output->sample_rate    = model_->sample_rate();
        output->samples        = delivered;
        output->first_audio_ms = delivered ? stats.first_audio_ms : -1;
        output->generation_ms  = stats.total_ms;
        output->prefill_calls  = stats.prefill_calls - previous_.prefill_calls;
        output->step_calls     = stats.step_calls - previous_.step_calls;
        output->prefill_npu_ms = stats.prefill_npu_ms - previous_.prefill_npu_ms;
        output->step_npu_ms    = stats.step_npu_ms - previous_.step_npu_ms;
        previous_              = stats;
        if (stats.cancelled || cancelled_.load()) return GENIEX_ERROR_COMMON_CANCELLED;
        if (verify_ && (output->prefill_calls == 0 || output->step_calls == 0 || output->prefill_npu_ms <= 0 ||
                           output->step_npu_ms <= 0)) {
            GENIEX_LOG_ERROR("TTS strict verification failed: missing positive prefill/step NPU profile");
            return GENIEX_ERROR_COMMON_NOT_SUPPORTED;
        }
        return GENIEX_SUCCESS;
    }
};

class PhononPlugin final : public geniex::Plugin {
   public:
    const char* version() override { return "phonon-qnn-tts-v1"; }
    int32_t     get_device_list(const geniex_GetDeviceListInput* input, geniex_GetDeviceListOutput* output) override {
        if (!input || !output) return GENIEX_ERROR_COMMON_INVALID_INPUT;
        *output    = {};
        auto ids   = static_cast<const char**>(std::malloc(sizeof(const char*)));
        auto names = static_cast<const char**>(std::malloc(sizeof(const char*)));
        if (!ids || !names) {
            std::free(ids);
            std::free(names);
            return GENIEX_ERROR_COMMON_MEMORY_ALLOCATION;
        }
        ids[0]               = "npu";
        names[0]             = "Qualcomm Hexagon NPU (Phonon QNN HTP)";
        output->device_ids   = ids;
        output->device_names = names;
        output->device_count = 1;
        return GENIEX_SUCCESS;
    }
};
}  // namespace

extern "C" PLUGIN_API geniex_PluginId plugin_id() { return "phonon_qnn"; }
extern "C" PLUGIN_API geniex::Plugin* create_plugin() { return new PhononPlugin(); }
extern "C" PLUGIN_API geniex::ITts* geniex_create_tts_v1() { return new PhononTts(); }
