package com.nexa.studio.ui.chat.markdown

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import com.mikepenz.markdown.model.MarkdownState
import com.mikepenz.markdown.model.rememberMarkdownState
import kotlinx.coroutines.delay

/**
 * Optimized Markdown state for streaming content
 * 
 * Strategy:
 * - Short text (<1500 chars): Immediate parsing (sync, no throttle)
 * - Long text (>=1500 chars): Throttled parsing (sync with debounce)
 * 
 * This approach ensures:
 * - No flickering (immediate = true)
 * - Good performance for short text
 * - Optimized performance for long text via throttling
 * 
 * @param content The markdown content to display
 * @param debounceMs Debounce time for long content updates
 * @return MarkdownState ready for rendering
 */
@Composable
fun rememberStreamingMarkdownState(
    content: String,
    debounceMs: Long = 150L
): MarkdownState {
    // Throttled content for long text optimization
    var throttledContent by remember { mutableStateOf(content) }
    
    // Apply throttling only for long content
    val isLongContent = content.length > 1500
    
    LaunchedEffect(content) {
        if (isLongContent) {
            // Long content: throttle updates to reduce parsing frequency
            delay(debounceMs)
            throttledContent = content
        } else {
            // Short content: immediate update
            throttledContent = content
        }
    }
    
    // Always use immediate = true for synchronous parsing (no flickering)
    return rememberMarkdownState(
        content = throttledContent,
        immediate = true
    )
}

