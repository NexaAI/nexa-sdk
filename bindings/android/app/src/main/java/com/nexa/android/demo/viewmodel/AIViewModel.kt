package com.nexa.android.demo.viewmodel

import android.content.Context
import android.content.SharedPreferences
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.nexa.android.demo.data.*
import com.nexa.sdk.bean.ChatMessage
import com.nexa.sdk.bean.VlmChatMessage
import com.nexa.sdk.bean.ModelConfig
import com.nexa.sdk.bean.GenerationConfig
import com.nexa.android.demo.service.LLMService
import com.nexa.android.demo.service.VLMService
import kotlinx.coroutines.flow.*
import kotlinx.coroutines.launch
import android.util.Log
import com.nexa.sdk.bean.VlmContent
import kotlinx.coroutines.Dispatchers

/**
 * AI 功能 ViewModel
 * 只包含 LLM 和 VLM 功能
 */
class AIViewModel : ViewModel() {
    
    // 服务实例
    private val llmService = LLMService()
    private val vlmService = VLMService()
    
    // SharedPreferences for model state persistence
    private var sharedPreferences: SharedPreferences? = null
    
    // 状态管理
    private val _llmState = MutableStateFlow<AIModelState>(AIModelState.Idle)
    val llmState: StateFlow<AIModelState> = _llmState.asStateFlow()
    
    private val _vlmState = MutableStateFlow<AIModelState>(AIModelState.Idle)
    val vlmState: StateFlow<AIModelState> = _vlmState.asStateFlow()
    
    // LLM 生成状态
    private val _llmGenerationState = MutableStateFlow<GenerationState>(GenerationState.Idle)
    val llmGenerationState: StateFlow<GenerationState> = _llmGenerationState.asStateFlow()
    
    // VLM 生成状态
    private val _vlmGenerationState = MutableStateFlow<GenerationState>(GenerationState.Idle)
    val vlmGenerationState: StateFlow<GenerationState> = _vlmGenerationState.asStateFlow()
    
    // 聊天消息
    private val _chatMessages = MutableStateFlow<List<ChatMessage>>(emptyList())
    val chatMessages: StateFlow<List<ChatMessage>> = _chatMessages.asStateFlow()
    
    // VLM 聊天消息
    private val _vlmChatMessages = MutableStateFlow<List<VlmChatMessage>>(emptyList())
    val vlmChatMessages: StateFlow<List<VlmChatMessage>> = _vlmChatMessages.asStateFlow()
    
    // LLM 当前生成的文本
    private val _llmCurrentText = MutableStateFlow("")
    val llmCurrentText: StateFlow<String> = _llmCurrentText.asStateFlow()
    
    // VLM 当前生成的文本
    private val _vlmCurrentText = MutableStateFlow("")
    val vlmCurrentText: StateFlow<String> = _vlmCurrentText.asStateFlow()
    
    /**
     * 初始化 ViewModel 并设置 SharedPreferences
     */
    fun initialize(context: Context) {
        sharedPreferences = context.getSharedPreferences("nexa_ai_models", Context.MODE_PRIVATE)
        // 尝试自动重新加载之前保存的模型
        autoReloadModels()
    }
    
