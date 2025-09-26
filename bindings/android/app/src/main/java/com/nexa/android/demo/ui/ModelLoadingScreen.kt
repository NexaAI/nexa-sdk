package com.nexa.android.demo.ui

import android.content.Intent
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CheckCircle
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.outlined.MailOutline
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.nexa.android.demo.data.AIModelState
import com.nexa.android.demo.utils.ModelFileManager
import com.nexa.android.demo.viewmodel.AIViewModel

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ModelLoadingScreen(
    llmState: AIModelState,
    vlmState: AIModelState,
    aiViewModel: AIViewModel,
    modifier: Modifier = Modifier
) {
    val context = LocalContext.current
    val modelFileManager = remember { ModelFileManager(context) }
    
    // File picker launcher for multiple model files
    val fileLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.StartActivityForResult()
    ) { result ->
        if (result.resultCode == android.app.Activity.RESULT_OK) {
            val uris = mutableListOf<android.net.Uri>()
            result.data?.let { data ->
                // Handle multiple file selection
                if (data.clipData != null) {
                    for (i in 0 until data.clipData!!.itemCount) {
                        data.clipData!!.getItemAt(i).uri?.let { uris.add(it) }
                    }
                } else {
                    data.data?.let { uris.add(it) }
                }
            }
            
            if (uris.isNotEmpty()) {
                modelFileManager.initializeModelsWithURIs(aiViewModel, uris)
            }
        }
    }

    Column(
        modifier = modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        // Header
        Text(
            text = "Model Loading",
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(bottom = 16.dp)
        )

        // Auto-reload info card
        Card(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 16.dp),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.primaryContainer
            )
        ) {
            Column(
                modifier = Modifier.padding(16.dp)
            ) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Icon(
                        imageVector = Icons.Default.Refresh,
                        contentDescription = "Auto-reload",
                        tint = MaterialTheme.colorScheme.primary
                    )
                    Text(
                        text = "Auto-reload Status",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "The app will automatically reload previously selected models on startup if the model files still exist. If no models are auto-loaded, use the 'Select Model Files' button below to choose your models.",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onPrimaryContainer
                )
            }
        }

        // Select model files button
        Button(
            onClick = {
                val intent = Intent(Intent.ACTION_OPEN_DOCUMENT).apply {
                    addCategory(Intent.CATEGORY_OPENABLE)
                    type = "*/*"
                    putExtra(Intent.EXTRA_ALLOW_MULTIPLE, true)
                }
                fileLauncher.launch(intent)
            },
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 24.dp)
        ) {
            Icon(
                imageVector = Icons.Outlined.MailOutline,
                contentDescription = "Select Files"
            )
            Spacer(modifier = Modifier.width(8.dp))
            Text("Select Model Files")
        }

        // Model status cards
        LazyColumn(
            verticalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            item {
                ModelStatusCard(
                    title = "LLM Model",
                    state = llmState,
                    modelType = "Large Language Model"
                )
            }
            
            item {
                ModelStatusCard(
                    title = "VLM Model", 
                    state = vlmState,
                    modelType = "Vision Language Model"
                )
            }
        }
    }
}

@Composable
private fun ModelStatusCard(
    title: String,
    state: AIModelState,
    modelType: String,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.fillMaxWidth()
            ) {
                // Status icon
                Icon(
                    imageVector = when (state) {
                        is AIModelState.Loading -> Icons.Default.Refresh
                        is AIModelState.Ready -> Icons.Default.CheckCircle
                        is AIModelState.Error -> Icons.Default.Warning
                        else -> Icons.Default.CheckCircle
                    },
                    contentDescription = "Status",
                    tint = when (state) {
                        is AIModelState.Loading -> MaterialTheme.colorScheme.primary
                        is AIModelState.Ready -> MaterialTheme.colorScheme.primary
                        is AIModelState.Error -> MaterialTheme.colorScheme.error
                        else -> MaterialTheme.colorScheme.onSurface
                    }
                )
                
                Spacer(modifier = Modifier.width(12.dp))
                
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = title,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold
                    )
                    Text(
                        text = modelType,
                        style = MaterialTheme.typography.bodyMedium,
                        color = MaterialTheme.colorScheme.onSurfaceVariant
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            
            // Status message
            Text(
                text = when (state) {
                    is AIModelState.Loading -> "Loading model..."
                    is AIModelState.Ready -> "Model ready for use"
                    is AIModelState.Error -> "Error: ${state.message}"
                    else -> "Unknown state"
                },
                style = MaterialTheme.typography.bodyMedium,
                color = when (state) {
                    is AIModelState.Loading -> MaterialTheme.colorScheme.primary
                    is AIModelState.Ready -> MaterialTheme.colorScheme.primary
                    is AIModelState.Error -> MaterialTheme.colorScheme.error
                    else -> MaterialTheme.colorScheme.onSurface
                }
            )
            
            // Loading progress
            if (state is AIModelState.Loading) {
                Spacer(modifier = Modifier.height(8.dp))
                LinearProgressIndicator(
                    modifier = Modifier.fillMaxWidth()
                )
            }
        }
    }
}
