package com.nexa.demo

import com.nexa.sdk.bean.GenerationConfig
import com.nexa.sdk.bean.SamplerConfig

data class GenerationConfigSample(
    var maxTokens: Int = 32,
    var stopWords: List<String>? = null,
    var stopCount: Int = 0,
    var nPast: Int = 0,
    var topK: Int = 40,
    var topP: Float = 0.95f,
    var temperature: Float = 0.8f,
    var imagePaths: List<String>? = null,
    var imageCount: Int = 0,
    var audioPaths: List<String>? = null,
    var audioCount: Int = 0,
    var systemPrompt: String = "",
    //vlm
    var nPredict: Int = 1024,
    var minProbability: Float = 0.05f,
    var xtcThreshold: Float = 0.10f,
    var xtcProbability: Float = 0.0f,
    var typicalP: Float = 0.0f,
    var penaltyLastN: Float = 1.0f,
    var penaltyPresent: Float = 0.0f,
    var mirostat: String = "Off",
    var seed: Int = -1,
    var jinja: Boolean = true
) {
    fun toGenerationConfig(grammarString:String? = null): GenerationConfig {
        return GenerationConfig(
            maxTokens = this.maxTokens,
            stopWords = this.stopWords?.toTypedArray(),
            stopCount = this.stopCount,
            nPast = this.nPast,
            samplerConfig = SamplerConfig(
                topK = this.topK,
                topP = this.topP,
                temperature = this.temperature,
                repetitionPenalty = this.penaltyLastN,
                presencePenalty = this.penaltyPresent,
                seed = this.seed,
                grammarString = grammarString
//                minProbability = this.minProbability,
//                xtcThreshold = this.xtcThreshold,
//                xtcProbability = this.xtcProbability,
//                typicalP = this.typicalP,
            ),
            imagePaths = this.imagePaths?.toTypedArray(),
            imageCount = this.imageCount,
            audioPaths = this.audioPaths?.toTypedArray(),
            audioCount = this.audioCount,
            // nPredict
            // nPredict = this.nPredict,
        )
    }
}