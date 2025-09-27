package com.nexa.android.demo.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import com.nexa.android.demo.data.GenerationState
import com.nexa.sdk.bean.ChatMessage

@Composable
fun ChatScreen(
    chatMessages: List<ChatMessage>,
    currentText: String,
    generationState: GenerationState,
    inputText: String,
    onInputTextChange: (String) -> Unit,
    onSendMessage: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .imePadding()
    ) {
        // Chat message list
        LazyColumn(
            modifier = Modifier
                .weight(1f)
                .padding(8.dp),
            reverseLayout = true
        ) {
            // Current generating message
            if (currentText.isNotEmpty()) {
                item {
                    ChatMessageItem(
                        message = ChatMessage(
                            role = "assistant",
                            content = currentText
                        ),
                        isGenerating = generationState is GenerationState.Generating
                    )
                }
            }
            items(chatMessages.reversed()) { message ->
                ChatMessageItem(message = message)
            }
        }

        // Input field and send button
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = inputText,
                onValueChange = onInputTextChange,
                label = { Text("Enter message") },
                modifier = Modifier.weight(1f),
                keyboardOptions = KeyboardOptions.Default.copy(imeAction = ImeAction.Send),
                keyboardActions = KeyboardActions(onSend = { onSendMessage() })
            )

            Spacer(modifier = Modifier.width(8.dp))

            FloatingActionButton(
                onClick = onSendMessage,
                modifier = Modifier.size(48.dp)
            ) {
                Icon(
                    imageVector = Icons.Default.Send,
                    contentDescription = "Send"
                )
            }
        }
    }
}
