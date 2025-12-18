package ai.nexa.agent.constant

import ai.nexa.agent.BuildConfig

class Configs {
    companion object {
        var USE_UNSAFE_HTTP = BuildConfig.DEBUG

        /**
         * 用于偏移量修正
         */
        val userChatMsgBoxHeight = 50

        val gptOssFinalTag = "final<|message|>"
    }
}