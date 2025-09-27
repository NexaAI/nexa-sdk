package com.nexa.android.demo.ui

import android.content.Context
import android.net.Uri
import android.util.Log
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.layout.imePadding
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.consumeWindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.unit.dp
import androidx.core.content.FileProvider
import coil.compose.AsyncImage
import com.nexa.android.demo.data.GenerationState
import com.nexa.sdk.bean.VlmChatMessage
import com.nexa.sdk.bean.VlmContent
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream

@Composable
fun VLMChatScreen(
    chatMessages: List<VlmChatMessage>,
    currentText: String,
    generationState: GenerationState,
    inputText: String,
    onInputTextChange: (String) -> Unit,
    onSendMessage: (String, String?) -> Unit
) {
    var selectedImagePath by remember { mutableStateOf<String?>(null) }
    val context = LocalContext.current
    var imageUri by remember { mutableStateOf<Uri?>(null) }

    // Gallery picker launcher
    val galleryLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.GetContent()
    ) { uri: Uri? ->
        uri?.let {
            imageUri = it
            // Copy image to app's internal storage
            val imagePath = copyImageToInternalStorage(context, it)
            selectedImagePath = imagePath
        }
    }

    // Camera launcher
    val cameraLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicture()
    ) { success ->
        if (success) {
            imageUri?.let { uri ->
                val imagePath = copyImageToInternalStorage(context, uri)
                selectedImagePath = imagePath
            }
        }
    }

    // Create temporary file for camera
    val tempImageFile = remember {
        File.createTempFile("temp_image", ".jpg", context.cacheDir)
    }

    // Create FileProvider URI for camera
    val cameraUri = remember(tempImageFile) {
        FileProvider.getUriForFile(
            context,
            "com.nexa.android.demo.fileprovider",
            tempImageFile
        )
    }

    LaunchedEffect(Unit) {
        imageUri = cameraUri
    }

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
                    VLMMessageItem(
                        message = VlmChatMessage(
                            role = "assistant",
                            contents = listOf(VlmContent(type = "text", text = currentText))
                        ),
                        isGenerating = generationState is GenerationState.Generating
                    )
                }
            }
            items(chatMessages.reversed()) { message ->
                VLMMessageItem(message = message)
            }
        }

        // Image preview with border (smaller size)
        if (selectedImagePath != null) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(120.dp)
                    .padding(horizontal = 8.dp, vertical = 4.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                ),
                border = BorderStroke(
                    width = 2.dp,
                    color = MaterialTheme.colorScheme.primary
                )
            ) {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    AsyncImage(
                        model = selectedImagePath,
                        contentDescription = "Selected image",
                        modifier = Modifier
                            .fillMaxSize()
                            .clip(RoundedCornerShape(8.dp)),
                        contentScale = ContentScale.Crop
                    )
                    
                    // Remove button overlay
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopEnd)
                            .padding(4.dp)
                    ) {
                        FloatingActionButton(
                            onClick = { selectedImagePath = null },
                            modifier = Modifier.size(28.dp),
                            containerColor = MaterialTheme.colorScheme.errorContainer,
                            contentColor = MaterialTheme.colorScheme.onErrorContainer
                        ) {
                            Text("×", style = MaterialTheme.typography.bodySmall)
                        }
                    }
                }
            }
        }
        
        // Image selection buttons
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceEvenly
        ) {
            OutlinedButton(
                onClick = {
                    cameraLauncher.launch(cameraUri)
                }
            ) {
                Text("📷 Take Photo")
            }

            OutlinedButton(
                onClick = {
                    galleryLauncher.launch("image/*")
                }
            ) {
                Text("🖼️ Select Image")
            }

        }

        Spacer(modifier = Modifier.height(8.dp))

        // Text input and send button
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = inputText,
                onValueChange = onInputTextChange,
                label = { Text("Describe the image or ask a question") },
                modifier = Modifier.weight(1f),
                keyboardOptions = KeyboardOptions.Default.copy(imeAction = ImeAction.Send),
                keyboardActions = KeyboardActions(onSend = { 
                    if (inputText.isNotEmpty() || selectedImagePath != null) {
                        onSendMessage(inputText, selectedImagePath)
                        selectedImagePath = null // Clear image preview after sending
                    }
                })
            )

            Spacer(modifier = Modifier.width(8.dp))

            FloatingActionButton(
                onClick = {
                    if (inputText.isNotEmpty() || selectedImagePath != null) {
                        onSendMessage(inputText, selectedImagePath)
                        selectedImagePath = null // Clear image preview after sending
                    }
                },
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

private fun copyImageToInternalStorage(context: Context, uri: Uri): String {
    val inputStream: InputStream? = context.contentResolver.openInputStream(uri)
    // 生成安全的文件名，只包含 ASCII 字符
    val fileName = "image_${System.currentTimeMillis()}.jpg"
    val file = File(context.filesDir, fileName)
    
    try {
        inputStream?.use { input ->
            FileOutputStream(file).use { output ->
                input.copyTo(output)
            }
        }
        
        // 验证文件是否成功创建
        if (!file.exists() || file.length() == 0L) {
            throw Exception("Failed to copy image file")
        }
        
        Log.d("VLMChatScreen", "Image copied successfully: ${file.absolutePath}")
        return file.absolutePath
    } catch (e: Exception) {
        Log.e("VLMChatScreen", "Error copying image: ${e.message}")
        throw e
    }
}
