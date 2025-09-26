package com.nexa.android.demo.ui

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp

@Composable
fun VLMQuestionScreen(
    onSendMessage: (String, String?) -> Unit
) {
    var inputText by remember { mutableStateOf("") }
    var selectedImagePath by remember { mutableStateOf<String?>(null) }

    Column(
        modifier = Modifier.fillMaxSize()
    ) {
        // Image preview area
        if (selectedImagePath != null) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp)
                    .padding(8.dp)
            ) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    Text("Image: ${selectedImagePath}")
                }
            }
        }

        // Input area
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(8.dp)
        ) {
            // Image selection buttons
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                Button(
                    onClick = {
                        // TODO: Implement camera capture
                        selectedImagePath = "/sdcard/Download/test_image.jpg"
                    }
                ) {
                    Text("Take Photo")
                }

                Button(
                    onClick = {
                        // TODO: Implement gallery selection
                        selectedImagePath = "/sdcard/Download/test_image.jpg"
                    }
                ) {
                    Text("Select Image")
                }

                if (selectedImagePath != null) {
                    Button(
                        onClick = { selectedImagePath = null }
                    ) {
                        Text("Remove")
                    }
                }
            }

            Spacer(modifier = Modifier.height(8.dp))

            // Text input
            OutlinedTextField(
                value = inputText,
                onValueChange = { inputText = it },
                label = { Text("Describe the image or ask a question") },
                modifier = Modifier.fillMaxWidth(),
                keyboardOptions = KeyboardOptions.Default.copy(imeAction = ImeAction.Send),
                keyboardActions = KeyboardActions(onSend = {
                    if (inputText.isNotEmpty() || selectedImagePath != null) {
                        onSendMessage(inputText, selectedImagePath)
                        inputText = ""
                        selectedImagePath = null
                    }
                })
            )

            Spacer(modifier = Modifier.height(8.dp))

            // Send button
            Button(
                onClick = {
                    if (inputText.isNotEmpty() || selectedImagePath != null) {
                        onSendMessage(inputText, selectedImagePath)
                        inputText = ""
                        selectedImagePath = null
                    }
                },
                modifier = Modifier.fillMaxWidth(),
                enabled = inputText.isNotEmpty() || selectedImagePath != null
            ) {
                Text("Send Message")
            }
        }
    }
}
