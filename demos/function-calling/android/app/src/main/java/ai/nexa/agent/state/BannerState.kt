package ai.nexa.agent.state

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.Alignment
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class BannerState {
    var message by mutableStateOf<String?>(null)
    var visible by mutableStateOf(false)
    var alignBottom by mutableStateOf(false)

    fun show(msg: String, duration: Long = 2000L, alignBottom: Boolean = false) {
        message = msg
        visible = true
        this.alignBottom = alignBottom
        CoroutineScope(Dispatchers.Main).launch {
            delay(duration)
            visible = false
        }
    }

    fun dismiss() { visible = false }
}

val LocalBannerState = staticCompositionLocalOf { BannerState() }