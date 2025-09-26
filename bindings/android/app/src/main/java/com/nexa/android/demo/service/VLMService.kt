package com.nexa.android.demo.service

import android.util.Log
import com.nexa.android.demo.data.*
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import com.nexa.sdk.VlmWrapper
import com.nexa.sdk.bean.*
import java.io.File

/**
 * VLM Service
 * Provides vision-language model functionality using Nexa AI SDK
 */
class VLMService {
    
    private var vlmWrapper: VlmWrapper? = null
    private var isInitialized = false
    
    /**
     * Initialize VLM model using Nexa AI SDK
     */
    suspend fun initialize(
        modelPath: String,
        mmprojPath: String,
        config: ModelConfig
    ): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            Log.d("VLMService", "Initializing VLM model: $modelPath")
            
                val vlmCreateInput = VlmCreateInput(
                    model_path = modelPath,
                    mmproj_path = mmprojPath,
                    tokenizer_path = null,
                    config = config,
                    plugin_id = "llama_cpp"
                )
            
            val result = VlmWrapper.builder()
                .vlmCreateInput(vlmCreateInput)
                .build()
                .mapCatching {
                    vlmWrapper = it
                    Unit
                }
            result.getOrThrow()
            
            isInitialized = true
            Log.d("VLMService", "VLM model initialized successfully")
            Log.d("VLMService", "VLM wrapper: ${vlmWrapper != null}")
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e("VLMService", "Failed to initialize VLM model (Ask Gemini)", e)
            Log.e("VLMService", "Model path: $modelPath")
            Log.e("VLMService", "mmproj path: $mmprojPath")
            Log.e("VLMService", "Config: $config")
            Result.failure(e)
        }
    }
    
    /**
     * Generate streaming visual question answering using Nexa AI SDK
     */
    fun generateStream(
        messages: List<VlmChatMessage>,
        config: GenerationConfig
    ): Flow<GenerationState> = flow {
        if (!isInitialized) {
            emit(GenerationState.Error("Model not initialized"))
            return@flow
        }
        
        try {
            emit(GenerationState.Generating)
            
            Log.d("VLMService", "Processing visual Q&A, message count: ${messages.size}")
            
            // Log message details for debugging
            messages.forEachIndexed { index, message ->
                Log.d("VLMService", "Message $index: role=${message.role}, contents=${message.contents.size}")
                message.contents.forEach { content ->
                    Log.d("VLMService", "  Content type: ${content.type}, text length: ${content.text?.length ?: 0}")
                    if (content.type == "image" && content.text != null) {
                        // 验证图片路径
                        val imageFile = File(content.text)
                        Log.d("VLMService", "  Image file exists: ${imageFile.exists()}, readable: ${imageFile.canRead()}")
                        if (!imageFile.exists()) {
                            Log.e("VLMService", "Image file does not exist: ${content.text}")
                            emit(GenerationState.Error("Image file not found: ${content.text}"))
                            return@flow
                        }
                    }
                }
            }
            
            val vlmMessages = messages.toTypedArray()
            
            val formattedPrompt = vlmWrapper?.applyChatTemplate(
                vlmMessages, 
                null, 
                true  // 启用 thinking 模式，与参考工程一致
            )?.getOrThrow()
            
            Log.d("VLMService", "Formatted prompt length: ${formattedPrompt?.formattedText?.length ?: 0}")
            Log.d("VLMService", "Formatted prompt preview: ${formattedPrompt?.formattedText?.take(100) ?: "null"}")
            
            // 清理格式化后的提示文本，确保 UTF-8 编码正确
            val cleanFormattedText = formattedPrompt?.formattedText
                ?.replace(Regex("[^\\x00-\\x7F]"), "_")
                ?.replace(Regex("[\\x80-\\xFF]"), "_")
                ?.trim() ?: ""
            
            Log.d("VLMService", "Cleaned formatted text length: ${cleanFormattedText.length}")
            Log.d("VLMService", "Cleaned formatted text preview: ${cleanFormattedText.take(100)}")
            
            Log.d("VLMService", "About to call generateStreamFlow")
            Log.d("VLMService", "VLM wrapper is null: ${vlmWrapper == null}")
            Log.d("VLMService", "Clean formatted text: '$cleanFormattedText'")
            
            if (vlmWrapper == null) {
                Log.e("VLMService", "VLM wrapper is null, cannot generate")
                emit(GenerationState.Error("VLM not initialized"))
                return@flow
            }
            
            vlmWrapper?.generateStreamFlow(
                cleanFormattedText, 
                GenerationConfig(
                    maxTokens = config.maxTokens,
                    stopWords = config.stopWords
                )
            )?.collect { result ->
                when (result) {
                    is LlmStreamResult.Token -> {
                        Log.d("VLMService", "Generated token: '${result.text}'")
                        emit(GenerationState.Token(result.text))
                    }
                    is LlmStreamResult.Completed -> {
                        Log.d("VLMService", "Generation completed successfully")
                        emit(GenerationState.Completed)
                    }
                    is LlmStreamResult.Error -> {
                        Log.e("VLMService", "Generation error: ${result.throwable.message}")
                        emit(GenerationState.Error(result.throwable.message ?: "Unknown error"))
                    }
                }
            }
            
        } catch (e: Exception) {
            Log.e("VLMService", "Error generating visual Q&A", e)
            emit(GenerationState.Error(e.message ?: "Generation failed"))
        }
    }
    
    /**
     * Apply chat template using Nexa AI SDK
     */
    suspend fun applyChatTemplate(
        messages: List<VlmChatMessage>,
        tools: List<String>? = null,
        enableThinking: Boolean = false
    ): Result<String> = withContext(Dispatchers.IO) {
        try {
            if (!isInitialized) {
                return@withContext Result.failure(Exception("Model not initialized"))
            }
            
            Log.d("VLMService", "Applying chat template")
            
            val vlmMessages = messages.toTypedArray()
            
            val result = vlmWrapper?.applyChatTemplate(
                vlmMessages,
                tools?.joinToString(","),
                enableThinking
            )?.getOrThrow()
            
            Result.success(result?.formattedText ?: "")
            
        } catch (e: Exception) {
            Log.e("VLMService", "Failed to apply chat template", e)
            Result.failure(e)
        }
    }
    
    /**
     * Release resources
     */
    fun close() {
        try {
            vlmWrapper?.close()
            vlmWrapper = null
            isInitialized = false
            Log.d("VLMService", "VLM resources released")
        } catch (e: Exception) {
            Log.e("VLMService", "Error releasing resources", e)
        }
    }
    
    /**
     * Check if service is initialized
     */
    fun isReady(): Boolean = isInitialized
}
