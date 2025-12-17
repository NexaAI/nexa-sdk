package ai.nexa.agent.bean

data class GoogleRequestData(
    /**
     * example:"help me add this event to my calendar"
     */
    val text: String? = null,
    val image: String? = null
)