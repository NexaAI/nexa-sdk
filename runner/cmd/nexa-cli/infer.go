package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

const modelLoadFailMsg = `⚠️ Oops. Model failed to load.

👉 Try these:
- Verify your system meets the model’s requirements.
- Seek help in our discord or slack.`

var (
	// disableStream *bool // reuse in run.go
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
	inferCmd.Flags().StringArrayVarP(&tool, "tool", "t", nil, "[llm|vlm] add tool to make function call")
	inferCmd.Flags().StringArrayVarP(&prompt, "prompt", "p", nil, "[embedder|tts] pass prompt")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "[reranker] query")
	inferCmd.Flags().StringArrayVarP(&document, "document", "d", nil, "[reranker] documents")
	inferCmd.Flags().StringVarP(&input, "input", "i", "", "[asr] input file (audio for asr)")
	inferCmd.Flags().StringVarP(&output, "output", "o", "", "[tts] output file (audio for tts)")
	inferCmd.Flags().StringVarP(&voice, "voice", "", "", "[tts] voice identifier")
	inferCmd.Flags().BoolVarP(&listVoice, "list-voice", "", false, "[tts] list available voices")
	inferCmd.Flags().Float64VarP(&speechSpeed, "speech-speed", "", 1.0, "[tts] speech speed (1.0 = normal)")
	inferCmd.Flags().StringVarP(&language, "language", "", "", "[asr] language code (e.g., en, zh, ja)")
	inferCmd.Flags().BoolVarP(&listLanguage, "list-language", "", false, "[asr] list available languages")

	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		model := normalizeModelName(args[0])

		s := store.Get()

		manifest, err := s.GetManifest(model)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(text.FgBlue.Sprintf("model not found, start download"))

			pull().Run(cmd, args)
			// check agin
			manifest, err = s.GetManifest(model)
		}
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
			return
		}

		var modelFile string
		var options []huh.Option[string]
		for k, v := range manifest.ModelFile {
			if v.Downloaded {
				options = append(options, huh.NewOption(
					fmt.Sprintf("%-10s [%7s]", k, humanize.IBytes(uint64(v.Size))),
					v.Name,
				))
			}
		}
		if len(options) >= 2 {
			if err = huh.NewSelect[string]().
				Title("Select a quant from local folder").
				Options(options...).
				Value(&modelFile).
				Run(); err != nil {
				fmt.Println(text.FgRed.Sprintf("select error: %s", err))
				return
			}
		} else {
			modelFile = options[0].Value
		}

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, modelFile)

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
			// inferEmbed(modelfile, nil)
		case types.ModelTypeReranker:
			// inferRerank(modelfile, nil)
		case types.ModelTypeTTS:
			inferTTS(manifest.PluginId, modelfile, "")
		case types.ModelTypeASR:
			inferASR(manifest.PluginId, modelfile, "")
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
			NCtx: 2048,
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

			templateOutput, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{Messages: history})
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
			NCtx: 2048,
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

			tmplOut, err := p.ApplyChatTemplate(nexa_sdk.VlmApplyChatTemplateInput{Messages: history})
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			res, err := p.Generate(nexa_sdk.VlmGenerateInput{
				PromptUTF8: tmplOut.FormattedText,
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
			fmt.Println(text.FgRed.Sprintf("Failed to list available voices: %s", err))
			return
		}
		fmt.Println(text.FgGreen.Sprintf("Available voices: %v", voices.VoiceIDs))
		return
	}

	// Check if prompt is provided
	if len(prompt) == 0 {
		fmt.Println(text.FgRed.Sprintf("text is required for TTS synthesis (use --prompt)"))
		fmt.Println()
		return
	}

	// Combine all prompt texts
	textToSynthesize := strings.Join(prompt, " ")

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

	fmt.Println(text.FgGreen.Sprintf("Synthesizing speech: \"%s\"", textToSynthesize))

	result, err := p.Synthesize(synthesizeInput)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Synthesis failed: %s", err))
		return
	}

	fmt.Println(text.FgGreen.Sprintf("✓ Audio saved to: %s", result.Result.AudioPath))
	// fmt.Printf("  Duration: %.2f seconds\n", result.Result.DurationSeconds)
	// fmt.Printf("  Sample rate: %d Hz\n", result.Result.SampleRate)
}

func inferASR(plugin, modelfile string, tokenizerPath string) {
	spin := render.NewSpinner("loading ASR model...")
	spin.Start()

	asrInput := nexa_sdk.AsrCreateInput{
		ModelPath:     modelfile,
		TokenizerPath: tokenizerPath,
		PluginID:      plugin,
	}
	if language != "" {
		asrInput.Language = language
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
			fmt.Println(text.FgRed.Sprintf("Failed to list available languages: %s", err))
			return
		}
		fmt.Println(text.FgGreen.Sprintf("Available languages: %v", lans.LanguageCodes))
		return
	}

	if input == "" {
		fmt.Println(text.FgRed.Sprintf("input audio file is required for ASR transcription"))
		fmt.Println()
		return
	}

	if _, err := os.Stat(input); os.IsNotExist(err) {
		fmt.Println(text.FgRed.Sprintf("input file '%s' does not exist", input))
		return
	}

	asrConfig := &nexa_sdk.ASRConfig{
		Timestamps: "segment",
		BeamSize:   5,
		Stream:     false,
	}

	transcribeInput := nexa_sdk.AsrTranscribeInput{
		AudioPath: input,
		Language:  language,
		Config:    asrConfig,
	}

	fmt.Println(text.FgGreen.Sprintf("Transcribing audio file: %s", input))

	result, err := p.Transcribe(transcribeInput)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Transcription failed: %s", err))
		return
	}

	fmt.Println(text.FgYellow.Sprint(result.Result.Transcript))
}
