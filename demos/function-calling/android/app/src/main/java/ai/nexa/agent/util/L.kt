package ai.nexa.agent.util

import android.util.Log

/**
 * TODO: 统一使用这个 class 打印 log，方便后面拓展维护
 */
class L {
    companion object {
        fun d(tag: String, msg: String) {
            Log.d(tag, msg)
        }

        fun e(tag: String, msg: String) {
            Log.e(tag, msg)
        }

        fun i(tag: String, msg: String) {
            Log.i(tag, msg)
        }

        fun w(tag: String, msg: String) {
            Log.w(tag, msg)
        }
    }
}