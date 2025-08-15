package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

const modelLoadFailMsg = `⚠️ Oops. Model failed to load.

👉 Try these:
- Verify your system meets the model's requirements.
- Seek help in our discord or slack.`

var (
	// disableStream *bool // reuse in run.go
	ngl          int32
	enableThink  bool
	tool         []string
	prompt       []string
	query        string
	document     []string
	input        string
	output       string
	voice        string
	listVoice    bool
	speechSpeed  float64
	language     string
	listLanguage bool
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{
		Use:   "infer <model-name>",
		Short: "Infer with a model",
		Long:  "Run inference with a specified model. The model must be downloaded and cached locally.",
	}

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	inferCmd.Flags().SortFlags = false
	inferCmd.Flags().Int32VarP(&ngl, "ngl", "n", 999, "[llm|vlm] num of layers pass to gpu")
	inferCmd.Flags().StringArrayVarP(&tool, "tool", "t", nil, "[llm|vlm] add tool to make function call")
	inferCmd.Flags().BoolVarP(&enableThink, "think", "", true, "[llm] Qwen3 enable thinking mode")
	inferCmd.Flags().StringArrayVarP(&prompt, "prompt", "p", nil, "[embedder|tts] pass prompt")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "[reranker] query")
	inferCmd.Flags().StringArrayVarP(&document, "document", "d", nil, "[reranker] documents")
	inferCmd.Flags().StringVarP(&input, "input", "i", "", "[cv] input file (image for cv)")
	inferCmd.Flags().StringVarP(&output, "output", "o", "", "[tts] output file (audio for tts)")
	inferCmd.Flags().StringVarP(&voice, "voice", "", "", "[tts] voice identifier")
	inferCmd.Flags().BoolVarP(&listVoice, "list-voice", "", false, "[tts] list available voices")
	inferCmd.Flags().Float64VarP(&speechSpeed, "speech-speed", "", 1.0, "[tts] speech speed (1.0 = normal)")
	// inferCmd.Flags().StringVarP(&language, "language", "", "", "[asr] language code (e.g., en, zh, ja)")           // TODO: Language support not implemented yet
	// inferCmd.Flags().BoolVarP(&listLanguage, "list-language", "", false, "[asr] list available languages")        // TODO: Language support not implemented yet

	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		// QNN

		s := store.Get()

		var modelName string

		switch args[0] {
		case "qwen3", "paddleocr", "yolov12":
			modelName = "nexaml/qnn-laptop-libs"
		case "omni-neural":
			modelName = "nexaml/qnn-laptop-libs-omni-neural"
		}

		manifest, err := s.GetManifest(modelName)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(render.GetTheme().Info.Sprintf("model not found, start download"))

			pull().Run(cmd, args)
			// check agin
			manifest, err = s.GetManifest(modelName)
		}
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("parse manifest error: %s", err))
			return
		}

		quant := "N/A"
		switch args[0] {
		case "qwen3":
			manifest.ModelType = types.ModelTypeLLM
		case "omni-neural":
			manifest.ModelType = types.ModelTypeVLM
		case "paddleocr":
			manifest.ModelType = types.ModelTypeCV
			manifest.PluginId = "paddleocr"
		case "yolov12":
			manifest.ModelType = types.ModelTypeCV
			manifest.PluginId = "yolov12"
		default:
			fmt.Println(render.GetTheme().Error.Sprintf("not support: %s", args[0]))
			return
		}

		// QNN

		fmt.Println(render.GetTheme().Quant.Sprintf("🔹 Quant=%s", quant))

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)

		switch manifest.ModelType {
		case types.ModelTypeLLM:
			inferLLM(manifest.PluginId, modelfile)
		case types.ModelTypeVLM:
			var mmprojfile string
			if manifest.MMProjFile.Name != "" {
				mmprojfile = s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
			}
			inferVLM(manifest.PluginId, modelfile, mmprojfile)
		case types.ModelTypeEmbedder:
			inferEmbedder(manifest.PluginId, modelfile)
		case types.ModelTypeReranker:
			inferReranker(manifest.PluginId, modelfile)
		case types.ModelTypeTTS:
			inferTTS(manifest.PluginId, modelfile, "")
		case types.ModelTypeASR:
			inferASR(manifest.PluginId, modelfile, "")
		case types.ModelTypeCV:
			inferCV(manifest.PluginId, modelfile)
		case types.ModelTypeImageGen:
			// ImageGen model is a directory, not a file
			inferImageGen(manifest.PluginId, s.ModelfilePath(manifest.Name, ""))
		default:
			panic("not support model type")
		}
	}
	return inferCmd
}

