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
	"unicode"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

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
	noInteractive bool
	// disableStream *bool // reuse in run.go
	ngl            int32
	nctx           int32
	maxTokens      int32
	imageMaxLength int32
	enableThink    bool
	hideThink      bool
	prompt         []string
	taskType       string
	query          string
	document       []string
	input          string
	output         string
	voice          string
	listVoice      bool
	speechSpeed    float64
	systemPrompt   string
	language       string
	listLanguage   bool

	// sampler config
	temperature       float32
	topP              float32
	topK              int32
	minP              float32
	repetitionPenalty float32
	presencePenalty   float32
	frequencyPenalty  float32
	seed              int32
	grammarPath       string
	grammarString     string
	enableJson        bool
)

var (
	samplerFlags = func() *pflag.FlagSet {
		samplerFlags := pflag.NewFlagSet("LLM/VLM Sampler", pflag.ExitOnError)
		samplerFlags.SortFlags = false
		samplerFlags.Float32VarP(&temperature, "temperature", "", 0.0, "sampling temperature")
		samplerFlags.Float32VarP(&topP, "top-p", "", 0.0, "top-p sampling")
		samplerFlags.Int32VarP(&topK, "top-k", "", 0, "top-k sampling")
		samplerFlags.Float32VarP(&minP, "min-p", "", 0.0, "min-p sampling")
		samplerFlags.Float32VarP(&repetitionPenalty, "repetition-penalty", "", 1.0, "repetition penalty")
		samplerFlags.Float32VarP(&presencePenalty, "presence-penalty", "", 0.0, "presence penalty")
		samplerFlags.Float32VarP(&frequencyPenalty, "frequency-penalty", "", 0.0, "frequency penalty")
		samplerFlags.Int32VarP(&seed, "seed", "", 0, "random seed")
		samplerFlags.StringVarP(&grammarPath, "grammar-path", "", "", "path to grammar file")
		samplerFlags.StringVarP(&grammarString, "grammar-string", "", "", "grammar in string format")
		samplerFlags.BoolVarP(&enableJson, "enable-json", "", false, "enable json output")
		return samplerFlags
	}()
	llmFlags = func() *pflag.FlagSet {
		llmFlags := pflag.NewFlagSet("LLM/VLM Model", pflag.ExitOnError)
		llmFlags.SortFlags = false
		llmFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
		llmFlags.Int32VarP(&ngl, "ngl", "n", 999, "num of layers pass to gpu")
		llmFlags.Int32VarP(&nctx, "nctx", "", 4096, "context window size")
		llmFlags.Int32VarP(&maxTokens, "max-tokens", "", 2048, "max tokens")
		llmFlags.BoolVarP(&enableThink, "think", "", true, "enable thinking mode")
		llmFlags.BoolVarP(&hideThink, "hide-think", "", false, "hide thinking output")
		llmFlags.StringVarP(&systemPrompt, "system-prompt", "s", "", "system prompt to set model behavior")
		llmFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
		llmFlags.StringVarP(&input, "input", "i", "", "prompt txt file")
		return llmFlags
	}()
	vlmFlags = func() *pflag.FlagSet {
		vlmFlags := pflag.NewFlagSet("VLM Specific", pflag.ExitOnError)
		vlmFlags.SortFlags = false
		vlmFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
		vlmFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
		vlmFlags.Int32VarP(&imageMaxLength, "image-max-length", "", 512, "max image length")
		return vlmFlags
	}()
)

