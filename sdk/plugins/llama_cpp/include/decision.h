// Copyright (c) 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

#pragma once

#include <string>
#include <vector>

#include "geniex.h"
#include "llama.h"

namespace geniex {

struct DecisionPrepared {
    std::string              prompt;
    std::string              reversed_prompt;
    std::vector<std::string> candidates;
    std::vector<size_t>      original_index;
    size_t                   true_index = 0;
    std::string              type;
    std::string              question_type;
    float                    temperature = 1.0f;
};

int32_t prepare_decision(const llama_model* model, const geniex_DecideInput* input, DecisionPrepared* output);

}  // namespace geniex
