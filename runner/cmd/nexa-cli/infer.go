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
	// disableStream *bool // reuse in run.go
	ngl            int32
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
		llmFlags.Int32VarP(&ngl, "ngl", "n", 999, "num of layers pass to gpu")
		llmFlags.Int32VarP(&maxTokens, "max-tokens", "", 2048, "max tokens")
		llmFlags.BoolVarP(&enableThink, "think", "", true, "enable thinking mode")
		llmFlags.BoolVarP(&hideThink, "hide-think", "", false, "hide thinking output")
		llmFlags.StringVarP(&systemPrompt, "system-prompt", "s", "", "system prompt to set model behavior")
		llmFlags.StringVarP(&input, "input", "i", "", "prompt txt file")
		return llmFlags
	}()
	vlmFlags = func() *pflag.FlagSet {
		vlmFlags := pflag.NewFlagSet("VLM Specific", pflag.ExitOnError)
		vlmFlags.SortFlags = false
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
	embedderFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
	embedderFlags.StringVarP(&taskType, "task-type", "", "default", "default|search_query|search_document")
	inferCmd.Flags().AddFlagSet(embedderFlags)

	rerankerFlags := pflag.NewFlagSet("Reranker", pflag.ExitOnError)
	rerankerFlags.SortFlags = false
	rerankerFlags.StringVarP(&query, "query", "q", "", "query")
	rerankerFlags.StringArrayVarP(&document, "document", "d", nil, "documents")
	inferCmd.Flags().AddFlagSet(rerankerFlags)

	cvFlags := pflag.NewFlagSet("CV", pflag.ExitOnError)
	cvFlags.SortFlags = false
	cvFlags.StringVarP(&input, "input", "i", "", "input image file")
	inferCmd.Flags().AddFlagSet(cvFlags)

	ttsFlags := pflag.NewFlagSet("TTS", pflag.ExitOnError)
	ttsFlags.SortFlags = false
	ttsFlags.StringArrayVarP(&prompt, "prompt", "p", nil, "pass prompt")
	ttsFlags.StringVarP(&voice, "voice", "", "", "voice identifier")
	ttsFlags.BoolVarP(&listVoice, "list-voice", "", false, "list available voices")
	ttsFlags.Float64VarP(&speechSpeed, "speech-speed", "", 1.0, "speech speed (1.0 = normal)")
	ttsFlags.StringVarP(&output, "output", "o", "", "output audio file")
	inferCmd.Flags().AddFlagSet(ttsFlags)

	imageGenFlags := pflag.NewFlagSet("ImageGen", pflag.ExitOnError)
	imageGenFlags.SortFlags = false
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
			return nil, fmt.Errorf("Download model failed")
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
			NCtx:         4096,
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
	if len(input) > 0 {
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
			NCtx:         4096,
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
	if len(input) > 0 {
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

	repl(ReplConfig{
		ParseFile: false,

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
			if err != nil || len(result.Embeddings) == 0 {
				return "", result.ProfileData, err
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
		DeviceID:  manifest.DeviceId,
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
