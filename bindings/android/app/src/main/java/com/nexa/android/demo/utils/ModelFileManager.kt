package com.nexa.android.demo.utils

import android.content.Context
import android.net.Uri
import android.util.Log
import com.nexa.android.demo.viewmodel.AIViewModel
import com.nexa.sdk.bean.ModelConfig
import java.io.File

/**
 * Manages model file operations using Storage Access Framework
 */
class ModelFileManager(private val context: Context) {

    /**
     * Initialize models with selected file URIs
     */
    fun initializeModelsWithURIs(aiViewModel: AIViewModel, modelURIs: List<Uri>) {
        Log.d("ModelFileManager", "Initializing models with ${modelURIs.size} selected files")
        
        // Find LLM model (look for .gguf files that are not VLM models)
        val llmModelURI = modelURIs.find { uri ->
            val fileName = getFileNameFromURI(uri)
            fileName.contains("Qwen", ignoreCase = true) && fileName.endsWith(".gguf", ignoreCase = true)
        }
        
        if (llmModelURI != null) {
            Log.d("ModelFileManager", "Found LLM model URI: $llmModelURI")
            // Convert URI to file path for SDK
            val llmModelPath = getFilePathFromURI(llmModelURI)
            if (llmModelPath != null) {
                val llmConfig = ModelConfig(
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
                aiViewModel.initializeLLM(llmModelPath, null, llmConfig)
            }
        }
        
        // Find VLM model and mmproj
        val vlmModelURI = modelURIs.find { uri ->
            val fileName = getFileNameFromURI(uri)
            fileName.contains("SmolVLM", ignoreCase = true) && fileName.endsWith(".gguf", ignoreCase = true)
        }
        
        val mmprojURI = modelURIs.find { uri ->
            val fileName = getFileNameFromURI(uri)
            fileName.contains("mmproj", ignoreCase = true) && fileName.endsWith(".json", ignoreCase = true)
        }
        
        if (vlmModelURI != null && mmprojURI != null) {
            Log.d("ModelFileManager", "Found VLM model URI: $vlmModelURI")
            Log.d("ModelFileManager", "Found mmproj URI: $mmprojURI")
            
            val vlmModelPath = getFilePathFromURI(vlmModelURI)
            val mmprojPath = getFilePathFromURI(mmprojURI)
            
            if (vlmModelPath != null && mmprojPath != null) {
                val vlmConfig = ModelConfig(
                    nCtx = 1024,       // 使用参考工程的配置
                    max_tokens = 2048,  // 使用参考工程的配置
                    nThreads = 4,      // 使用参考工程的配置
                    nThreadsBatch = 4, // 使用参考工程的配置
                    nBatch = 1,        // 使用参考工程的配置
                    nUBatch = 1,       // 使用参考工程的配置
                    nSeqMax = 1        // 使用参考工程的配置
                )
                aiViewModel.initializeVLM(vlmModelPath, mmprojPath, vlmConfig)
            }
        }
    }


    /**
     * Get file name from URI
     */
    private fun getFileNameFromURI(uri: Uri): String {
        return try {
            val cursor = context.contentResolver.query(uri, null, null, null, null)
            cursor?.use {
                if (it.moveToFirst()) {
                    val nameIndex = it.getColumnIndex(android.provider.OpenableColumns.DISPLAY_NAME)
                    if (nameIndex >= 0) {
                        it.getString(nameIndex) ?: "unknown"
                    } else "unknown"
                } else "unknown"
            } ?: "unknown"
        } catch (e: Exception) {
            Log.e("ModelFileManager", "Error getting file name from URI: $uri", e)
            "unknown"
        }
    }

    /**
     * Get file path from URI (for SDK compatibility)
     */
    private fun getFilePathFromURI(uri: Uri): String? {
        return try {
            // For SAF URIs, we need to copy to a temporary location that the SDK can access
            val inputStream = context.contentResolver.openInputStream(uri)
            val tempDir = File(context.cacheDir, "models")
            if (!tempDir.exists()) {
                tempDir.mkdirs()
            }
            
            val fileName = getFileNameFromURI(uri)
            val tempFile = File(tempDir, fileName)
            
            inputStream?.use { input ->
                tempFile.outputStream().use { output ->
                    input.copyTo(output)
                }
            }
            
        Log.d("ModelFileManager", "Copied file to: ${tempFile.absolutePath}")
        Log.d("ModelFileManager", "File exists: ${tempFile.exists()}, size: ${tempFile.length()}, readable: ${tempFile.canRead()}")
        
        // Check if file is complete (not empty and has reasonable size)
        if (tempFile.length() == 0L) {
            Log.e("ModelFileManager", "File is empty after copy!")
            return null
        }
        
        // Check file header for GGUF format
        try {
            val header = ByteArray(4)
            tempFile.inputStream().use { it.read(header) }
            val headerString = String(header)
            Log.d("ModelFileManager", "File header: $headerString")
            if (!headerString.startsWith("GGUF")) {
                Log.w("ModelFileManager", "File may not be a valid GGUF file, header: $headerString")
            }
        } catch (e: Exception) {
            Log.w("ModelFileManager", "Could not read file header: ${e.message}")
        }
        
        tempFile.absolutePath
        } catch (e: Exception) {
            Log.e("ModelFileManager", "Error copying file from URI: $uri", e)
            null
        }
    }
}
