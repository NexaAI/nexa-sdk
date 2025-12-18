package com.nexa.studio.ui.chat.menu

import ai.nexa.agent.ui.theme.menuRowIcon
import ai.nexa.agent.ui.theme.menuRowText
import ai.nexa.agent.ui.theme.popupBg
import ai.nexa.agent.ui.theme.popupMask
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.painter.Painter
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalView
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Popup


@Composable
fun MessageMenuPopup(
    show: Boolean,
    onDismiss: () -> Unit,
    offset: Offset = Offset.Zero,
    menuActions: List<MenuAction>
) {
    if (!show) return

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.popupMask)
            .pointerInput(Unit) {
                detectTapGestures { onDismiss() }
            }
    ) {
        val statusBarHeight = statusBarHeightPx()
        Popup(
            alignment = Alignment.TopStart,
            offset = IntOffset(
                x = offset.x.toInt(),
                y = (offset.y - statusBarHeight).toInt()
            ),
            onDismissRequest = { onDismiss() }
        ) {
            Card(
                modifier = Modifier.width(190.dp),
                shape = RoundedCornerShape(5.dp),
                elevation = CardDefaults.cardElevation(defaultElevation = 5.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.popupBg)
            ) {
                Column {
                    menuActions.forEach { action ->
                        MenuRow(
                            icon = action.icon,
                            text = action.text,
                            onClick = {
                                action.onClick()
                                onDismiss()
                            }
                        )
                    }
                }
            }
        }
    }
}

@Composable
fun statusBarHeightPx(): Int {
    val context = LocalView.current.context
    val resId = context.resources.getIdentifier("status_bar_height", "dimen", "android")
    return if (resId > 0) context.resources.getDimensionPixelSize(resId) else 0
}

@Composable
fun MenuRow(icon: Painter, text: String, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() }
            .padding(horizontal = 16.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text,
            color = MaterialTheme.colorScheme.menuRowText,
            modifier = Modifier.weight(1f),
            fontSize = 16.sp
        )
        Icon(
            icon,
            contentDescription = text,
            tint = MaterialTheme.colorScheme.menuRowIcon,
            modifier = Modifier.size(20.dp)
        )
    }
}