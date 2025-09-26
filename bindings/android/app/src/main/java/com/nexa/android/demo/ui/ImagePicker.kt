package com.nexa.android.demo.ui

import android.content.Context
import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.Image
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.content.FileProvider
import coil.compose.AsyncImage
import java.io.File
import java.io.FileOutputStream
import java.io.InputStream

@Composable
fun ImagePicker(
    selectedImagePath: String?,
    onImageSelected: (String?) -> Unit
) {
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
            onImageSelected(imagePath)
        }
    }

    // Camera launcher
    val cameraLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.TakePicture()
    ) { success ->
        if (success) {
            imageUri?.let { uri ->
                val imagePath = copyImageToInternalStorage(context, uri)
                onImageSelected(imagePath)
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
        modifier = Modifier.fillMaxWidth()
    ) {
        // Image preview
        if (selectedImagePath != null) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp)
                    .padding(8.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
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

            if (selectedImagePath != null) {
                OutlinedButton(
                    onClick = { onImageSelected(null) }
                ) {
                    Text("❌ Remove")
                }
            }
        }
    }
}

private fun copyImageToInternalStorage(context: Context, uri: Uri): String {
    val inputStream: InputStream? = context.contentResolver.openInputStream(uri)
    val fileName = "image_${System.currentTimeMillis()}.jpg"
    val file = File(context.filesDir, fileName)
    
    inputStream?.use { input ->
        FileOutputStream(file).use { output ->
            input.copyTo(output)
        }
    }
    
    return file.absolutePath
}