func inferLLM(plugin, modelfile string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
		ModelPath: modelfile,
		PluginID:  plugin,
		Config: nexa_sdk.ModelConfig{
			NCtx:       4096,
			NGpuLayers: ngl,
		},
	})
	spin.Stop()

	if err != nil {
		slog.Error("failed to create LLM", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.LlmChatMessage

	repl(ReplConfig{
		ParseFile: false,

		Reset: p.Reset,

		SaveKVCache: func(path string) error {
			_, err := p.SaveKVCache(nexa_sdk.LlmSaveKVCacheInput{Path: path})
			return err
		},

		LoadKVCache: func(path string) error {
			_, err := p.LoadKVCache(nexa_sdk.LlmLoadKVCacheInput{Path: path})
			return err
		},

		Run: func(prompt string, _, _ []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			history = append(history, nexa_sdk.LlmChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})

			templateOutput, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
				Messages:    history,
				EnableThink: enableThink,
			})
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			res, err := p.Generate(nexa_sdk.LlmGenerateInput{
				PromptUTF8: templateOutput.FormattedText,
				OnToken:    on_token,
				Config: &nexa_sdk.GenerationConfig{
					MaxTokens: 2048,
				},
			},
			)
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			history = append(history, nexa_sdk.LlmChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res.FullText})
			return res.FullText, res.ProfileData, nil
		},
	})
}

func inferVLM(plugin, modelfile string, mmprojfile string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewVLM(nexa_sdk.VlmCreateInput{
		ModelPath:  modelfile,
		MmprojPath: mmprojfile,
		PluginID:   plugin,
		Config: nexa_sdk.ModelConfig{
			NCtx:       4096,
			NGpuLayers: ngl,
		},
	})
	spin.Stop()

	if err != nil {
		slog.Error("failed to create VLM", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.VlmChatMessage

	repl(ReplConfig{
		ParseFile: true,

		Reset: p.Reset,

		SaveKVCache: func(path string) error {
			return fmt.Errorf("VLM does not support KV cache saving")
		},

		LoadKVCache: func(path string) error {
			return fmt.Errorf("VLM does not support KV cache loading")
		},

		Run: func(prompt string, images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			msg := nexa_sdk.VlmChatMessage{Role: nexa_sdk.VlmRoleUser}
			msg.Contents = append(msg.Contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeText, Text: prompt})
			for _, image := range images {
				msg.Contents = append(msg.Contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeImage, Text: image})
			}
			for _, audio := range audios {
				msg.Contents = append(msg.Contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeAudio, Text: audio})
			}

			history = append(history, msg)

			res, err := p.Generate(nexa_sdk.VlmGenerateInput{
				PromptUTF8: prompt,
				OnToken:    on_token,
				Config: &nexa_sdk.GenerationConfig{
					MaxTokens:  2048,
					ImagePaths: images,
					AudioPaths: audios,
				},
			})
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			history = append(history, nexa_sdk.VlmChatMessage{
				Role: nexa_sdk.VlmRoleAssistant,
				Contents: []nexa_sdk.VlmContent{
					{Type: nexa_sdk.VlmContentTypeText, Text: res.FullText},
				},
			})

			return res.FullText, res.ProfileData, nil
		},
	})
}

func inferTTS(plugin, modelfile string, vocoderfile string) {
	spin := render.NewSpinner("loading TTS model...")
	spin.Start()

	ttsInput := nexa_sdk.TtsCreateInput{
		ModelPath:   modelfile,
		VocoderPath: vocoderfile,
		PluginID:    plugin,
	}

	p, err := nexa_sdk.NewTTS(ttsInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create TTS", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	if listVoice {
		voices, err := p.ListAvailableVoices()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Failed to list available voices: %s", err))
			return
		}
		fmt.Println(render.GetTheme().Success.Sprintf("Available voices: %v", voices.VoiceIDs))
		return
	}

	prompts := prompt
	if len(prompts) == 0 {
		fmt.Println(render.GetTheme().Error.Sprintf("text is required for TTS synthesis (use --prompt)"))
		fmt.Println()
		return
	}

	// Check for empty strings in prompts
	for _, p := range prompt {
		if strings.TrimSpace(p) == "" {
			fmt.Println(render.GetTheme().Error.Sprintf("prompt cannot be empty"))
			fmt.Println()
			return
		}
	}

	// Combine all prompt texts
	textToSynthesize := strings.Join(prompts, " ")

	// Generate output filename if not specified
	outputFile := output
	if outputFile == "" {
		outputFile = fmt.Sprintf("tts_output_%d.wav", time.Now().Unix())
	}

	// Create TTS config
	ttsConfig := &nexa_sdk.TTSConfig{
		Voice:      "af_heart",
		Speed:      float32(speechSpeed),
		SampleRate: 24000,
		Seed:       42,
	}

	if voice != "" {
		ttsConfig.Voice = voice
	}

	// Synthesize speech
	synthesizeInput := nexa_sdk.TtsSynthesizeInput{
		TextUTF8:   textToSynthesize,
		Config:     ttsConfig,
		OutputPath: outputFile,
	}

	fmt.Println(render.GetTheme().Success.Sprintf("Synthesizing speech: \"%s\"", textToSynthesize))

	result, err := p.Synthesize(synthesizeInput)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("Synthesis failed: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("✓ Audio saved: %s", result.Result.AudioPath))
	printProfile(result.ProfileData)
}

