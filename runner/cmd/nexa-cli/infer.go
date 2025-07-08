package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/render"
	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

var (
	// disableStream *bool // reuse in run.go
	modelType string
	tool      []string
	prompt    []string
	query     string
	document  []string
	ttsOutput string
	ttsVoice  string
	ttsSpeed  float64
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{
		Use:   "infer <model-name>",
		Short: "Infer with a model",
		Long:  "Run inference with a specified model. The model must be downloaded and cached locally.",
	}

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	inferCmd.Flags().SortFlags = false
	inferCmd.Flags().StringVarP(&modelType, "model-type", "m", "llm", "specify model type [llm/vlm/embedder/reranker/tts]")
	inferCmd.Flags().BoolVarP(&disableStream, "disable-stream", "s", false, "[llm|vlm] disable stream mode")
	inferCmd.Flags().StringSliceVarP(&tool, "tool", "t", nil, "[llm|vlm] add tool to make function call")
	inferCmd.Flags().StringSliceVarP(&prompt, "prompt", "p", nil, "[embedder] pass prompt")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "[reranker] query")
	inferCmd.Flags().StringSliceVarP(&document, "document", "d", nil, "[reranker] documents")
	inferCmd.Flags().StringVarP(&ttsOutput, "output", "o", "output.wav", "[tts] output audio file")
	inferCmd.Flags().StringVarP(&ttsVoice, "voice", "", "", "[tts] voice identifier")
	inferCmd.Flags().Float64VarP(&ttsSpeed, "speed", "", 1.0, "[tts] speech speed (1.0 = normal)")

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

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile)

		switch modelType {
		case types.ModelTypeLLM:
			if manifest.MMProjFile == "" {
				if len(tool) == 0 {
					inferLLM(modelfile, nil)
					return
				} else {
					panic("TODO")
				}
			} else {
				// compat vlm
				t := s.ModelfilePath(manifest.Name, manifest.MMProjFile)
				inferVLM(modelfile, &t)
			}
		case types.ModelTypeVLM:
			t := s.ModelfilePath(manifest.Name, manifest.MMProjFile)
			inferVLM(modelfile, &t)
		case types.ModelTypeEmbedder:
			inferEmbed(modelfile, nil)
		case types.ModelTypeReranker:
			inferRerank(modelfile, nil)
		case types.ModelTypeTTS:
			inferTTS(modelfile, nil)
		default:
			panic("not support model type")
		}
	}
	return inferCmd
}

func inferLLM(model string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewLLM(model, tokenizer, 8192, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.ChatMessage

	repl(ReplConfig{
		Stream:    !disableStream,
		ParseFile: false,

		Clear: p.Reset,

		SaveKVCache: func(path string) error {
			return p.SaveKVCache(path)
		},

		LoadKVCache: func(path string) error {
			return p.LoadKVCache(path)
		},

		GetProfilingData: func() (*nexa_sdk.ProfilingData, error) {
			return p.GetProfilingData()
		},

		Run: func(prompt string, _, _ []string) (string, error) {
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})

			formatted, err := p.ApplyChatTemplate(history)
			if err != nil {
				if errors.Is(err, nexa_sdk.ErrChatTemplateNotFound) {
					// Chat template can be not found for some non-instruct-tuned models, we directly use the original prompt in those cases.
					formatted = prompt
					err = nil
				} else {
					return "", err
				}
			}

			res, err := p.Generate(formatted)
			if err != nil {
				return "", err
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res})

			return res, nil
		},

		RunStream: func(ctx context.Context, prompt string, _, _ []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				if errors.Is(e, nexa_sdk.ErrChatTemplateNotFound) {
					// Chat template can be not found for some non-instruct-tuned models, we directly use the original prompt in those cases.
					formatted = prompt
					e = nil
				} else {
					errCh <- e
					return
				}
			}

			var full strings.Builder
			// fmt.Printf(text.FgBlack.Sprint(formatted[:lastLen]))
			// fmt.Printf(text.FgCyan.Sprint(formatted[lastLen:]))
			dCh, eCh := p.GenerateStream(ctx, formatted)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})
		},
	})
}

func inferVLM(model string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewVLM(model, tokenizer, 8192, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.ChatMessage
	var lastLen int

	repl(ReplConfig{
		Stream:    !disableStream,
		ParseFile: true,

		Clear: p.Reset,

		GetProfilingData: func() (*nexa_sdk.ProfilingData, error) {
			return p.GetProfilingData()
		},

		Run: func(prompt string, images, audios []string) (string, error) {
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, err := p.ApplyChatTemplate(history)
			if err != nil {
				return "", err
			}

			res, err := p.Generate(prompt, images, audios)
			if err != nil {
				return "", err
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res})
			lastLen = len(formatted) + len(res)

			return res, nil
		},

		RunStream: func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			// fmt.Println(text.FgBlack.Sprint(prompt))

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}

			var full strings.Builder
			dCh, eCh := p.GenerateStream(ctx, formatted[lastLen:], images, audios)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})
			lastLen = len(formatted) + len(full.String())
		},
	})
}

