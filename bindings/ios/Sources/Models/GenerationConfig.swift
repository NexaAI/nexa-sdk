import Foundation

/*
LLM / VLM generation configuration (IMPROVED: support multiple images and audios)
typedef struct {
    int32_t          max_tokens;     /* Maximum tokens to generate */
    const char**     stop;           /* Array of stop sequences */
    int32_t          stop_count;     /* Number of stop sequences */
    int32_t          n_past;         /* Number of past tokens to consider */
    ml_SamplerConfig sampler_config; /* Advanced sampling config */
    // --- Improved multimodal support ---
    ml_Path*         image_paths;    /* Array of image paths for VLM (NULL if none) */
    int32_t          image_count;    /* Number of images */
    ml_Path*         audio_paths;    /* Array of audio paths for VLM (NULL if none) */
    int32_t          audio_count;    /* Number of audios */
} ml_GenerationConfig;
*/

/// LLM / VLM generation configuration
public struct GenerationConfig: Codable {
    public var maxTokens: Int32                 // Maximum tokens to generate
    public var stop: [String]                   // Array of stop sequences
    public var nPast: Int32                     // Number of past tokens to consider
    public var samplerConfig: SamplerConfig     // Advanced sampling config
    public var imagePaths: [String]            // Array of image paths for VLM
    public var audioPaths: [String]            // Array of audio paths for VLM

    public init(
        maxTokens: Int32 = 1024,
        stop: [String] = [],
        nPast: Int32 = 0,
        samplerConfig: SamplerConfig = .default,
        imagePaths: [String] = [],
        audioPaths: [String] = []
    ) {
        self.maxTokens = maxTokens
        self.stop = stop
        self.nPast = nPast
        self.samplerConfig = samplerConfig
        self.imagePaths = imagePaths
        self.audioPaths = audioPaths
    }

    public static let `default` = GenerationConfig()
}
