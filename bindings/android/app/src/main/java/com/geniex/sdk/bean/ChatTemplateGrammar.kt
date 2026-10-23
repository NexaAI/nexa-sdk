// Copyright (c) 2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package com.geniex.sdk.bean

/** Grammar metadata generated while applying a chat template. */
data class ChatTemplateGrammar(
    val grammarType: Int,
    val grammar: String?,
    val lazy: Boolean,
    val generationPrompt: String?,
    val triggers: Array<GrammarTrigger> = emptyArray(),
    val preservedTokens: IntArray = intArrayOf()
)
