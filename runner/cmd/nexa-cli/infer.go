package main

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/render"
	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

const modelLoadFailMsg = `⚠️ Oops. Model failed to load.

👉 Try these:
- Verify your system meets the model’s requirements.
- Seek help in our discord or slack.`

var (
	// disableStream *bool // reuse in run.go
	modelType       string
	tool            []string
	prompt          []string
	query           string
	document        []string
	input           string
	output          string
	voiceIdentifier string
	speechSpeed     float64
	language        string
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{
		Use:   "infer <model-name>",
		Short: "Infer with a model",
		Long:  "Run inference with a specified model. The model must be downloaded and cached locally.",
	}

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	inferCmd.Flags().SortFlags = false
	inferCmd.Flags().StringVarP(&modelType, "model-type", "m", "llm", "specify model type [llm/vlm/embedder/reranker/tts/asr]")
	inferCmd.Flags().BoolVarP(&disableStream, "disable-stream", "s", false, "[llm|vlm] disable stream mode")
	inferCmd.Flags().StringArrayVarP(&tool, "tool", "t", nil, "[llm|vlm] add tool to make function call")
	inferCmd.Flags().StringArrayVarP(&prompt, "prompt", "p", nil, "[embedder|tts] pass prompt")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "[reranker] query")
	inferCmd.Flags().StringArrayVarP(&document, "document", "d", nil, "[reranker] documents")
	inferCmd.Flags().StringVarP(&input, "input", "i", "", "[asr] input file (audio for asr)")
	inferCmd.Flags().StringVarP(&output, "output", "o", "", "[tts] output file (audio for tts)")
	inferCmd.Flags().StringVarP(&voiceIdentifier, "voice-identifier", "", "", "[tts] voice identifier")
	inferCmd.Flags().Float64VarP(&speechSpeed, "speech-speed", "", 1.0, "[tts] speech speed (1.0 = normal)")
	inferCmd.Flags().StringVarP(&language, "language", "l", "", "[asr] language code (e.g., en, zh, ja)")

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

		switch modelType {
		case types.ModelTypeLLM:
			if !isVLM(manifest) {
				if len(tool) == 0 {
					inferLLM(modelfile, nil)
					return
				} else {
					panic("TODO")
				}
			} else {
				// compat vlm
				var t *string
				if manifest.MMProjFile.Name != "" {
					tokenizer := s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
					t = &tokenizer
				}
				inferVLM(modelfile, t)
			}
		case types.ModelTypeVLM:
			var t *string
			if manifest.MMProjFile.Name != "" {
				tokenizer := s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
				t = &tokenizer
			}
			inferVLM(modelfile, t)
		case types.ModelTypeEmbedder:
			inferEmbed(modelfile, nil)
		case types.ModelTypeReranker:
			inferRerank(modelfile, nil)
		case types.ModelTypeTTS:
			inferTTS(modelfile, nil)
		case types.ModelTypeASR:
			inferASR(modelfile, nil)
		default:
			panic("not support model type")
		}
	}
	return inferCmd
}

// isContainPreprocessor checks if the model has a preprocess.json file
func isVLM(m *types.ModelManifest) bool {
	if m.MMProjFile.Name != "" {
		return true
	}
	for _, file := range m.ExtraFiles {
		if strings.Contains(file.Name, "preprocessor") {
			return true
		}
	}
	return false
}

func inferLLM(model string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewLLM(model, tokenizer, 8192, nil)
	spin.Stop()

	if err != nil {
		if errors.Is(err, nexa_sdk.SDKErrorModelLoad) {
			fmt.Println(modelLoadFailMsg)
		} else {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		}
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.ChatMessage
	var lastLen int

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
				return "", err
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
				errCh <- e
				return
			}

			var full strings.Builder

			var promptToSend string
			if isMLX(model) {
				// fmt.Printf(text.FgBlack.Sprint(formatted[:lastLen]))
				// fmt.Printf(text.FgCyan.Sprint(formatted[lastLen:]))
				promptToSend = formatted[lastLen:]
			} else {
				promptToSend = formatted
			}

			dCh, eCh := p.GenerateStream(ctx, promptToSend)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			content := full.String()
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: content})
			lastLen = len(formatted) + len(content)
		},
	})
}

