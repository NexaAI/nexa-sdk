package ai.nexa.agent.bean

data class GoogleRequestData2(
    /**
     * example:"help me add this event to my calendar"
     */
    val query: String? = null,
    val base64: String? = null
)