var (
	ErrNoAudio = errors.New("no audio file provided")
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{
		GroupID: "inference",
		Use:     "infer <model-name>",
		Short:   "Infer with a model",
		Long:    "Run inference with a specified model. The model must be downloaded and cached locally.",
	}

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	// NOTE: flagset use same flag name will be ignored, but usage is different, so we keep them in different flagset

	inferCmd.Flags().AddFlagSet(samplerFlags)
	inferCmd.Flags().AddFlagSet(llmFlags)
	inferCmd.Flags().AddFlagSet(vlmFlags)

	embedderFlags := pflag.NewFlagSet("Embedder", pflag.ExitOnError)
	embedderFlags.SortFlags = false
	embedderFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
	embedderFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
	embedderFlags.StringVarP(&taskType, "task-type", "", "default", "default|search_query|search_document")
	inferCmd.Flags().AddFlagSet(embedderFlags)

	rerankerFlags := pflag.NewFlagSet("Reranker", pflag.ExitOnError)
	rerankerFlags.SortFlags = false
	rerankerFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
	rerankerFlags.StringVarP(&query, "query", "q", "", "query")
	rerankerFlags.StringArrayVarP(&document, "document", "d", nil, "documents")
	inferCmd.Flags().AddFlagSet(rerankerFlags)

	cvFlags := pflag.NewFlagSet("CV", pflag.ExitOnError)
	cvFlags.SortFlags = false
	cvFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
	cvFlags.StringVarP(&input, "input", "i", "", "input image file")
	inferCmd.Flags().AddFlagSet(cvFlags)

	ttsFlags := pflag.NewFlagSet("TTS", pflag.ExitOnError)
	ttsFlags.SortFlags = false
	ttsFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
	ttsFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
	ttsFlags.StringVarP(&voice, "voice", "", "", "voice identifier")
	ttsFlags.BoolVarP(&listVoice, "list-voice", "", false, "list available voices")
	ttsFlags.Float64VarP(&speechSpeed, "speech-speed", "", 1.0, "speech speed (1.0 = normal)")
	ttsFlags.StringVarP(&output, "output", "o", "", "output audio file")
	inferCmd.Flags().AddFlagSet(ttsFlags)

	imageGenFlags := pflag.NewFlagSet("ImageGen", pflag.ExitOnError)
	imageGenFlags.SortFlags = false
	imageGenFlags.BoolVarP(&noInteractive, "no-interactive", "", false, "disable interactive mode")
	imageGenFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
	imageGenFlags.StringVarP(&output, "output", "o", "", "output image file")
	inferCmd.Flags().AddFlagSet(imageGenFlags)

	// inferCmd.Flags().StringVarP(&language, "language", "", "", "[asr] language code (e.g., en, zh, ja)")           // TODO: Language support not implemented yet
	// inferCmd.Flags().BoolVarP(&listLanguage, "list-language", "", false, "[asr] list available languages")        // TODO: Language support not implemented yet

	inferCmd.SetUsageFunc(func(c *cobra.Command) error {
		flagGroups := []*pflag.FlagSet{
			samplerFlags, llmFlags, vlmFlags, embedderFlags, rerankerFlags, cvFlags, ttsFlags, imageGenFlags,
		}
		w := c.OutOrStdout()
		fmt.Fprint(w, "Usage:")
		if c.Runnable() {
			fmt.Fprintf(w, "\n  %s", c.UseLine())
		}
		if len(c.Aliases) > 0 {
			fmt.Fprintf(w, "\n\nAliases:\n")
			fmt.Fprintf(w, "  %s", c.NameAndAliases())
		}

		for _, flags := range flagGroups {
			fmt.Fprintf(w, "\n\n%s Flags:\n", flags.Name())
			fmt.Fprint(w, strings.TrimRightFunc(flags.FlagUsages(), unicode.IsSpace))
		}

		if c.HasAvailableInheritedFlags() {
			fmt.Fprintf(w, "\n\nGlobal Flags:\n")
			fmt.Fprint(w, strings.TrimRightFunc(c.InheritedFlags().FlagUsages(), unicode.IsSpace))
		}
		fmt.Fprintln(w)
		return nil
	})

	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.Get()

		manifest, err := ensureModelAvailable(s, args[0])
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("check model error: %s", err))
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
			checkDependency()
			inferVLM(manifest, quant)
		case types.ModelTypeEmbedder:
			inferEmbedder(manifest, quant)
		case types.ModelTypeReranker:
			inferReranker(manifest, quant)
		case types.ModelTypeTTS:
			inferTTS(manifest, quant)
		case types.ModelTypeASR:
			checkDependency()
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

