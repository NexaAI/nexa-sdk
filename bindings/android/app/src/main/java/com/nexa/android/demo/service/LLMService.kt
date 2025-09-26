package com.nexa.android.demo.service

import android.util.Log
import com.nexa.android.demo.data.*
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import com.nexa.sdk.LlmWrapper
import com.nexa.sdk.bean.*

/**
 * LLM Service
 * Provides large language model functionality using Nexa AI SDK
 */
class LLMService {
    
    private var llmWrapper: LlmWrapper? = null
    private var isInitialized = false
    
    /**
     * Initialize LLM model using Nexa AI SDK
     */
    suspend fun initialize(
        modelPath: String,
        tokenizerPath: String?,
        config: ModelConfig
    ): Result<Unit> = withContext(Dispatchers.IO) {
        try {
            Log.d("LLMService", "Initializing LLM model: $modelPath")
            
            val llmCreateInput = LlmCreateInput(
                model_path = modelPath,
                tokenizer_path = tokenizerPath,
                config = config,
                plugin_id = "llama_cpp"
            )
            
            val result = LlmWrapper.builder()
                .llmCreateInput(llmCreateInput)
                .dispatcher(Dispatchers.IO)
                .build()
                .mapCatching {
                    llmWrapper = it
                    Unit
                }
            result.getOrThrow()
            
            isInitialized = true
            Log.d("LLMService", "LLM model initialized successfully")
            Result.success(Unit)
        } catch (e: Exception) {
            Log.e("LLMService", "Failed to initialize LLM model (Ask Gemini)", e)
            Log.e("LLMService", "Model path: $modelPath")
            Log.e("LLMService", "Config: $config")
            Result.failure(e)
        }
    }
    
    /**
     * Generate streaming text using Nexa AI SDK
     */
    fun generateStream(
        prompt: String,
        config: GenerationConfig
    ): Flow<GenerationState> = flow {
        if (!isInitialized) {
            emit(GenerationState.Error("Model not initialized"))
            return@flow
        }
        
        try {
            emit(GenerationState.Generating)
            
            // 清理提示文本，确保 UTF-8 编码正确
            val cleanPrompt = prompt.replace(Regex("[^\\x00-\\x7F]"), "_")
                .replace(Regex("[\\x80-\\xFF]"), "_")
                .trim()
            
            Log.d("LLMService", "Original prompt: '$prompt'")
            Log.d("LLMService", "Cleaned prompt: '$cleanPrompt'")
            
            val generationConfig = GenerationConfig(
                maxTokens = config.maxTokens,
                stopWords = config.stopWords
            )
            
            llmWrapper?.generateStreamFlow(cleanPrompt, generationConfig)
                ?.collect { result ->
                    when (result) {
                        is LlmStreamResult.Token -> {
                            emit(GenerationState.Token(result.text))
                        }
                        is LlmStreamResult.Completed -> {
                            emit(GenerationState.Completed)
                        }
                        is LlmStreamResult.Error -> {
                            emit(GenerationState.Error(result.throwable.message ?: "Unknown error"))
                        }
                    }
                }
            
        } catch (e: Exception) {
            Log.e("LLMService", "Error generating text", e)
            emit(GenerationState.Error(e.message ?: "Generation failed"))
        }
    }
    
    /**
     * Apply chat template using Nexa AI SDK
     */
    suspend fun applyChatTemplate(
        messages: List<ChatMessage>,
        tools: List<String>? = null,
        enableThinking: Boolean = false
    ): Result<String> = withContext(Dispatchers.IO) {
        try {
            if (!isInitialized) {
                return@withContext Result.failure(Exception("Model not initialized"))
            }
            
                val chatMessages = messages.toTypedArray()
            
            val result = llmWrapper?.applyChatTemplate(
                messages = chatMessages,
                tools = tools?.joinToString(","),
                enableThinking = enableThinking
            )?.getOrThrow()
            
                Result.success(result?.formattedText ?: "")
            
        } catch (e: Exception) {
            Log.e("LLMService", "Failed to apply chat template", e)
            Result.failure(e)
        }
    }
    
    /**
     * Release resources
     */
    fun close() {
        try {
            llmWrapper?.close()
            llmWrapper = null
            isInitialized = false
            Log.d("LLMService", "LLM resources released")
        } catch (e: Exception) {
            Log.e("LLMService", "Error releasing resources", e)
        }
    }
    
    /**
     * Check if service is initialized
     */
    fun isReady(): Boolean = isInitialized
}
