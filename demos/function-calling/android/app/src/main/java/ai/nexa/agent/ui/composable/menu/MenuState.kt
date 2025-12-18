package com.nexa.studio.ui.chat.menu

import ai.nexa.agent.bean.ChatUiMessage
import androidx.compose.ui.geometry.Offset

data class MenuState(
    val show: Boolean = false,
    val type: String = "",
    val offset: Offset = Offset.Zero,
    val message: ChatUiMessage? = null
)