/*
 /** Profiling data structure for LLM/VLM performance metrics */
 typedef struct {
     int64_t ttft_us;            /* Time to first token (us) */
     int64_t total_time_us;      /* Total generation time (us) */
     int64_t prompt_time_us;     /* Prompt processing time (us) */
     int64_t decode_time_us;     /* Token generation time (us) */
     double tokens_per_second;  /* Decoding speed (tokens/sec) */
     int64_t total_tokens;      /* Total tokens generated */
     int64_t prompt_tokens;     /* Number of prompt tokens */
     int64_t generated_tokens;  /* Number of generated tokens */
     const char* stop_reason;  /* Stop reason: "eos", "length", "user", "stop_sequence" */
 } ml_ProfilingData;

 */
import NexaBridge

public struct ProfilingData: CustomStringConvertible {
    public let ttft: Int64
    public let totalTokens: Int64
    public let stopReasen: String
    public let tokensPerSecond: Double
    public let totalTime: Int64
    public let promptTime: Int64
    public let decodeTime: Int64
    public let promptTokens: Int64
    public let generatedTokens: Int64

    public init(
        ttft: Int64 = 0,
        totalTokens: Int64 = 0,
        stopReasen: String = "",
        tokensPerSecond: Double = 0,
        totalTime: Int64 = 0,
        promptTime: Int64 = 0,
        decodeTime: Int64 = 0,
        promptTokens: Int64 = 0,
        generatedTokens: Int64 = 0
    ) {
        self.ttft = ttft
        self.totalTokens = totalTokens
        self.stopReasen = stopReasen
        self.tokensPerSecond = tokensPerSecond
        self.totalTime = totalTime
        self.promptTime = promptTime
        self.decodeTime = decodeTime
        self.promptTokens = promptTokens
        self.generatedTokens = generatedTokens
    }

    init(from cProfillingData: ml_ProfilingData) {
        self.ttft = cProfillingData.ttft_us
        self.totalTokens = cProfillingData.total_tokens
        self.stopReasen = cProfillingData.stop_reason == nil ? "" : String(cString: cProfillingData.stop_reason!)
        self.tokensPerSecond = cProfillingData.tokens_per_second
        self.totalTime = cProfillingData.total_time_us
        self.promptTime = cProfillingData.prompt_time_us
        self.decodeTime = cProfillingData.decode_time_us
        self.promptTokens = cProfillingData.prompt_tokens
        self.generatedTokens = cProfillingData.generated_tokens
    }

    public var description: String {
        """
        TTFT: \(ttft) ms
        Total tokens: \(totalTokens)
        Prompt tokens: \(promptTokens)
        Generated tokens: \(promptTokens)
        Stop reason: \(stopReasen)
        Tokens per second: \(tokensPerSecond) tokens/s
        Total time: \(totalTime)  ms
        Prompt time: \(promptTime) ms
        Decode time: \(decodeTime) ms
        """
    }
}
