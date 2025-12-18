package ai.nexa.agent.state

data class CommonSettingState(
    val theme: String = "auto",
    val isThinking: Boolean = true,
    val autoOffload: Boolean = true,
    val nCtx: Int = 1024,
    val loading: Boolean = true
)

sealed class CommonSettingIntent {
    object RestCommonSetting : CommonSettingIntent()
    object LoadTheme : CommonSettingIntent()
    data class SetTheme(val theme: String) : CommonSettingIntent()
    data class SetIsThinking(val value: Boolean) : CommonSettingIntent()
    data class SetAutoOffLoad(val value: Boolean) : CommonSettingIntent()
    data class SetNCtx(val value: Int) : CommonSettingIntent()
}

sealed class CommonSettingEvent {
    object Error : CommonSettingEvent()
}