    /**
     * 自动重新加载之前保存的模型
     */
    private fun autoReloadModels() {
        val (llmPath, vlmPath, mmprojPath) = getSavedModelPaths()
        
        Log.d("AIViewModel", "Auto-reload check:")
        Log.d("AIViewModel", "  LLM path: $llmPath")
        Log.d("AIViewModel", "  VLM path: $vlmPath")
        Log.d("AIViewModel", "  mmproj path: $mmprojPath")
        
        if (llmPath != null) {
            val llmFile = java.io.File(llmPath)
            Log.d("AIViewModel", "  LLM file exists: ${llmFile.exists()}, readable: ${llmFile.canRead()}")
            if (llmFile.exists()) {
                Log.d("AIViewModel", "Auto-reloading LLM model: $llmPath")
                val config = ModelConfig(
                    nCtx = 1024,
                    max_tokens = 2048,
                    nThreads = 4,
                    nThreadsBatch = 4,
                    nBatch = 1,
                    nUBatch = 1,
                    nSeqMax = 1,
                    nGpuLayers = 0,
                    config_file_path = "",
                    verbose = true
                )
                initializeLLM(llmPath, null, config)
            } else {
                Log.w("AIViewModel", "LLM model file not found: $llmPath")
                // 清除无效的路径
                clearModelStates()
            }
        } else {
            Log.d("AIViewModel", "No saved LLM path found")
        }
        
        if (vlmPath != null && mmprojPath != null) {
            val vlmFile = java.io.File(vlmPath)
            val mmprojFile = java.io.File(mmprojPath)
            Log.d("AIViewModel", "  VLM file exists: ${vlmFile.exists()}, readable: ${vlmFile.canRead()}")
            Log.d("AIViewModel", "  mmproj file exists: ${mmprojFile.exists()}, readable: ${mmprojFile.canRead()}")
            
            if (vlmFile.exists() && mmprojFile.exists()) {
                Log.d("AIViewModel", "Auto-reloading VLM model: $vlmPath, $mmprojPath")
                val config = ModelConfig(
                    nCtx = 1024,       // 使用参考工程的配置
                    max_tokens = 2048,  // 使用参考工程的配置
                    nThreads = 4,      // 使用参考工程的配置
                    nThreadsBatch = 4, // 使用参考工程的配置
                    nBatch = 1,        // 使用参考工程的配置
                    nUBatch = 1,       // 使用参考工程的配置
                    nSeqMax = 1        // 使用参考工程的配置
                )
                initializeVLM(vlmPath, mmprojPath, config)
            } else {
                Log.w("AIViewModel", "VLM model files not found: vlm=$vlmPath, mmproj=$mmprojPath")
                // 清除无效的路径
                clearModelStates()
            }
        } else {
            Log.d("AIViewModel", "No saved VLM paths found: vlm=$vlmPath, mmproj=$mmprojPath")
        }
    }
    
    /**
     * 保存模型状态和路径到 SharedPreferences
     */
    private fun saveModelStates() {
        sharedPreferences?.edit()?.apply {
            putBoolean("llm_ready", _llmState.value is AIModelState.Ready)
            putBoolean("vlm_ready", _vlmState.value is AIModelState.Ready)
            apply()
        }
    }
    
    /**
     * 保存模型路径
     */
    private fun saveModelPaths(llmPath: String? = null, vlmPath: String? = null, mmprojPath: String? = null) {
        sharedPreferences?.edit()?.apply {
            llmPath?.let { putString("llm_path", it) }
            vlmPath?.let { putString("vlm_path", it) }
            mmprojPath?.let { putString("mmproj_path", it) }
            apply()
        }
    }
    
    /**
     * 获取保存的模型路径
     */
    private fun getSavedModelPaths(): Triple<String?, String?, String?> {
        return Triple(
            sharedPreferences?.getString("llm_path", null),
            sharedPreferences?.getString("vlm_path", null),
            sharedPreferences?.getString("mmproj_path", null)
        )
    }
    
    /**
     * 清除保存的模型路径和状态
     */
    fun clearModelStates() {
        sharedPreferences?.edit()?.apply {
            remove("llm_path")
            remove("vlm_path")
            remove("mmproj_path")
            remove("llm_ready")
            remove("vlm_ready")
            apply()
        }
        _llmState.value = AIModelState.Idle
        _vlmState.value = AIModelState.Idle
    }
    
    /**
     * 初始化 LLM 模型
     */
    fun initializeLLM(modelPath: String, tokenizerPath: String?, config: ModelConfig) {
        viewModelScope.launch {
            _llmState.value = AIModelState.Loading
            val result = llmService.initialize(modelPath, tokenizerPath, config)
            if (result.isSuccess) {
                _llmState.value = AIModelState.Ready
                saveModelPaths(llmPath = modelPath) // Save model path
                saveModelStates() // Save state when successful
            } else {
                _llmState.value = AIModelState.Error(result.exceptionOrNull()?.message ?: "初始化失败")
            }
        }
    }
    
    /**
     * 初始化 VLM 模型
     */
    fun initializeVLM(modelPath: String, mmprojPath: String, config: ModelConfig) {
        viewModelScope.launch {
            _vlmState.value = AIModelState.Loading
            val result = vlmService.initialize(modelPath, mmprojPath, config)
            if (result.isSuccess) {
                _vlmState.value = AIModelState.Ready
                saveModelPaths(vlmPath = modelPath, mmprojPath = mmprojPath) // Save model paths
                saveModelStates() // Save state when successful
            } else {
                _vlmState.value = AIModelState.Error(result.exceptionOrNull()?.message ?: "初始化失败")
            }
        }
    }
    
