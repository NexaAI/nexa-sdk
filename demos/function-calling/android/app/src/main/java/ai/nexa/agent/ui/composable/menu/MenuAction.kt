package com.nexa.studio.ui.chat.menu

import androidx.compose.ui.graphics.painter.Painter

data class MenuAction(
    val icon: Painter,
    val text: String,
    val onClick: () -> Unit
)