func inferVLM(model string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewVLM(model, tokenizer, 8192, nil)
	spin.Stop()
	if err != nil {
		if errors.Is(err, nexa_sdk.SDKErrorModelLoad) {
			fmt.Println(modelLoadFailMsg)
		} else {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		}
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.ChatMessage

	repl(ReplConfig{
		Stream:    !disableStream,
		ParseFile: true,

		Clear: p.Reset,

		GetProfilingData: func() (*nexa_sdk.ProfilingData, error) {
			return p.GetProfilingData()
		},

		Run: func(prompt string, images, audios []string) (string, error) {
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, err := p.ApplyChatTemplate(history, images, audios)
			if err != nil {
				return "", err
			}

			res, err := p.Generate(formatted, images, audios)
			if err != nil {
				return "", err
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res})

			return res, nil
		},

		RunStream: func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history, images, audios)
			if e != nil {
				errCh <- e
				return
			}

			var full strings.Builder

			dCh, eCh := p.GenerateStream(ctx, formatted, images, audios)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			content := full.String()
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: content})
		},
	})
}

func isMLX(model string) bool {
	pathParts := strings.Split(model, "/")
	encodedName := pathParts[len(pathParts)-2]
	nameBytes, _ := base64.StdEncoding.DecodeString(encodedName)
	name := strings.ToLower(string(nameBytes))
	isMLX := strings.Contains(name, "mlx")
	return isMLX
}

func inferEmbed(modelfile string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewEmbedder(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		if errors.Is(err, nexa_sdk.SDKErrorModelLoad) {
			fmt.Println(modelLoadFailMsg)
		} else {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
			fmt.Println()
		}
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
		if errors.Is(err, nexa_sdk.SDKErrorModelLoad) {
			fmt.Println(modelLoadFailMsg)
		} else {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		}
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
	spin := render.NewSpinner("loading model...")

	spin.Start()
	p, err := nexa_sdk.NewTTS(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	// Get text input - prioritize prompt, fallback to input file
	var inputText string
	if len(prompt) > 0 && prompt[0] != "" {
		inputText = prompt[0]
	} else {
		fmt.Println(text.FgRed.Sprintf("text is required for TTS synthesis (use --prompt)"))
		fmt.Println()
		return
	}

	// Configure TTS
	config := &nexa_sdk.TTSConfig{
		Voice:      voiceIdentifier,
		Speed:      float32(speechSpeed),
		Seed:       -1,
		SampleRate: 44100,
	}

	// Synthesize text to speech
	result, err := p.Synthesize(inputText, config)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	}

	if output != "" {
		err = saveWAV(result, output)
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error saving audio: %s", err))
		}
	} else {
		fmt.Println(text.FgRed.Sprintf("output file is required for TTS synthesis (use --output)"))
	}
	fmt.Println()

	if data, err := p.GetProfilingData(); err == nil {
		printProfiling(data)
	}
}

func inferASR(modelfile string, tokenizer *string) {
	spin := render.NewSpinner("loading model...")

	spin.Start()
	p, err := nexa_sdk.NewASR(modelfile, tokenizer, &language, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	// Check input file
	if input == "" {
		fmt.Println(text.FgRed.Sprintf("input audio file is required for ASR transcription"))
		fmt.Println()
		return
	}

	// Load audio file
	audio, sampleRate, err := loadWavFile(input)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error loading audio file: %s", err))
		return
	}

	// Configure ASR
	config := &nexa_sdk.ASRConfig{
		Timestamps: "word",
		BeamSize:   5,
		Stream:     false,
	}

	// Transcribe audio to text
	result, err := p.Transcribe(audio, int32(sampleRate), config)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	}

	if result != nil {
		fmt.Println(text.FgYellow.Sprint(result.Transcript))
	}
	fmt.Println()
}

// WAVHeader represents the WAV file header structure
type WAVHeader struct {
	Riff          [4]byte // "RIFF"
	FileSize      uint32  // File size - 8
	Wave          [4]byte // "WAVE"
	Fmt           [4]byte // "fmt "
	FmtSize       uint32  // Format chunk size
	AudioFormat   uint16  // Audio format (1 = PCM)
	NumChannels   uint16  // Number of channels
	SampleRate    uint32  // Sample rate
	ByteRate      uint32  // Byte rate
	BlockAlign    uint16  // Block align
	BitsPerSample uint16  // Bits per sample
	Data          [4]byte // "data"
	DataSize      uint32  // Data chunk size
}

