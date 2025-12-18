package ai.nexa.agent.state

import ai.nexa.agent.bean.ChatUiMessage
import android.se.omapi.Session
import java.io.File


data class ChatUiState(
    val input: String = "",
    val operationState: OperationState = OperationState.DEFAULT,
    val chatMsgList: MutableList<ChatUiMessage> = mutableListOf(),
    val selectedImageFiles: List<File> = emptyList(),
    val selectedAudioFiles: List<File> = emptyList(),
    val resetSession: Boolean = false,
    val serverIpAddress: String = "192.168.1.107:8088"
)

enum class OperationState {
    DEFAULT, WAITING_SEND_RESPONSE
}

sealed class ChatIntent {
    data class InputChanged(val text: String) : ChatIntent()
    data class SendMessage(
        val text: String,
        val imageFiles: List<File> = emptyList(),
        val audioFile: List<File> = emptyList()
    ) : ChatIntent()

    object StopWaitingResponse : ChatIntent()
    data class AddImage(val files: List<File>) : ChatIntent()
    object NewSession : ChatIntent()
    data class ChangeServerIPAddress(val ipAddress: String) : ChatIntent()
    data class RemoveImage(val file: File) : ChatIntent()
}

sealed class ChatSingleEvent {
}
