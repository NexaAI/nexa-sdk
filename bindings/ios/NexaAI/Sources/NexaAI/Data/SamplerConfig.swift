import Foundation

/*

/** Text generation sampling parameters */
 typedef struct {
     float   temperature;        /* Sampling temperature (0.0-2.0) */
     float   top_p;              /* Nucleus sampling parameter (0.0-1.0) */
     int32_t top_k;              /* Top-k sampling parameter */
     float   min_p;              /* Minimum probability for nucleus sampling */
     float   repetition_penalty; /* Penalty for repeated tokens */
     float   presence_penalty;   /* Penalty for token presence */
     float   frequency_penalty;  /* Penalty for token frequency */
     int32_t seed;               /* Random seed (-1 for random) */
     ml_Path grammar_path;       /* Optional grammar file path */
     const char* grammar_string; /* Optional grammar string (BNF-like format) */
 } ml_SamplerConfig;

*/

/// Text generation sampling parameters
public struct SamplerConfig: Codable {
    public var temperature: Float
    public var topP: Float
    public var topK: Int32
    public var minP: Float
    public var repetitionPenalty: Float
    public var presencePenalty: Float
    public var frequencyPenalty: Float
    public var seed: Int32

    // Handle grammar configuration - prioritize grammarString over grammarPath
    public var grammarPath: String?
    public var grammarString: String?

    public static let `default` = SamplerConfig()

    public init(
        temperature: Float = 0.8,
        topP: Float = 0.95,
        topK: Int32 = 40,
        minP: Float = 0.05,
        repetitionPenalty: Float = 1.0,
        presencePenalty: Float = 0.0,
        frequencyPenalty: Float = 0.0,
        seed: Int32 = 0,
        grammarPath: String? = nil,
        grammarString: String? = nil
    ) {
        self.temperature = temperature
        self.topP = topP
        self.topK = topK
        self.minP = minP
        self.repetitionPenalty = repetitionPenalty
        self.presencePenalty = presencePenalty
        self.frequencyPenalty = frequencyPenalty
        self.seed = seed
        self.grammarPath = grammarPath
        self.grammarString = grammarString
    }
}
