package com.geniex.sdk.bean

/** Formatted LLM prompt and optional grammar metadata produced from the supplied tools. */
data class LlmApplyChatTemplateOutput @JvmOverloads constructor(
    val formattedText: String,
    val chatTemplateGrammar: ChatTemplateGrammar? = null
)