func inferASR(plugin, modelfile string, tokenizerPath string) {
	spin := render.NewSpinner("loading ASR model...")
	spin.Start()

	asrInput := nexa_sdk.AsrCreateInput{
		ModelPath:     modelfile,
		TokenizerPath: tokenizerPath,
		PluginID:      plugin,
		Language:      language,
	}
	p, err := nexa_sdk.NewASR(asrInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create ASR", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	if listLanguage {
		lans, err := p.ListSupportedLanguages()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Failed to list available languages: %s", err))
			return
		}
		fmt.Println(render.GetTheme().Success.Sprintf("Available languages: %v", lans.LanguageCodes))
		listLanguage = false
	}

	repl(ReplConfig{
		ParseFile: true,

		Run: func(_prompt string, _images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			if len(audios) == 0 {
				return "", nexa_sdk.ProfileData{}, errors.New("no audio file provided")
			}

			asrConfig := &nexa_sdk.ASRConfig{
				Timestamps: "segment",
				BeamSize:   5,
				Stream:     false,
			}

			transcribeInput := nexa_sdk.AsrTranscribeInput{
				AudioPath: audios[0],
				Language:  language,
				Config:    asrConfig,
			}

			fmt.Println(render.GetTheme().Success.Sprintf("Transcribing audio file: %s", audios[0]))

			result, err := p.Transcribe(transcribeInput)
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Transcription failed: %s", err))
				return "", nexa_sdk.ProfileData{}, err
			}
			on_token(result.Result.Transcript)

			return result.Result.Transcript, result.ProfileData, nil
		},
	})
}

func inferCV(plugin, modelfile string) {
	spin := render.NewSpinner("loading CV model...")
	spin.Start()

	cvInput := nexa_sdk.CVCreateInput{
		Config: nexa_sdk.CVModelConfig{
			Capabilities:         nexa_sdk.CVCapabilityOCR,
			DetModelPath:         modelfile,
			RecModelPath:         modelfile,
			ModelPath:            "",
			ConfigFilePath:       "",
			CharDictPath:         "",
			SystemLibraryPath:    "",
			BackendLibraryPath:   "",
			ExtensionLibraryPath: "",
		},
		PluginID: plugin,
		DeviceID: "",
	}

	p, err := nexa_sdk.NewCV(cvInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create CV", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	if input == "" {
		fmt.Println(text.FgRed.Sprintf("input image file is required for CV inference"))
		fmt.Println()
		return
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		fmt.Println(text.FgRed.Sprintf("input file '%s' does not exist", input))
		return
	}

	inferInput := nexa_sdk.CVInferInput{
		InputImagePath: input,
	}

	fmt.Println(text.FgGreen.Sprintf("Performing CV inference on image: %s", input))

	result, err := p.Infer(inferInput)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("CV inference failed: %s", err))
		return
	}

	fmt.Println(text.FgGreen.Sprintf("✓ CV inference completed successfully"))
	fmt.Println(text.FgGreen.Sprintf("  Found %d results", result.ResultCount))

	for _, cvResult := range result.Results {
		if cvResult.Text != "" {
			fmt.Printf("[%s] %s\n", text.FgHiMagenta.Sprintf("%.3f", cvResult.Confidence), text.FgYellow.Sprintf("\"%s\"", cvResult.Text))
		}
	}
}

