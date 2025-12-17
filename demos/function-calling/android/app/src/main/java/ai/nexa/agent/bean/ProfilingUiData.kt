package ai.nexa.agent.bean

data class ProfilingUiData(
    var ttftMs: Double = 0.0,            // Time to first token (ms)
    var promptTimeMs: Double = 0.0,      // Prompt processing time (ms)
    var decodeTimeMs: Double = 0.0,      // Token generation time (ms)
    var promptTokens: Int = 0,           // Number of prompt tokens
    var generatedTokens: Int = 0,        // Number of generated tokens
    val audioDurationMs: Long,   /* Audio duration (ms) */
    val prefillSpeed: Double = 0.0,   /* Prefill speed (tokens/sec) */
    val decodingSpeed: Double,  /* Decoding speed (tokens/sec) */
    val realTimeFactor: Double = 0.0,     /* Real-Time Factor(RTF) (1.0 = real-time, >1.0 = faster, <1.0 = slower) */
    var stopReason: String? = null,
    var usedMem: Long = 0,               // Current heap memory used (bytes)
    var maxMem: Long = 0                 // Max heap memory available (bytes)
    //    var totalTimeMs: Double = 0.0,       // Total generation time (ms)
    //    var tokensPerSecond: Double = 0.0,   // Decoding speed (tokens/sec)
//    var totalTokens: Int = 0,            // Total tokens generated
)
