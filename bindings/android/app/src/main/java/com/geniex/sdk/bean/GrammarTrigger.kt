package com.geniex.sdk.bean

/** A token, word, or pattern that activates a lazy chat-template grammar. */
data class GrammarTrigger(
    val triggerType: Int,
    val value: String?,
    val token: Int
)