func inferEmbed(modelfile string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewEmbedder(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	}
	defer p.Destroy()

	if len(prompt) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 text prompt is accept"))
		fmt.Println()
		return
	}

	res, err := p.Embed(prompt)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	} else {
		nEmbed := len(res) / len(prompt)
		for i := range res {
			if i%nEmbed == 0 {
				fmt.Print(text.FgYellow.Sprintf("\n===> %d\n", i/nEmbed))
			}
			fmt.Print(text.FgYellow.Sprintf("%f ", res[i]))
		}
		fmt.Println()
	}
	fmt.Println()

	if data, err := p.GetProfilingData(); err == nil {
		printProfiling(data)
	}
}

func inferRerank(modelfile string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewReranker(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	if len(query) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 query is accept"))
		fmt.Println()
		return
	}
	if len(document) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 document is accept"))
		fmt.Println()
		return
	}

	res, err := p.Rerank(query, document)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	} else {
		fmt.Println()
		for i := range res {
			fmt.Println(text.FgYellow.Sprintf("%d => %f", i, res[i]))
		}
		fmt.Println()
	}

	if data, err := p.GetProfilingData(); err == nil {
		printProfiling(data)
	}
}

func inferTTS(modelfile string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))

	fmt.Println("[DEBUG] inferTTS", modelfile, tokenizer)

	spin.Start()
	p, err := nexa_sdk.NewTTS(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	// Get text input
	inputText := prompt[0]
	if inputText == "" {
		fmt.Println(text.FgRed.Sprintf("text is required for TTS synthesis"))
		fmt.Println()
		return
	}

	// Configure TTS
	config := &nexa_sdk.TTSConfig{
		Voice:      ttsVoice,
		Speed:      float32(ttsSpeed),
		Seed:       -1,
		SampleRate: 22050,
	}

	fmt.Printf("Synthesizing text: %s\n", inputText)
	if ttsVoice != "" {
		fmt.Printf("Using voice: %s\n", ttsVoice)
	}
	fmt.Printf("Speech speed: %.2f\n", ttsSpeed)

	// Synthesize text to speech
	result, err := p.Synthesize(inputText, config)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	}

	if result != nil {
		fmt.Printf("TTS synthesis completed: %d samples, %.2f seconds, %d Hz, %d channels\n",
			result.NumSamples, result.DurationSeconds, result.SampleRate, result.Channels)

		// Save audio to file
		err = saveTTSResult(result, ttsOutput)
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error saving audio: %s", err))
		} else {
			fmt.Printf("Audio saved to: %s\n", ttsOutput)
		}
	}
	fmt.Println()

	if data, err := p.GetProfilingData(); err == nil {
		printProfiling(data)
	}
}

// saveTTSResult saves TTS audio result to a WAV file
func saveTTSResult(result *nexa_sdk.TTSResult, filename string) error {
	if result == nil || len(result.Audio) == 0 {
		return fmt.Errorf("no audio data to save")
	}

	// For now, we'll save as a simple binary file
	// In a real implementation, we'd format this as WAV
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write basic WAV header (simplified)
	header := make([]byte, 44)
	// "RIFF" chunk identifier
	copy(header[0:4], "RIFF")
	// File size - 8 bytes
	fileSize := uint32(36 + len(result.Audio)*4)
	header[4] = byte(fileSize)
	header[5] = byte(fileSize >> 8)
	header[6] = byte(fileSize >> 16)
	header[7] = byte(fileSize >> 24)
	// "WAVE" format
	copy(header[8:12], "WAVE")
	// "fmt " sub-chunk
	copy(header[12:16], "fmt ")
	// Sub-chunk size (16 for PCM)
	header[16] = 16
	// Audio format (1 for PCM)
	header[20] = 1
	// Number of channels
	header[22] = byte(result.Channels)
	// Sample rate
	sampleRate := uint32(result.SampleRate)
	header[24] = byte(sampleRate)
	header[25] = byte(sampleRate >> 8)
	header[26] = byte(sampleRate >> 16)
	header[27] = byte(sampleRate >> 27)
	// Byte rate
	byteRate := sampleRate * uint32(result.Channels) * 4
	header[28] = byte(byteRate)
	header[29] = byte(byteRate >> 8)
	header[30] = byte(byteRate >> 16)
	header[31] = byte(byteRate >> 24)
	// Block align
	header[32] = byte(result.Channels * 4)
	// Bits per sample
	header[34] = 32
	// "data" sub-chunk
	copy(header[36:40], "data")
	// Data size
	dataSize := uint32(len(result.Audio) * 4)
	header[40] = byte(dataSize)
	header[41] = byte(dataSize >> 8)
	header[42] = byte(dataSize >> 16)
	header[43] = byte(dataSize >> 24)

	// Write header
	_, err = file.Write(header)
	if err != nil {
		return err
	}

	// Write audio data as 32-bit float (little endian)
	for _, sample := range result.Audio {
		bytes := make([]byte, 4)
		bits := *(*uint32)(unsafe.Pointer(&sample))
		bytes[0] = byte(bits)
		bytes[1] = byte(bits >> 8)
		bytes[2] = byte(bits >> 16)
		bytes[3] = byte(bits >> 24)
		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
	}

	return nil
}
