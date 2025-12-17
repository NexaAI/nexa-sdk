package ai.nexa.agent.bean

import java.util.UUID

enum class ChatMessageType {
    USER, USER_IMAGE, USER_Audio, BOT, THINK, IN_THINK, Error
}

enum class ChatRole(val role: String) {
    USER("user"), SYSTEM("system"), ASSISTANT("assistant")
}

data class ChatUiMessage(
    val id: String = UUID.randomUUID().toString(),
    val role: String,
    var content: String,
    /**
     * 当消息是 AI 的回复数据时，这里会保留原始数据，因为在 gpt-oss-20b 模型下，回复消息有 template 标签
     */
    var originContent: String? = null,
    val type: ChatMessageType = ChatMessageType.BOT,
    val imageList: MutableList<String> = mutableListOf(),
    val audioList: MutableList<String> = mutableListOf(),
    val profilingData: ProfilingUiData? = null,
    val timestamp: Long = System.currentTimeMillis(),
    val extMsg: Any? = null
)
