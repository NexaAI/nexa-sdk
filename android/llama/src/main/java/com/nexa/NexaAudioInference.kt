package com.nexa
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn

class NexaAudioInference(
    private val modelPath: String,
    private val projectorPath: String,
    private var imagePath: String,
    private var stopWords: List<String> = emptyList(),
    private var temperature: Float = 0.8f,
    private var maxNewTokens: Int = 64,
    private var topK: Int = 40,
    private var topP: Float = 0.95f
) {
    init {
        System.loadLibrary("audio-android")
    }

    private var ctxParamsPointer: Long = 0
    private var ctxPointer: Long = 0
    private var generatedTokenNum: Int = 0
    private var generatedText: String = ""
    private var isModelLoaded: Boolean = false

    private external fun init_ctx_params( model: String, project: String, audio_path:String): Long
    private external fun init_ctx(ctxParamsPointer: Long): Long
    private external fun free_ctx(ctxPointer: Long)
    private external fun init_npast():Long
    private external fun init_params(ctxParamsPointer: Long): Long
    private external fun init_sampler(ctxPointer:Long, omniParamsPointer: Long, prompt: String, audioPath: String, npastPointer: Long): Long
    private external fun sampler(ctxOmniPointer :Long , samplerPointer: Long, npastPointer: Long): String
    private external fun free_sampler(samplerPointer: Long)

    @Synchronized
    fun loadModel() {
        if(isModelLoaded){
            throw RuntimeException("Model is already loaded.")
        }
        try {
            val audiuo_path  = "/storage/emulated/0/Android/data/ai.nexa.app_java/files/jfk.wav"
            ctxParamsPointer = init_ctx_params(modelPath, projectorPath, audiuo_path)
            ctxPointer = init_ctx(ctxParamsPointer)
            isModelLoaded = true
        } catch (e: Exception) {
            println(e)
        } catch (e: UnsatisfiedLinkError) {
            throw RuntimeException("Native method not found: ${e.message}")
        }catch (e:Error){
            println(e)
        }
    }

    fun dispose() {
        if(ctxParamsPointer!=0L){
            ctxParamsPointer = 0;
        }
        if (ctxPointer != 0L) {
            free_ctx(ctxPointer)
            ctxPointer = 0;
        }
    }

    private fun shouldStop(): Boolean {
        if(this.generatedTokenNum >= this.maxNewTokens){
            return true
        }

        return stopWords.any { generatedText.contains(it, ignoreCase = true) }
    }

    private fun resetGeneration() {
        generatedTokenNum = 0
        generatedText = ""
    }

    @Synchronized
    fun createCompletionStream(
        prompt: String,
        imagePath: String? = null,
        stopWords: List<String>? = null,
        temperature: Float? = null,
        maxNewTokens: Int? = null,
        topK: Int? = null,
        topP: Float? = null
    ): Flow<String> = flow {
        if(!isModelLoaded){
            throw RuntimeException("Model is not loaded.")
        }
        resetGeneration()
        val imagePathToUse = imagePath ?: this@NexaAudioInference.imagePath

        val audiuo_path  = "/storage/emulated/0/Android/data/ai.nexa.app_java/files/jfk.wav"
        val npastPointer = init_npast()
        val allParamsPointer = init_params(ctxParamsPointer)
        val sampler = init_sampler(ctxPointer, allParamsPointer, prompt, audiuo_path, npastPointer)

        try {
            while (true) {
                val sampledText = sampler(ctxPointer, sampler, npastPointer)
                generatedTokenNum += 1
                generatedText += sampledText
                if(shouldStop()){
                    break
                }
                emit(sampledText)
            }
        } finally {
            resetGeneration()
            free_sampler(sampler)
        }
    }.flowOn(Dispatchers.IO)
}
