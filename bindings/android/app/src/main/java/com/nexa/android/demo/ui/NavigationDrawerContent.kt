package com.nexa.android.demo.ui

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp

@Composable
fun NavigationDrawerContent(
    selectedScreen: String,
    onScreenSelected: (String) -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp)
    ) {

        val screens = listOf(
            "Model Loading" to "Load and initialize AI models",
            "LLM Chat" to "Text conversation with LLM",
            "VLM Chat" to "Visual question answering"
        )

        screens.forEach { (screen, description) ->
            NavigationDrawerItem(
                label = { Text(screen) },
                selected = selectedScreen == screen,
                onClick = { onScreenSelected(screen) },
                modifier = Modifier.fillMaxWidth()
            )
            Text(
                text = description,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant,
                modifier = Modifier.padding(start = 16.dp, bottom = 8.dp)
            )
        }
    }
}
