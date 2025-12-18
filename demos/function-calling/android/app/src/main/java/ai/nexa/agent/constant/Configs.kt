package ai.nexa.agent.constant

import ai.nexa.agent.BuildConfig

class Configs {
    companion object {
        var USE_UNSAFE_HTTP = BuildConfig.DEBUG

        /**
         * 用于偏移量修正
         */
        val userChatMsgBoxHeight = 50

        const val gptOssFinalTag = "final<|message|>"
        const val DEFAULT_SERVER_IP = "192.168.0.107:8088"
    }
}