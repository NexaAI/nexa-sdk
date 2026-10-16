// Copyright (c) 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

#include "plugin/Plugin.h"

#include <cstdlib>
#include <exception>

#if defined(__linux__) && !defined(__ANDROID__)
#include <dlfcn.h>

#include <filesystem>
#endif

#include "build_config.h"
#include "llm.h"
#include "logging.h"
#include "version.h"  // GENIEX_QAIRT_VERSION from third-party/geniex-qairt/core/include
#include "vlm.h"

namespace geniex {

#if defined(__linux__) && !defined(__ANDROID__)
// The QNN HTP stubs declare `NEEDED libcdsprpc.so`, but Debian/Ubuntu FastRPC packages ship only
// libcdsprpc.so.1. Without the bare name the stub's absolute-path load fails and QNN falls back to
// a system QAIRT stub of a different version ("Stub lib id mismatch"). When the bare name is not
// resolvable, preload the bundled alias (SONAME libcdsprpc.so -> libcdsprpc.so.1) so the stub's
// dependency is satisfied by SONAME match. Handles are intentionally kept for the process lifetime.
static void ensure_cdsprpc_alias() {
    if (dlopen("libcdsprpc.so", RTLD_NOW | RTLD_GLOBAL)) return;

    Dl_info info{};
    if (!dladdr(reinterpret_cast<void*>(&ensure_cdsprpc_alias), &info) || !info.dli_fname) return;
    const auto alias = std::filesystem::path(info.dli_fname).parent_path() / "libgeniex_cdsprpc_alias.so";
    if (!std::filesystem::exists(alias)) return;

    if (dlopen(alias.c_str(), RTLD_NOW | RTLD_GLOBAL)) {
        GENIEX_LOG_DEBUG("libcdsprpc.so not found; using bundled alias for libcdsprpc.so.1");
    } else {
        GENIEX_LOG_WARN("libcdsprpc.so not found and bundled alias failed to load: {}", dlerror());
    }
}
#endif

class QairtPlugin : public Plugin {
   public:
    QairtPlugin() { GENIEX_LOG_TRACE("creating and initializing qairt plugin"); }

    ~QairtPlugin() override { GENIEX_LOG_TRACE("destroying qairt plugin"); }

    const char* version() override { return GENIEX_QAIRT_VERSION; }

    int32_t get_device_list(const geniex_GetDeviceListInput* input, geniex_GetDeviceListOutput* output) override {
        if (!input || !output) {
            return GENIEX_ERROR_COMMON_INVALID_INPUT;
        }

        auto ids   = (const char**)malloc(sizeof(const char*));
        auto names = (const char**)malloc(sizeof(const char*));
        if (!ids || !names) {
            free(ids);
            free(names);
            return GENIEX_ERROR_COMMON_MEMORY_ALLOCATION;
        }
        ids[0]   = "NPU";
        names[0] = "Qualcomm NPU (QAIRT)";

        output->device_ids   = ids;
        output->device_names = names;
        output->device_count = 1;
        return GENIEX_SUCCESS;
    }

    ILlm* create_llm() override { return new geniex::QairtLlm; }

    IVlm* create_vlm() override { return new geniex::QairtVlm; }
};

}  // namespace geniex

#ifdef GENIEX_STATIC

#else

geniex_PluginId plugin_id() { return geniex::build_config::kPluginIdQairt; }

geniex::Plugin* create_plugin() {
    try {
        GENIEX_LOG_TRACE("creating qairt plugin");
#if defined(__linux__) && !defined(__ANDROID__)
        geniex::ensure_cdsprpc_alias();
#endif
        return new geniex::QairtPlugin;
    } catch (const std::exception& e) {
        GENIEX_LOG_ERROR("failed to create qairt plugin: {}", e.what());
        return nullptr;
    }
}

#endif
