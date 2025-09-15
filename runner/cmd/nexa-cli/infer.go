package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/record"
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
	maxTokens    int32
	enableThink  bool
	hideThink    bool
	prompt       []string
	taskType     string
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

var (
	ErrNoAudio = errors.New("no audio file provided")
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
	inferCmd.Flags().Int32VarP(&maxTokens, "max-tokens", "", 2048, "[llm|vlm] max tokens")
	inferCmd.Flags().BoolVarP(&enableThink, "think", "", true, "[llm|vlm] enable thinking mode")
	inferCmd.Flags().BoolVarP(&hideThink, "hide-think", "", false, "[llm|vlm] hide thinking output")
	inferCmd.Flags().StringArrayVarP(&prompt, "prompt", "p", nil, "[embedder|tts|image_gen] pass prompt")
	inferCmd.Flags().StringVarP(&taskType, "task-type", "", "default", "[embedder] task type: default|search_query|search_document")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "[reranker] query")
	inferCmd.Flags().StringArrayVarP(&document, "document", "d", nil, "[reranker] documents")
	inferCmd.Flags().StringVarP(&input, "input", "i", "", "[cv] input file (image for cv)")
	inferCmd.Flags().StringVarP(&output, "output", "o", "", "[tts|image_gen] output file (audio for tts / image for image_gen)")
	inferCmd.Flags().StringVarP(&voice, "voice", "", "", "[tts] voice identifier")
	inferCmd.Flags().BoolVarP(&listVoice, "list-voice", "", false, "[tts] list available voices")
	inferCmd.Flags().Float64VarP(&speechSpeed, "speech-speed", "", 1.0, "[tts] speech speed (1.0 = normal)")
	// inferCmd.Flags().StringVarP(&language, "language", "", "", "[asr] language code (e.g., en, zh, ja)")           // TODO: Language support not implemented yet
	// inferCmd.Flags().BoolVarP(&listLanguage, "list-language", "", false, "[asr] list available languages")        // TODO: Language support not implemented yet

	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.Get()

		manifest, err := ensureModelAvailable(s, normalizeModelName(args[0]), cmd, args)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("parse manifest error: %s", err))
			return
		}

		quant, err := selectQuant(manifest)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("quant error: %s", err))
			return
		}
		// fmt.Println(render.GetTheme().Quant.Sprintf("🔹 Quant=%s", quant))

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		switch manifest.ModelType {
		case types.ModelTypeLLM:
			inferLLM(manifest, quant)
		case types.ModelTypeVLM:
			inferVLM(manifest, quant)
		case types.ModelTypeEmbedder:
			inferEmbedder(manifest, quant)
		case types.ModelTypeReranker:
			inferReranker(manifest, quant)
		case types.ModelTypeTTS:
			inferTTS(manifest, quant)
		case types.ModelTypeASR:
			inferASR(manifest, quant)
		case types.ModelTypeCV:
			inferCV(manifest, quant)
		case types.ModelTypeImageGen:
			// ImageGen model is a directory, not a file
			inferImageGen(manifest, quant)
		default:
			panic("not support model type")
		}
	}
	return inferCmd
}

func ensureModelAvailable(s *store.Store, model string, cmd *cobra.Command, args []string) (*types.ModelManifest, error) {
	manifest, err := s.GetManifest(model)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println(render.GetTheme().Info.Sprintf("model not found, start download"))
		pull().Run(cmd, args)
		manifest, err = s.GetManifest(model)
	}
	return manifest, err
}

func selectQuant(manifest *types.ModelManifest) (string, error) {
	var options []huh.Option[string]
	for k, v := range manifest.ModelFile {
		if v.Downloaded {
			options = append(options, huh.NewOption(fmt.Sprintf("%-10s [%7s]", k, humanize.IBytes(uint64(v.Size))), k))
		}
	}
	if len(options) == 0 {
		return "", fmt.Errorf("no quant found")
	}
	if len(options) == 1 {
		return options[0].Value, nil
	}
	var quant string
	if err := huh.NewSelect[string]().Title("Select a quant from local folder").Options(options...).Value(&quant).Run(); err != nil {
		return "", err
	}
	return quant, nil
}

