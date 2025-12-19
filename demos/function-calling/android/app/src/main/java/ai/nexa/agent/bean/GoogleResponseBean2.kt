package ai.nexa.agent.bean

import kotlinx.serialization.Serializable

@Serializable
class GoogleResponseBean2 {
    var isSuccess: Boolean = false
    var data: Data? = null
    var message: String? = null

    @Serializable
    class Data {
        var htmlLink: String? = null
        var calendarId: String? = null
        var summary: String? = null
        var start: String? = null
        var end: String? = null
        var timeZone: String? = null
    }
}
