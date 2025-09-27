package com.nexa.android.demo.ui

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.nexa.sdk.bean.ChatMessage
import com.nexa.android.demo.data.GenerationState

@Composable
fun ChatMessageItem(
    message: ChatMessage,
    isGenerating: Boolean = false
) {
    val isUser = message.role == "user"

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        horizontalArrangement = if (isUser) Arrangement.End else Arrangement.Start
    ) {
        Card(
            modifier = Modifier.widthIn(max = 300.dp),
            colors = CardDefaults.cardColors(
                containerColor = if (isUser) MaterialTheme.colorScheme.primaryContainer else MaterialTheme.colorScheme.surfaceVariant
            )
        ) {
            Column(modifier = Modifier.padding(8.dp)) {
                Text(
                    text = message.content,
                    color = if (isUser) MaterialTheme.colorScheme.onPrimaryContainer else MaterialTheme.colorScheme.onSurfaceVariant
                )
                if (isGenerating) {
                    LinearProgressIndicator(modifier = Modifier.fillMaxWidth())
                }
            }
        }
    }
}