func inferLLM(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading model...")
	spin.Start()

	p, err := nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
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

		Reset: func() error {
			err := p.Reset()
			if err == nil {
				history = nil
			}
			return err
		},

		SaveKVCache: func(path string) error {
			_, err := p.SaveKVCache(nexa_sdk.LlmSaveKVCacheInput{Path: path})
			return err
		},

		LoadKVCache: func(path string) error {
			_, err := p.LoadKVCache(nexa_sdk.LlmLoadKVCacheInput{Path: path})
			if err == nil {
				history = nil
			}
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
					MaxTokens: maxTokens,
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

func inferVLM(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	var mmprojfile string
	if manifest.MMProjFile.Name != "" {
		mmprojfile = s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
	}
	var tokenizerfile string
	if manifest.TokenizerFile.Name != "" {
		tokenizerfile = s.ModelfilePath(manifest.Name, manifest.TokenizerFile.Name)
	}
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewVLM(nexa_sdk.VlmCreateInput{
		ModelName:     manifest.ModelName,
		ModelPath:     modelfile,
		MmprojPath:    mmprojfile,
		TokenizerPath: tokenizerfile,
		PluginID:      manifest.PluginId,
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

		Reset: func() error {
			err := p.Reset()
			if err == nil {
				history = nil
			}
			return err
		},

		Record: func() (*string, error) {
			t := strconv.Itoa(int(time.Now().Unix()))
			outputFile := filepath.Join(os.TempDir(), "nexa-cli", t+".wav")
			rec, err := record.NewRecorder(outputFile)
			if err != nil {
				return nil, err
			}

			fmt.Println(render.GetTheme().Info.Sprint("Recording is going on, press Ctrl-C to stop"))

			err = rec.Run()
			if err != nil {
				return nil, err
			}
			outfile := rec.GetOutputFile()
			return &outfile, nil
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

			tmplOut, err := p.ApplyChatTemplate(nexa_sdk.VlmApplyChatTemplateInput{
				Messages:    history,
				EnableThink: enableThink,
			})
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			res, err := p.Generate(nexa_sdk.VlmGenerateInput{
				PromptUTF8: tmplOut.FormattedText,
				OnToken:    on_token,
				Config: &nexa_sdk.GenerationConfig{
					MaxTokens:  maxTokens,
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

func inferTTS(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading TTS model...")
	spin.Start()

	ttsInput := nexa_sdk.TtsCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
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

func inferASR(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading ASR model...")
	spin.Start()

	asrInput := nexa_sdk.AsrCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
		Language:  language,
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

		Record: func() (*string, error) {
			streamConfig := nexa_sdk.ASRStreamConfig{
				ChunkDuration:   4.0,
				OverlapDuration: 3.5,
				SampleRate:      16000,
				MaxQueueSize:    10,
				BufferSize:      1024,
				Timestamps:      "segment",
				BeamSize:        4,
			}
			_, err := p.StreamBegin(nexa_sdk.AsrStreamBeginInput{
				StreamConfig: &streamConfig,
				Language:     "en",
				OnTranscription: func(text string, _ any) {
					tWidth := getTerminalWidth()
					if len(text) > tWidth {
						text = "..." + text[len(text)-tWidth+3:]
					}
					text += strings.Repeat(" ", tWidth-len(text))

					fmt.Print("\r")
					fmt.Print(render.GetTheme().ModelOutput.Sprint(text))
				},
				UserData: nil,
			})
			slog.Debug("ASR StreamBegin", "error", err)
			if err != nil && !errors.Is(err, nexa_sdk.ErrCommonNotSupport) {
				return nil, err
			}
			defer p.StreamStop(nexa_sdk.AsrStreamStopInput{})

			// streaming not supported, fallback to file input
			if err != nil {
				t := strconv.Itoa(int(time.Now().Unix()))
				outputFile := filepath.Join(os.TempDir(), "nexa-cli", t+".wav")
				rec, err := record.NewRecorder(outputFile)
				if err != nil {
					return nil, err
				}

				fmt.Println(render.GetTheme().Info.Sprint("Recording is going on, press Ctrl-C to stop"))

				err = rec.Run()
				if err != nil {
					return nil, err
				}
				outfile := rec.GetOutputFile()

				asrConfig := &nexa_sdk.ASRConfig{
					Timestamps: "segment",
					BeamSize:   5,
					Stream:     false,
				}

				transcribeInput := nexa_sdk.AsrTranscribeInput{
					AudioPath: outfile,
					Language:  language,
					Config:    asrConfig,
				}

				result, err := p.Transcribe(transcribeInput)
				if err != nil {
					return nil, err
				}

				fmt.Println(render.GetTheme().ModelOutput.Sprint(result.Result.Transcript))
				render.GetTheme().Reset()
				fmt.Println()
				printProfile(result.ProfileData)

			} else {

				rec, err := record.NewStreamRecorder()
				if err != nil {
					return nil, err
				}
				fmt.Println(render.GetTheme().Info.Sprint("Streaming ASR recording started, press Ctrl-C to stop"))
				fmt.Println()

				if err := rec.Start(); err != nil {
					return nil, err
				}
				defer rec.Stop()

				buffer := make([]float32, streamConfig.BufferSize)
				for {
					n, err := rec.ReadFloat32(buffer)
					if err == io.EOF {
						break
					}

					if err := p.StreamPushAudio(nexa_sdk.AsrStreamPushAudioInput{
						AudioData: buffer[:n],
					}); err != nil {
						fmt.Println(render.GetTheme().Error.Sprintf("error pushing audio data: %s", err))
						fmt.Println()
						return nil, err
					}
				}
			}

			return nil, nil
		},

		Run: func(_prompt string, _images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			if len(audios) == 0 {
				return "", nexa_sdk.ProfileData{}, ErrNoAudio
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
				return "", nexa_sdk.ProfileData{}, err
			}
			on_token(result.Result.Transcript)

			return result.Result.Transcript, result.ProfileData, nil
		},
	})
}

func inferCV(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading CV model...")
	spin.Start()

	cvInput := nexa_sdk.CVCreateInput{
		ModelName: manifest.ModelName,
		Config: nexa_sdk.CVModelConfig{
			Capabilities: nexa_sdk.CVCapabilityOCR,
			DetModelPath: modelfile,
			RecModelPath: modelfile,
		},
		PluginID: manifest.PluginId,
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
		fmt.Println(render.GetTheme().Error.Sprintf("input image file is required for CV inference"))
		fmt.Println()
		return
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		fmt.Println(render.GetTheme().Error.Sprintf("input file '%s' does not exist", input))
		return
	}

	inferInput := nexa_sdk.CVInferInput{
		InputImagePath: input,
	}

	fmt.Println(render.GetTheme().Info.Sprintf("Performing CV inference on image: %s", input))

	result, err := p.Infer(inferInput)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("CV inference failed: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Info.Sprintf("✓ CV inference completed successfully"))
	fmt.Println(render.GetTheme().Info.Sprintf("  Found %d results", len(result.Results)))

	for _, cvResult := range result.Results {
		fmt.Printf("[%s] %s\n", render.GetTheme().Info.Sprintf("%.3f", cvResult.Confidence), render.GetTheme().Success.Sprintf("\"%s\"", cvResult.Text))
	}
}

func inferEmbedder(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading embedding model...")
	spin.Start()

	embedderInput := nexa_sdk.EmbedderCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
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
		TaskType: taskType,
		Texts:    prompts,
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

	embeddingDim := int(dimOutput.Dimension)
	for i, text := range prompts {
		startIdx := i * embeddingDim
		endIdx := startIdx + embeddingDim

		fmt.Printf("\n%s [%d]: %s\n", render.GetTheme().Info.Sprintf("Prompt"), i+1, text)
		embedding := result.Embeddings[startIdx:endIdx]

		fmt.Printf("%s: [%.6f, %.6f, %.6f, ..., %.6f, %.6f, %.6f] (length: %d)\n",
			render.GetTheme().Info.Sprintf("Embedding"),
			embedding[0], embedding[1], embedding[2],
			embedding[len(embedding)-3], embedding[len(embedding)-2], embedding[len(embedding)-1],
			len(embedding))
	}

	fmt.Println()
}

func inferImageGen(manifest *types.ModelManifest, _ string) {
	s := store.Get()
	modeldir := s.ModelfilePath(manifest.Name, "")
	prompts := prompt
	if len(prompts) == 0 {
		fmt.Println(render.GetTheme().Error.Sprintf("text prompt is required for image generation (use --prompt)"))
		fmt.Println()
		return
	}

	spin := render.NewSpinner("loading ImageGen model...")
	spin.Start()
	p, err := nexa_sdk.NewImageGen(nexa_sdk.ImageGenCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modeldir,
		PluginID:  manifest.PluginId,
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

func inferReranker(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading reranker model...")
	spin.Start()

	rerankerInput := nexa_sdk.RerankerCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
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
