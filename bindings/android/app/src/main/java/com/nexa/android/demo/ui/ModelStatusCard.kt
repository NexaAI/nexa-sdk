package com.nexa.android.demo.ui

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.nexa.android.demo.data.AIModelState

@Composable
fun ModelStatusCard(
    llmState: AIModelState,
    vlmState: AIModelState,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.surfaceVariant
        )
    ) {
        Column(
            modifier = Modifier.padding(16.dp)
        ) {
            Text(
                text = "Model Status",
                style = MaterialTheme.typography.titleMedium
            )
            Spacer(modifier = Modifier.height(8.dp))

            // LLM Status
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("LLM:")
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(getStateText(llmState))
                    Spacer(modifier = Modifier.width(8.dp))
                    when (llmState) {
                        is AIModelState.Loading -> CircularProgressIndicator(modifier = Modifier.size(16.dp))
                        is AIModelState.Ready -> Icon(
                            imageVector = Icons.Default.Check,
                            contentDescription = "Ready",
                            tint = MaterialTheme.colorScheme.primary
                        )
                        is AIModelState.Error -> Icon(
                            imageVector = Icons.Default.Warning,
                            contentDescription = "Error",
                            tint = MaterialTheme.colorScheme.error
                        )
                        else -> {}
                    }
                }
            }

            Spacer(modifier = Modifier.height(4.dp))

            // VLM Status
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("VLM:")
                Row(verticalAlignment = Alignment.CenterVertically) {
                    Text(getStateText(vlmState))
                    Spacer(modifier = Modifier.width(8.dp))
                    when (vlmState) {
                        is AIModelState.Loading -> CircularProgressIndicator(modifier = Modifier.size(16.dp))
                        is AIModelState.Ready -> Icon(
                            imageVector = Icons.Default.Check,
                            contentDescription = "Ready",
                            tint = MaterialTheme.colorScheme.primary
                        )
                        is AIModelState.Error -> Icon(
                            imageVector = Icons.Default.Warning,
                            contentDescription = "Error",
                            tint = MaterialTheme.colorScheme.error
                        )
                        else -> {}
                    }
                }
            }

            // Show error details if any
            if (llmState is AIModelState.Error) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = "LLM Error: ${llmState.message}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.error
                )
            }

            if (vlmState is AIModelState.Error) {
                Spacer(modifier = Modifier.height(4.dp))
                Text(
                    text = "VLM Error: ${vlmState.message}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.error
                )
            }
        }
    }
}

@Composable
fun getStateText(state: AIModelState): String {
    return when (state) {
        is AIModelState.Idle -> "Not initialized"
        is AIModelState.Loading -> "Loading..."
        is AIModelState.Ready -> "Ready"
        is AIModelState.Error -> "Error"
    }
}