func inferEmbedder(plugin, modelfile string) {
	spin := render.NewSpinner("loading embedding model...")
	spin.Start()

	embedderInput := nexa_sdk.EmbedderCreateInput{
		ModelPath: modelfile,
		PluginID:  plugin,
	}

	p, err := nexa_sdk.NewEmbedder(embedderInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create embedder", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	prompts := prompt
	if len(prompts) == 0 {
		fmt.Println(render.GetTheme().Error.Sprintf("at least one --prompt is required for embedding generation"))
		fmt.Println()
		return
	}

	dimOutput, err := p.EmbeddingDimension()
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("failed to get embedding dimension: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("Embedding dimension: %d", dimOutput.Dimension))
	fmt.Println(render.GetTheme().Success.Sprintf("Processing %d prompts", len(prompts)))

	embedInput := nexa_sdk.EmbedderEmbedInput{
		Texts: prompts,
		Config: &nexa_sdk.EmbeddingConfig{
			BatchSize:       int32(len(prompts)),
			Normalize:       true,
			NormalizeMethod: "l2",
		},
	}

	result, err := p.Embed(embedInput)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("embedding generation failed: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("✓ Embedding generation completed successfully"))

	for i, text := range prompts {
		if i < len(result.Embeddings) {
			fmt.Printf("\n%s [%d]: %s\n", render.GetTheme().Info.Sprintf("Prompt"), i+1, text)
			fmt.Printf("%s: [%.6f]\n", render.GetTheme().Info.Sprintf("Embedding"), result.Embeddings[i])
		}
	}

	fmt.Println()
}

func inferImageGen(plugin, modeldir string) {
	prompts := prompt
	if len(prompts) == 0 {
		fmt.Println(render.GetTheme().Error.Sprintf("text prompt is required for image generation (use --prompt)"))
		fmt.Println()
		return
	}

	spin := render.NewSpinner("loading ImageGen model...")
	spin.Start()
	p, err := nexa_sdk.NewImageGen(nexa_sdk.ImageGenCreateInput{
		ModelPath: modeldir,
		PluginID:  plugin,
		DeviceID:  "cuda", // Currently only CUDA is supported
	})
	spin.Stop()
	if err != nil {
		slog.Error("failed to create ImageGen", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	if output == "" {
		output = fmt.Sprintf("imagegen_output_%d.png", time.Now().Unix())
	}

	fmt.Println(render.GetTheme().Info.Sprintf("Generating image: \"%s\"", prompts[0]))

	result, err := p.Txt2Img(nexa_sdk.ImageGenTxt2ImgInput{
		PromptUTF8: prompts[0],
		Config: &nexa_sdk.ImageGenerationConfig{
			Prompts:         prompts,
			NegativePrompts: []string{"blurry, low quality, distorted"},
			Height:          512,
			Width:           512,
			SamplerConfig: nexa_sdk.ImageSamplerConfig{
				Method:        "ddim",
				Steps:         20,
				GuidanceScale: 7.5,
				Eta:           0.0,
				Seed:          42,
			},
			SchedulerConfig: nexa_sdk.SchedulerConfig{
				Type:              "ddim",
				NumTrainTimesteps: 1000,
				StepsOffset:       1,
				BetaStart:         0.00085,
				BetaEnd:           0.012,
				BetaSchedule:      "scaled_linear",
				PredictionType:    "epsilon",
				TimestepType:      "discrete",
				TimestepSpacing:   "leading",
				InterpolationType: "linear",
				ConfigPath:        "",
			},
			Strength: 1.0,
		},
		OutputPath: output,
	})
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("Image generation failed: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("✓ Image saved to: %s", result.OutputImagePath))
}

func inferReranker(plugin, modelfile string) {
	spin := render.NewSpinner("loading reranker model...")
	spin.Start()

	rerankerInput := nexa_sdk.RerankerCreateInput{
		ModelPath: modelfile,
		PluginID:  plugin,
	}

	p, err := nexa_sdk.NewReranker(rerankerInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create reranker", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	// Check if query is provided
	if query == "" {
		fmt.Println(render.GetTheme().Error.Sprintf("--query is required for reranking"))
		fmt.Println()
		return
	}

	// Check if documents are provided
	if len(document) == 0 {
		fmt.Println(render.GetTheme().Error.Sprintf("at least one --document is required for reranking"))
		fmt.Println()
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("Query: %s", query))
	fmt.Println(render.GetTheme().Success.Sprintf("Processing %d documents", len(document)))

	// Create rerank input
	rerankInput := nexa_sdk.RerankerRerankInput{
		Query:     query,
		Documents: document,
		Config: &nexa_sdk.RerankConfig{
			BatchSize:       int32(len(document)),
			Normalize:       true,
			NormalizeMethod: "softmax",
		},
	}

	// Perform reranking
	result, err := p.Rerank(rerankInput)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("reranking failed: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("✓ Reranking completed successfully"))
	fmt.Println(render.GetTheme().Success.Sprintf("  Generated %d scores", len(result.Scores)))

	// Display results
	for i, doc := range document {
		if i < len(result.Scores) {
			fmt.Printf("\n%s [%d]: %s\n", render.GetTheme().Info.Sprintf("Document"), i+1, doc)
			fmt.Printf("%s: %.6f\n", render.GetTheme().Info.Sprintf("Score"), result.Scores[i])
		}
	}
}
