package ai.nexa.agent.bean

import kotlinx.serialization.Serializable

@Serializable
class ResponseBean {
    var meta: String? = null
    var content: MutableList<Content>? = null
    var structuredContent: String? = null
    var isIsError: Boolean = false
        private set

    fun setIsError(isError: Boolean) {
        this.isIsError = isError
    }

    @Serializable
    class Content {
        var type: String? = null
        var text: String? = null
        var annotations: String? = null
        var meta: String? = null
    }
}