func ensureModelAvailable(s *store.Store, name string) (*types.ModelManifest, error) {
	name = normalizeModelName(name)
	manifest, err := s.GetManifest(name)
	if errors.Is(err, os.ErrNotExist) {
		fmt.Println(render.GetTheme().Info.Sprintf("model not found, start download"))
		err = pullModel(name)
		if err != nil {
			return nil, fmt.Errorf("download model failed")
		}
		manifest, err = s.GetManifest(name)
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
	var samplerConfig *nexa_sdk.SamplerConfig
	if temperature > 0 || topK > 0 || topP > 0 || minP > 0 ||
		repetitionPenalty != 1.0 || presencePenalty != 0.0 || frequencyPenalty != 0.0 ||
		seed != 0 || grammarPath != "" || grammarString != "" ||
		enableJson {
		samplerConfig = &nexa_sdk.SamplerConfig{
			Temperature:       temperature,
			TopP:              topP,
			TopK:              topK,
			MinP:              minP,
			RepetitionPenalty: repetitionPenalty,
			PresencePenalty:   presencePenalty,
			FrequencyPenalty:  frequencyPenalty,
			Seed:              seed,
			GrammarPath:       grammarPath,
			GrammarString:     grammarString,
			EnableJson:        enableJson,
		}
	}

	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading model...")
	spin.Start()

	p, err := nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
		DeviceID:  manifest.DeviceId,
		Config: nexa_sdk.ModelConfig{
			NCtx:         nctx,
			NGpuLayers:   ngl,
			SystemPrompt: systemPrompt,
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

	if systemPrompt != "" {
		history = append(history, nexa_sdk.LlmChatMessage{Role: nexa_sdk.LLMRoleSystem, Content: systemPrompt})
	}
	if len(input) > 0 && !noInteractive {
		content, err := os.ReadFile(input)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("read prompt file error: %s", err))
			return
		}
		history = append(history, nexa_sdk.LlmChatMessage{Role: nexa_sdk.LLMRoleUser, Content: string(content)})
		applyChatTemplateInput, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
			Messages:    history,
			EnableThink: enableThink,
		})
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("apply chat template error: %s", err))
			return
		}

		res, err := p.Generate(nexa_sdk.LlmGenerateInput{
			PromptUTF8: applyChatTemplateInput.FormattedText,
			OnToken: func(token string) bool {
				fmt.Print(token)
				return true
			},
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens: maxTokens,
			},
		})
		fmt.Println()
		fmt.Println()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("generate error: %s", err))
			return
		}
		printProfile(res.ProfileData)
		// return
	}

	repl(ReplConfig{
		ParseFile:     false,
		NoInteractive: noInteractive,

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
					MaxTokens:     maxTokens,
					SamplerConfig: samplerConfig,
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
	var samplerConfig *nexa_sdk.SamplerConfig
	if temperature > 0 || topK > 0 || topP > 0 || minP > 0 ||
		repetitionPenalty != 1.0 || presencePenalty != 0.0 || frequencyPenalty != 0.0 ||
		seed != 0 || grammarPath != "" || grammarString != "" ||
		enableJson {
		samplerConfig = &nexa_sdk.SamplerConfig{
			Temperature:       temperature,
			TopP:              topP,
			TopK:              topK,
			MinP:              minP,
			RepetitionPenalty: repetitionPenalty,
			PresencePenalty:   presencePenalty,
			FrequencyPenalty:  frequencyPenalty,
			Seed:              seed,
			GrammarPath:       grammarPath,
			GrammarString:     grammarString,
			EnableJson:        enableJson,
		}
	}

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
		DeviceID:      manifest.DeviceId,
		Config: nexa_sdk.ModelConfig{
			NCtx:         nctx,
			NGpuLayers:   ngl,
			SystemPrompt: systemPrompt,
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

	if systemPrompt != "" {
		history = append(history, nexa_sdk.VlmChatMessage{Role: nexa_sdk.VlmRoleSystem, Contents: []nexa_sdk.VlmContent{{Type: nexa_sdk.VlmContentTypeText, Text: systemPrompt}}})
	}
	if len(input) > 0 && !noInteractive {
		content, err := os.ReadFile(input)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("read prompt file error: %s", err))
			return
		}
		history = append(history, nexa_sdk.VlmChatMessage{Role: nexa_sdk.VlmRoleUser, Contents: []nexa_sdk.VlmContent{{Type: nexa_sdk.VlmContentTypeText, Text: string(content)}}})
		applyChatTemplateInput, err := p.ApplyChatTemplate(nexa_sdk.VlmApplyChatTemplateInput{
			Messages:    history,
			EnableThink: enableThink,
		})
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("apply chat template error: %s", err))
			return
		}

		res, err := p.Generate(nexa_sdk.VlmGenerateInput{
			PromptUTF8: applyChatTemplateInput.FormattedText,
			OnToken: func(token string) bool {
				fmt.Print(token)
				return true
			},
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens: maxTokens,
			},
		})
		fmt.Println()
		fmt.Println()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("generate error: %s", err))
			return
		}
		printProfile(res.ProfileData)
		// return
	}

	repl(ReplConfig{
		ParseFile:     true,
		NoInteractive: noInteractive,

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
					MaxTokens:      maxTokens,
					SamplerConfig:  samplerConfig,
					ImagePaths:     images,
					ImageMaxLength: imageMaxLength,
					AudioPaths:     audios,
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

func inferEmbedder(manifest *types.ModelManifest, quant string) {
	s := store.Get()
	modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)
	spin := render.NewSpinner("loading embedding model...")
	spin.Start()

	embedderInput := nexa_sdk.EmbedderCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modelfile,
		PluginID:  manifest.PluginId,
		DeviceID:  manifest.DeviceId,
	}

	p, err := nexa_sdk.NewEmbedder(embedderInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create embedder", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	dimOutput, err := p.EmbeddingDimension()
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("failed to get embedding dimension: %s", err))
		return
	}

	fmt.Println(render.GetTheme().Success.Sprintf("Embedding dimension: %d", dimOutput.Dimension))

	// Non-interactive mode: use command line arguments directly
	if noInteractive {
		if len(prompt) == 0 {
			fmt.Println(render.GetTheme().Error.Sprintf("--prompt is required for embedding"))
			return
		}

		// Create embed input
		embedInput := nexa_sdk.EmbedderEmbedInput{
			TaskType: taskType,
			Texts:    prompt,
			Config: &nexa_sdk.EmbeddingConfig{
				BatchSize:       1,
				Normalize:       true,
				NormalizeMethod: "l2",
			},
		}

		// Perform embedding
		result, err := p.Embed(embedInput)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Embedding failed: %s", err))
			return
		}

		if len(result.Embeddings) == 0 {
			fmt.Println(render.GetTheme().Error.Sprintf("no embeddings generated"))
			return
		}

		// Output results
		n, emb := len(result.Embeddings), result.Embeddings
		info := render.GetTheme().Info.Sprintf("Embedding")
		var out string
		if n > 6 {
			out = render.GetTheme().Success.Sprintf(
				"[%.6f, %.6f, %.6f, ..., %.6f, %.6f, %.6f] (length: %d)",
				emb[0], emb[1], emb[2],
				emb[n-3], emb[n-2], emb[n-1], n,
			)
		} else {
			out = render.GetTheme().Success.Sprintf("%v (length: %d)", emb, n)
		}

		fmt.Printf("%s: %s\n", info, out)
		printProfile(result.ProfileData)
		return
	}

	// Interactive mode: use repl
	repl(ReplConfig{
		ParseFile:     false,
		NoInteractive: false,

		Run: func(prompt string, _, _ []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			embedInput := nexa_sdk.EmbedderEmbedInput{
				TaskType: taskType,
				Texts:    []string{strings.TrimSpace(prompt)},
				Config: &nexa_sdk.EmbeddingConfig{
					BatchSize:       1,
					Normalize:       true,
					NormalizeMethod: "l2",
				},
			}

			result, err := p.Embed(embedInput)
			if err != nil {
				return "", result.ProfileData, err
			}

			if len(result.Embeddings) == 0 {
				return "", result.ProfileData, fmt.Errorf("no embeddings generated")
			}

			n, emb := len(result.Embeddings), result.Embeddings
			info := render.GetTheme().Info.Sprintf("Embedding")
			var out string
			if n > 6 {
				out = render.GetTheme().Success.Sprintf(
					"[%.6f, %.6f, %.6f, ..., %.6f, %.6f, %.6f] (length: %d)",
					emb[0], emb[1], emb[2],
					emb[n-3], emb[n-2], emb[n-1], n,
				)
			} else {
				out = render.GetTheme().Success.Sprintf("%v (length: %d)", emb, n)
			}

			on_token(fmt.Sprintf("%s: %s", info, out))

			return "", result.ProfileData, nil
		},
	})
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
		DeviceID:  manifest.DeviceId,
	}

	p, err := nexa_sdk.NewReranker(rerankerInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create reranker", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	// Non-interactive mode: use command line arguments directly
	if noInteractive {
		if query == "" {
			fmt.Println(render.GetTheme().Error.Sprintf("--query is required for reranking"))
			return
		}
		if len(document) == 0 {
			fmt.Println(render.GetTheme().Error.Sprintf("at least one --document is required for reranking"))
			return
		}

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
			fmt.Println(render.GetTheme().Error.Sprintf("Reranking failed: %s", err))
			return
		}

		// Output results
		fmt.Printf("Query: %s\n", query)
		fmt.Printf("Processing %d documents\n", len(document))
		fmt.Printf("✓ Reranking completed successfully\n")
		fmt.Printf("  Generated %d scores\n", len(result.Scores))

		for i, doc := range document {
			if i < len(result.Scores) {
				fmt.Printf("\nDocument [%d]: %s\n", i+1, doc)
				fmt.Printf("Score: %.6f\n", result.Scores[i])
			}
		}

		printProfile(result.ProfileData)
		return
	}

	// Interactive mode: use repl
	repl(ReplConfig{
		ParseFile:     false,
		NoInteractive: false,

		Run: func(prompt string, _, _ []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			// Parse prompt to extract query and documents
			// Format: "query|doc1|doc2|doc3"
			parts := strings.Split(prompt, "|")
			if len(parts) < 2 {
				return "", nexa_sdk.ProfileData{}, fmt.Errorf("format: query|doc1|doc2|doc3")
			}

			queryText := strings.TrimSpace(parts[0])
			documents := make([]string, 0, len(parts)-1)
			for i := 1; i < len(parts); i++ {
				doc := strings.TrimSpace(parts[i])
				if doc != "" {
					documents = append(documents, doc)
				}
			}

			if len(documents) == 0 {
				return "", nexa_sdk.ProfileData{}, fmt.Errorf("at least one document is required")
			}

			// Create rerank input
			rerankInput := nexa_sdk.RerankerRerankInput{
				Query:     queryText,
				Documents: documents,
				Config: &nexa_sdk.RerankConfig{
					BatchSize:       int32(len(documents)),
					Normalize:       true,
					NormalizeMethod: "softmax",
				},
			}

			// Perform reranking
			result, err := p.Rerank(rerankInput)
			if err != nil {
				return "", result.ProfileData, err
			}

			// Format output
			output := fmt.Sprintf("Query: %s\n", queryText)
			output += fmt.Sprintf("Processing %d documents\n", len(documents))
			output += "✓ Reranking completed successfully\n"
			output += fmt.Sprintf("  Generated %d scores\n", len(result.Scores))

			for i, doc := range documents {
				if i < len(result.Scores) {
					output += fmt.Sprintf("\nDocument [%d]: %s\n", i+1, doc)
					output += fmt.Sprintf("Score: %.6f\n", result.Scores[i])
				}
			}

			on_token(output)
			return output, result.ProfileData, nil
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
		DeviceID:  manifest.DeviceId,
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

	// Non-interactive mode: use command line arguments directly
	if noInteractive {
		if len(prompt) == 0 {
			fmt.Println(render.GetTheme().Error.Sprintf("text is required for TTS synthesis (use --prompt)"))
			return
		}

		// Check for empty strings in prompts
		for _, p := range prompt {
			if strings.TrimSpace(p) == "" {
				fmt.Println(render.GetTheme().Error.Sprintf("prompt cannot be empty"))
				return
			}
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

		fmt.Printf("Synthesizing speech: \"%s\"\n", textToSynthesize)

		result, err := p.Synthesize(synthesizeInput)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("TTS synthesis failed: %s", err))
			return
		}

		fmt.Printf("✓ Audio saved: %s\n", result.Result.AudioPath)
		printProfile(result.ProfileData)
		return
	}

	// Interactive mode: use repl
	repl(ReplConfig{
		ParseFile:     false,
		NoInteractive: false,

		Run: func(promptText string, _, _ []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			// Interactive mode: use prompt input
			textToSynthesize := strings.TrimSpace(promptText)
			if textToSynthesize == "" {
				return "", nexa_sdk.ProfileData{}, fmt.Errorf("text cannot be empty")
			}

			// Generate output filename
			outputFile := fmt.Sprintf("tts_output_%d.wav", time.Now().Unix())

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

			on_token(fmt.Sprintf("Synthesizing speech: \"%s\"\n", textToSynthesize))

			result, err := p.Synthesize(synthesizeInput)
			if err != nil {
				return "", result.ProfileData, err
			}

			output := fmt.Sprintf("✓ Audio saved: %s", result.Result.AudioPath)
			on_token(output)
			return output, result.ProfileData, nil
		},
	})
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
		DeviceID:  manifest.DeviceId,
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
		ParseFile:     true,
		NoInteractive: noInteractive,

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

				buffer := make([]float32, 512)
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

				fmt.Println()
				render.GetTheme().Reset()
				fmt.Println()
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
		DeviceID: manifest.DeviceId,
	}

	p, err := nexa_sdk.NewCV(cvInput)
	spin.Stop()

	if err != nil {
		slog.Error("failed to create CV", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	repl(ReplConfig{
		ParseFile:     true,
		NoInteractive: noInteractive,

		Run: func(_prompt string, images, _ []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			// Use images from prompt (repl handles non-interactive input parsing)
			if len(images) == 0 {
				return "", nexa_sdk.ProfileData{}, fmt.Errorf("image file is required for CV inference")
			}

			imagePath := images[0]
			if _, err := os.Stat(imagePath); os.IsNotExist(err) {
				return "", nexa_sdk.ProfileData{}, fmt.Errorf("input file '%s' does not exist", imagePath)
			}

			inferInput := nexa_sdk.CVInferInput{
				InputImagePath: imagePath,
			}

			on_token(fmt.Sprintf("Performing CV inference on image: %s\n", imagePath))

			result, err := p.Infer(inferInput)
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			output := fmt.Sprintf("✓ CV inference completed successfully\n")
			output += fmt.Sprintf("  Found %d results\n", len(result.Results))

			for _, cvResult := range result.Results {
				output += fmt.Sprintf("[%.3f] \"%s\"\n", cvResult.Confidence, cvResult.Text)
			}

			on_token(output)
			return output, nexa_sdk.ProfileData{}, nil
		},
	})
}

func inferImageGen(manifest *types.ModelManifest, _ string) {
	s := store.Get()
	modeldir := s.ModelfilePath(manifest.Name, "")

	spin := render.NewSpinner("loading ImageGen model...")
	spin.Start()
	p, err := nexa_sdk.NewImageGen(nexa_sdk.ImageGenCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: modeldir,
		PluginID:  manifest.PluginId,
		DeviceID:  manifest.DeviceId,
	})
	spin.Stop()
	if err != nil {
		slog.Error("failed to create ImageGen", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	repl(ReplConfig{
		ParseFile:     false,
		NoInteractive: noInteractive,

		Run: func(promptText string, _, _ []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			// Use prompt input (repl handles non-interactive input parsing)
			textToGenerate := strings.TrimSpace(promptText)
			if textToGenerate == "" {
				return "", nexa_sdk.ProfileData{}, fmt.Errorf("text prompt cannot be empty")
			}

			// Generate output filename
			outputFile := fmt.Sprintf("imagegen_output_%d.png", time.Now().Unix())

			on_token(fmt.Sprintf("Generating image: \"%s\"\n", textToGenerate))

			result, err := p.Txt2Img(nexa_sdk.ImageGenTxt2ImgInput{
				PromptUTF8: textToGenerate,
				Config: &nexa_sdk.ImageGenerationConfig{
					Prompts:         []string{textToGenerate},
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
				OutputPath: outputFile,
			})
			if err != nil {
				return "", nexa_sdk.ProfileData{}, err
			}

			output := fmt.Sprintf("✓ Image saved to: %s", result.OutputImagePath)
			on_token(output)
			return output, nexa_sdk.ProfileData{}, nil
		},
	})
}
