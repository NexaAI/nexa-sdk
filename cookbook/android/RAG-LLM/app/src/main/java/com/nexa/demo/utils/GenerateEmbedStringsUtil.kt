package com.nexa.demo.utils

import android.content.Context
import android.util.Log
import com.nexa.demo.RAGConfig
import com.nexa.demo.bean.EmbedResultBean
import com.nexa.demo.bean.ModelData
import com.nexa.demo.bean.getNexaManifest
import com.nexa.demo.bean.modelDir
import com.nexa.demo.bean.modelFile
import com.nexa.demo.bean.tokenFile
import com.nexa.sdk.EmbedderWrapper
import com.nexa.sdk.bean.EmbedderCreateInput
import com.nexa.sdk.bean.EmbeddingConfig
import com.nexa.sdk.bean.ModelConfig
import java.io.File

object GenerateEmbedStringsUtil {

    private const val TAG = "GenerateEmbedStringsUtil"

    private lateinit var embedderWrapper: EmbedderWrapper

    suspend fun load(context: Context, selectModelData: ModelData) {
        if (::embedderWrapper.isInitialized) {
            return
        }
        val nexaManifestBean = selectModelData.getNexaManifest(context)
        val pluginId = nexaManifestBean?.PluginId ?: "cpu_gpu"
        val embedderCreateInput = EmbedderCreateInput(
            model_name = nexaManifestBean?.ModelName
                ?: "",  // Model name for NPU plugin
            model_path = selectModelData.modelFile(context)!!.absolutePath,
            tokenizer_path = selectModelData.tokenFile(context)?.absolutePath,
            config = ModelConfig(
                npu_lib_folder_path = context.applicationInfo.nativeLibraryDir,
                npu_model_folder_path = selectModelData.modelDir(context).absolutePath,
                nGpuLayers = 999
            ),
            plugin_id = pluginId,
            device_id = null
        )

        EmbedderWrapper.builder()
            .embedderCreateInput(embedderCreateInput)
            .build().onSuccess { wrapper ->
                embedderWrapper = wrapper
            }.onFailure { error ->

            }
    }

    /**
     * Split text into chunks by word count
     */
    private fun chunkText(text: String, chunkSize: Int = RAGConfig.CHUNK_SIZE): List<String> {
        val words = text.split(Regex("\\s+"))
        val chunks = mutableListOf<String>()
        
        var i = 0
        while (i < words.size) {
            val endIndex = minOf(i + chunkSize, words.size)
            val chunk = words.subList(i, endIndex).joinToString(" ")
            chunks.add(chunk)
            i += chunkSize
        }
        
        return chunks
    }

    /**
     * Embed file paths with chunking support
     */
    suspend fun embed(txtFilePaths: Array<String>): ArrayList<EmbedResultBean> {
        val result = arrayListOf<EmbedResultBean>()
        
        // Process each file
        txtFilePaths.forEach { filePath ->
            try {
                // Read file content
                val fileContent = File(filePath).readText()
                
                // Split into chunks
                val chunks = chunkText(fileContent)
                
                // Embed each chunk
                chunks.forEachIndexed { chunkIndex, chunkText ->
                    embedderWrapper.embed(arrayOf(chunkText), EmbeddingConfig()).onSuccess { embeddings ->
                        result.add(
                            EmbedResultBean(
                                path = filePath,
                                txt = chunkText,
                                chunkIndex = chunkIndex,
                                result = embeddings.embeddings,
                                embedResult = embeddings
                            )
                        )
                        Log.d(TAG, "Embedded chunk $chunkIndex from $filePath")
                    }.onFailure { error ->
                        Log.e(TAG, "Failed to embed chunk $chunkIndex from $filePath: $error")
                    }
                }
            } catch (e: Exception) {
                Log.e(TAG, "Failed to read file $filePath: $e")
            }
        }
        
        return result
    }
    
    /**
     * Embed text directly (for queries)
     */
    suspend fun embedText(text: String): FloatArray? {
        var result: FloatArray? = null
        embedderWrapper.embed(arrayOf(text), EmbeddingConfig()).onSuccess { embeddings ->
            result = embeddings.embeddings
        }.onFailure { error ->
            Log.e(TAG, "Failed to embed text: $error")
        }
        return result
    }

    fun computeCosineSimilarity(
        embedding1: FloatArray?,
        embedding2: FloatArray?
    ): Float {
        if (embedding1 == null || embedding2 == null) return 0.0f
        if (embedding1.isEmpty() || embedding2.isEmpty()) return 0.0f
        if (embedding1.size != embedding2.size) return 0.0f

        var dotProduct = 0.0f
        var norm1 = 0.0f
        var norm2 = 0.0f

        for (i in embedding1.indices) {
            dotProduct += embedding1[i] * embedding2[i]
            norm1 += embedding1[i] * embedding1[i]
            norm2 += embedding2[i] * embedding2[i]
        }

        val epsilon = 1e-8f
        norm1 = kotlin.math.sqrt(norm1 + epsilon)
        norm2 = kotlin.math.sqrt(norm2 + epsilon)
        Log.d("nfl", "dotProduct > 0 ? ${dotProduct > 0}")
        return dotProduct / (norm1 * norm2)
    }
}