    /**
     * 生成文本
     */
    fun generateText(prompt: String, config: com.nexa.sdk.bean.GenerationConfig) {
        viewModelScope.launch {
            _llmGenerationState.value = GenerationState.Generating
            _llmCurrentText.value = ""
            
            llmService.generateStream(prompt, config).collect { state ->
                when (state) {
                    is GenerationState.Token -> _llmCurrentText.value += state.text
                    is GenerationState.Completed -> {
                        _llmGenerationState.value = GenerationState.Completed
                        // 将生成的回复添加到聊天记录
                        if (_llmCurrentText.value.isNotEmpty()) {
                            val assistantMessage = ChatMessage(
                                role = "assistant",
                                content = _llmCurrentText.value
                            )
                            _chatMessages.value = _chatMessages.value + assistantMessage
                        }
                        _llmCurrentText.value = ""
                    }
                    is GenerationState.Error -> {
                        _llmGenerationState.value = GenerationState.Error(state.message)
                        Log.e("AIViewModel", "LLM Generation error: ${state.message}")
                    }
                    else -> {}
                }
            }
        }
    }
    
    /**
     * 发送聊天消息并获取回复
     */
    fun sendMessage(content: String) {
        val userMessage = ChatMessage(role = "user", content = content)
        _chatMessages.value = _chatMessages.value + userMessage
        
        // 生成回复
        generateText(content, GenerationConfig(maxTokens = 128)) // 减少 token 限制，让回答更简洁
    }
    
    /**
     * 处理视觉问答
     */
    fun processVisualQuestion(messages: List<VlmChatMessage>, config: GenerationConfig) {
        viewModelScope.launch {
            _vlmGenerationState.value = GenerationState.Generating
            _vlmCurrentText.value = ""
            
            vlmService.generateStream(messages, config).collect { state ->
                when (state) {
                    is GenerationState.Token -> _vlmCurrentText.value += state.text
                    is GenerationState.Completed -> {
                        _vlmGenerationState.value = GenerationState.Completed
                        // 将生成的回复添加到聊天记录
                        if (_vlmCurrentText.value.isNotEmpty()) {
                            val assistantMessage = VlmChatMessage(
                                role = "assistant",
                                contents = listOf(VlmContent(type = "text", text = _vlmCurrentText.value))
                            )
                            _vlmChatMessages.value = _vlmChatMessages.value + assistantMessage
                        }
                        _vlmCurrentText.value = ""
                    }
                    is GenerationState.Error -> {
                        _vlmGenerationState.value = GenerationState.Error(state.message)
                        Log.e("AIViewModel", "VLM Generation error: ${state.message}")
                    }
                    else -> {}
                }
            }
        }
    }
    
    /**
     * 发送 VLM 消息（支持图片）
     */
    fun sendVLMMessage(content: String, imagePath: String? = null) {
        viewModelScope.launch {
            // 验证和清理图片路径
            val cleanImagePath = imagePath?.let { path ->
                try {
                    // 检查文件是否存在
                    val file = java.io.File(path)
                    if (!file.exists()) {
                        Log.w("AIViewModel", "Image file does not exist: $path")
                        null
                    } else {
                        // 直接使用原始路径，不进行过度清理
                        Log.d("AIViewModel", "Using image path: $path")
                        path
                    }
                } catch (e: Exception) {
                    Log.e("AIViewModel", "Error processing image path: $path", e)
                    null
                }
            }
            
            // 清理文本内容，确保 UTF-8 编码正确
            val cleanContent = content.replace(Regex("[^\\x00-\\x7F]"), "_")
                .replace(Regex("[\\x80-\\xFF]"), "_")
                .trim()
            
            Log.d("AIViewModel", "Original content: '$content'")
            Log.d("AIViewModel", "Cleaned content: '$cleanContent'")
            
            // 创建用户消息
            val userMessage = if (cleanImagePath != null) {
                // 带图片的消息
                VlmChatMessage(
                    role = "user",
                    contents = listOf(
                        VlmContent(type = "text", text = cleanContent),
                        VlmContent(type = "image", text = cleanImagePath)
                    )
                )
            } else {
                // 纯文本消息
                VlmChatMessage(
                    role = "user",
                    contents = listOf(VlmContent(type = "text", text = cleanContent))
                )
            }
            
            // 添加到聊天记录
            _vlmChatMessages.value = _vlmChatMessages.value + userMessage
            
            // 创建用于处理的消息列表（与参考工程一致，不使用 system prompt）
            val vlmMessages = listOf(userMessage)
            
            // 处理视觉问答
            processVisualQuestion(vlmMessages, GenerationConfig(maxTokens = 32)) // 使用与参考工程一致的 token 限制
        }
    }
    
    /**
     * 清理资源
     */
    override fun onCleared() {
        super.onCleared()
        // 服务会在需要时自动清理资源
    }
}