package com.nexa.android.demo.data

/**
 * AI 模型状态
 */
sealed class AIModelState {
    object Idle : AIModelState()
    object Loading : AIModelState()
    object Ready : AIModelState()
    data class Error(val message: String) : AIModelState()
}

/**
 * 生成状态
 */
sealed class GenerationState {
    object Idle : GenerationState()
    object Generating : GenerationState()
    data class Token(val text: String) : GenerationState()
    object Completed : GenerationState()
    data class Error(val message: String) : GenerationState()
}

/**
 * 嵌入结果
 */
data class EmbeddingResult(
    val embeddings: List<List<Float>>,
    val dimension: Int
)

/**
 * 重排序结果
 */
data class RerankResult(
    val documents: List<String>,
    val scores: List<Float>
)
