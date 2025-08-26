/*
 /** Profiling data structure for LLM/VLM performance metrics */
 typedef struct {
     int64_t ttft;        /* Time to first token (us) */
     int64_t prompt_time; /* Prompt processing time (us) */
     int64_t decode_time; /* Token generation time (us) */

     int64_t prompt_tokens;    /* Number of prompt tokens */
     int64_t generated_tokens; /* Number of generated tokens */
     int64_t audio_duration;   /* Audio duration (us) */

     double prefill_speed;    /* Prefill speed (tokens/sec) */
     double decoding_speed;   /* Decoding speed (tokens/sec) */
     double real_time_factor; /* Real-Time Factor(RTF) (1.0 = real-time, >1.0 = faster, <1.0 = slower) */

     const char* stop_reason; /* Stop reason: "eos", "length", "user", "stop_sequence" */
 } ml_ProfileData;

 */
import NexaBridge

public struct ProfileData: CustomStringConvertible {
    public let ttft: Int64
    public let promptTime: Int64
    public let decodeTime: Int64

    public let promptTokens: Int64
    public let generatedTokens: Int64
    public let audioDuration: Int64

    public let prefillSpeed: Double
    public let decodingSpeed: Double
    public let realTimeFactor: Double

    public let stopReason: String

    public init(
        ttft: Int64 = 0,
        promptTime: Int64 = 0,
        decodeTime: Int64 = 0,
        promptTokens: Int64 = 0,
        generatedTokens: Int64 = 0,
        audioDuration: Int64 = 0,
        prefillSpeed: Double = 0.0,
        decodingSpeed: Double = 0.0,
        realTimeFactor: Double = 0.0,
        stopReason: String = ""
    ) {
        self.ttft = ttft
        self.promptTime = promptTime
        self.decodeTime = decodeTime
        self.promptTokens = promptTokens
        self.generatedTokens = generatedTokens
        self.audioDuration = audioDuration
        self.prefillSpeed = prefillSpeed
        self.decodingSpeed = decodingSpeed
        self.realTimeFactor = realTimeFactor
        self.stopReason = stopReason
    }

    init(from cProfileData: ml_ProfileData) {
        self.ttft = cProfileData.ttft
        self.promptTime = cProfileData.prompt_time
        self.decodeTime = cProfileData.decode_time
        self.promptTokens = cProfileData.prompt_tokens
        self.generatedTokens = cProfileData.generated_tokens
        self.audioDuration = cProfileData.audio_duration
        self.prefillSpeed = cProfileData.prefill_speed
        self.decodingSpeed = cProfileData.decoding_speed
        self.realTimeFactor = cProfileData.real_time_factor
        self.stopReason = cProfileData.stop_reason == nil ? "" : String(cString: cProfileData.stop_reason!)
    }

    public var description: String {
        """
        TTFT: \(ttft) us
        Prompt Time: \(promptTime) us
        Decode Time: \(decodeTime) us
        Prompt Tokens: \(promptTokens)
        Generated Tokens: \(generatedTokens)
        Audio Duration: \(audioDuration) us
        Prefill Speed: \(prefillSpeed) t/s
        Decoding Speed: \(decodingSpeed) t/s
        Real Time Factor: \(realTimeFactor)
        Stop reason: \(stopReason)
        """
    }
}
