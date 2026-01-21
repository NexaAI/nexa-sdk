//package com.nexa.demo.utils
//
//import com.nexa.sdk.LlmWrapper
//import com.nexa.sdk.bean.ChatMessage
//import com.nexa.sdk.bean.LlmApplyChatTemplateOutput
//import com.nexa.sdk.jni.LLmExt
//import kotlinx.coroutines.Dispatchers
//import kotlinx.coroutines.withContext
//import kotlin.reflect.full.memberProperties
//import kotlin.reflect.jvm.isAccessible
//
//class BugFixTest {
//
//    companion object {
//        suspend fun applyChatTemplate(
//            lLmExt: LLmExt,
//            llmWrapper: LlmWrapper,
//            messages: Array<ChatMessage>,
//            tools: String?,
//            enableThinking: Boolean
//        ): Result<LlmApplyChatTemplateOutput> =
//            withContext(Dispatchers.IO) {
//                var handle = 0L
//                llmWrapper::class.memberProperties.forEach {
//                    if (it.name == "handle") {
//                        it.isAccessible = true
//                        handle = it.call(llmWrapper) as Long
//                        return@forEach
//                    }
//                }
//                if (handle == 0L) {
//                    return@withContext Result.failure(IllegalStateException("LLM not initialized"))
//                }
//                runCatching {
//                    lLmExt.applyChatTemplateExt(handle, messages, tools, enableThinking)
//                }
//            }
//    }
//}