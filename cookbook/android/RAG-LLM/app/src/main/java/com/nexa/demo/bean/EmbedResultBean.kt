package com.nexa.demo.bean

import com.nexa.sdk.bean.EmbedResult

data class EmbedResultBean(
    val path: String,  // Original file path
    val txt: String? = null,  // The chunk text content
    val chunkIndex: Int = 0,  // Index of the chunk in the original file
    val result: FloatArray,  // Embedding vector
    var score: Float = 0f,
    val embedResult: EmbedResult
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as EmbedResultBean

        if (chunkIndex != other.chunkIndex) return false
        if (score != other.score) return false
        if (path != other.path) return false
        if (txt != other.txt) return false
        if (!result.contentEquals(other.result)) return false
        if (embedResult != other.embedResult) return false

        return true
    }

    override fun hashCode(): Int {
        var result1 = chunkIndex
        result1 = 31 * result1 + score.hashCode()
        result1 = 31 * result1 + path.hashCode()
        result1 = 31 * result1 + (txt?.hashCode() ?: 0)
        result1 = 31 * result1 + result.contentHashCode()
        result1 = 31 * result1 + embedResult.hashCode()
        return result1
    }


}