// saveWAV saves the TTS result as a WAV file
func saveWAV(result *nexa_sdk.TTSResult, filename string) error {
	if result == nil || len(result.Audio) == 0 {
		return fmt.Errorf("no audio data to save")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create audio file: %v", err)
	}
	defer file.Close()

	// Prepare WAV header
	header := WAVHeader{}
	copy(header.Riff[:], "RIFF")
	copy(header.Wave[:], "WAVE")
	copy(header.Fmt[:], "fmt ")
	copy(header.Data[:], "data")

	header.FmtSize = 16
	header.AudioFormat = 1 // PCM
	header.NumChannels = uint16(result.Channels)
	header.SampleRate = uint32(result.SampleRate)
	header.BitsPerSample = 16
	header.ByteRate = uint32(result.SampleRate) * uint32(result.Channels) * uint32(header.BitsPerSample) / 8
	header.BlockAlign = uint16(result.Channels) * header.BitsPerSample / 8
	header.DataSize = uint32(len(result.Audio)) * uint32(result.Channels) * uint32(header.BitsPerSample) / 8
	header.FileSize = uint32(binary.Size(header)) - 8 + header.DataSize

	// Write header
	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		return fmt.Errorf("failed to write WAV header: %v", err)
	}

	// Convert float samples to 16-bit PCM and write
	pcmSamples := make([]int16, len(result.Audio)*int(result.Channels))
	for i, sample := range result.Audio {
		// Clamp audio values to [-1.0, 1.0] and convert to 16-bit PCM
		clampedSample := math.Max(-1.0, math.Min(1.0, float64(sample)))
		pcmSamples[i] = int16(clampedSample * 32767.0)
	}

	err = binary.Write(file, binary.LittleEndian, pcmSamples)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %v", err)
	}

	return nil
}

func loadWavFile(path string) ([]float32, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("could not open audio file: %w", err)
	}
	defer f.Close()

	// Read WAV header
	var header WAVHeader
	if err := binary.Read(f, binary.LittleEndian, &header); err != nil {
		return nil, 0, fmt.Errorf("could not read WAV header: %w", err)
	}

	// Validate WAV header
	if string(header.Riff[:]) != "RIFF" || string(header.Wave[:]) != "WAVE" {
		return nil, 0, errors.New("invalid WAV file format")
	}

	if header.AudioFormat != 1 {
		return nil, 0, errors.New("only PCM WAV files are supported")
	}

	// Calculate number of samples
	bytesPerSample := header.BitsPerSample / 8
	totalSamples := int(header.DataSize) / int(bytesPerSample)
	samplesPerChannel := totalSamples / int(header.NumChannels)

	// Read and convert audio data
	var samples []float32

	switch header.BitsPerSample {
	case 16:
		// 16-bit PCM
		intSamples := make([]int16, totalSamples)
		if err := binary.Read(f, binary.LittleEndian, &intSamples); err != nil {
			return nil, 0, fmt.Errorf("could not read 16-bit PCM data: %w", err)
		}

		// Convert to float and extract first channel
		samples = make([]float32, samplesPerChannel)
		for i := range samplesPerChannel {
			sampleIndex := i * int(header.NumChannels) // Take first channel
			samples[i] = float32(intSamples[sampleIndex]) / 32768.0
		}

	case 32:
		// 32-bit float PCM
		floatSamples := make([]float32, totalSamples)
		if err := binary.Read(f, binary.LittleEndian, &floatSamples); err != nil {
			return nil, 0, fmt.Errorf("could not read 32-bit float PCM data: %w", err)
		}

		// Extract first channel
		samples = make([]float32, samplesPerChannel)
		for i := range samplesPerChannel {
			sampleIndex := i * int(header.NumChannels) // Take first channel
			samples[i] = floatSamples[sampleIndex]
		}

	default:
		return nil, 0, fmt.Errorf("unsupported bits per sample: %d", header.BitsPerSample)
	}

	return samples, int(header.SampleRate), nil
}
