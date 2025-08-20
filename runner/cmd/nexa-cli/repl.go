package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/ollama/ollama/readline"

	"github.com/NexaAI/nexa-sdk/runner/internal/record"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

var help = [][2]string{
	{"/?, /h, /help", "Show this help message"},
	{"/exit", "Exit the REPL"},
	{"/clear", "Clear the screen and conversation history"},
	{"/load <filename>", "Load conversation history from a file"},
	{"/save <filename>", "Save conversation history to a file"},
	{"/mic", "Record audio for transcription"},
}

// LLM, VLM
type ReplConfig struct {
	ParseFile bool

	Reset       func() error
	SaveKVCache func(path string) error
	LoadKVCache func(path string) error

	Run func(prompt string, images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error)
}

func (cfg *ReplConfig) fill() {
	notSupport := fmt.Errorf("notSupport")

	if cfg.Reset == nil {
		cfg.Reset = func() error { return nil }
	}
	if cfg.SaveKVCache == nil {
		cfg.SaveKVCache = func(string) error { return notSupport }
	}
	if cfg.LoadKVCache == nil {
		cfg.LoadKVCache = func(string) error { return notSupport }
	}
}

func printProfile(pd nexa_sdk.ProfileData) {
	var text string

	if pd.AudioDuration > 0 { // ASR TTS
		text = fmt.Sprintf("processing_time %.2fs  |  audio_duration %.2fs  |  RTF %.2f (%.1fx realtime)",
			float64(pd.TotalTimeUs())/1e6,
			float64(pd.AudioDuration)/1e6,
			pd.RealTimeFactor,
			1.0/pd.RealTimeFactor)

	} else if pd.DecodingSpeed != 0 {
		text = fmt.Sprintf("— %.1f tok/s • %d tok • %.1f s first token -",
			pd.DecodingSpeed,
			pd.GeneratedTokens,
			float64(pd.TTFT)/1e6)

	} else {
		if pd.TotalTimeUs() != 0 {
			text = fmt.Sprintf("- %.1f s -",
				float64(pd.TotalTimeUs())/1e6,
			)
		}
	}

	fmt.Print(render.GetTheme().Profile.Sprint(text))
	fmt.Println()
	fmt.Println()
}

type MultilineState int

const (
	MultilineNone MultilineState = iota
	MultilinePrompt
)

func repl(cfg ReplConfig) {
	cfg.fill()

	l, err := readline.New(readline.Prompt{
		Prompt:         render.GetTheme().Prompt.Sprint("> "),
		AltPrompt:      render.GetTheme().Prompt.Sprint(". "),
		Placeholder:    "Send a message, press /? for help",
		AltPlaceholder: `Use """ to end multi-line input`,
	})
	if err != nil {
		panic(err)
	}
	// defer l.Close()

	fmt.Print(readline.StartBracketedPaste)
	defer fmt.Printf(readline.EndBracketedPaste)

	var cancel func()
	cSignal := make(chan os.Signal, 1)
	signal.Notify(cSignal, os.Interrupt)
	go func() {
		for range cSignal {
			if cancel != nil {
				cancel()
			}
		}
	}()

	var sb strings.Builder
	var multiline MultilineState
	var recordAudios []string

	for {
		// print stashed content
		if multiline == MultilineNone && recordAudios != nil {
			fmt.Println(render.GetTheme().Info.Sprintf("Current stash audios: %s", strings.Join(recordAudios, ", ")))
		}

		line, err := l.Readline()

		// check err or exit
		switch {
		case errors.Is(err, io.EOF):
			fmt.Println()
			return
		case errors.Is(err, readline.ErrInterrupt):
			if line == "" {
				fmt.Println("\nUse Ctrl + d or /exit to exit.")
				fmt.Println()
			}
			l.Prompt.UseAlt = false
			sb.Reset()
			continue
		case err != nil:
			return
		}

		// check multiline state and paste state
		switch {
		case multiline != MultilineNone:
			// check if there's a multiline terminating string
			before, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(before)
			if !ok {
				fmt.Fprintln(&sb)
				continue
			}

			multiline = MultilineNone
			l.Prompt.UseAlt = false

		case strings.HasPrefix(line, `"""`):
			line := strings.TrimPrefix(line, `"""`)
			line, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(line)
			if !ok {
				// no multiline terminating string; need more input
				fmt.Fprintln(&sb)
				multiline = MultilinePrompt
				l.Prompt.UseAlt = true
			}

		case l.Pasting:
			fmt.Fprintln(&sb, line)
			continue

		default:
			sb.WriteString(line)
		}

		// empty input or multiline state
		if (sb.Len() == 0 && len(recordAudios) == 0) ||
			multiline != MultilineNone {
			continue
		}

		// read input
		line = sb.String()
		sb.Reset()

		// parse file
		var images, audios []string
		if cfg.ParseFile {
			line, images, audios = parseFiles(line)
		}

		// check if it's a command
		if len(images) == 0 && len(audios) == 0 && strings.HasPrefix(line, "/") {

			fileds := strings.Fields(strings.TrimSpace(line))

			switch fileds[0] {
			case "/?", "/h", "/help":
				fmt.Println("Commands:")
				for _, h := range help {
					fmt.Printf("  %-25s %s\n", h[0], h[1])
				}
				fmt.Println()

			case "/exit":
				return

			case "/clear":
				cfg.Reset()
				recordAudios = nil
				fmt.Print("\033[H\033[2J")

			case "/load":
				if len(fileds) != 2 {
					fmt.Println(render.GetTheme().Error.Sprintf("Usage: /load <filename>"))
					fmt.Println()
					continue
				}
				cfg.Reset()
				err := cfg.LoadKVCache(fileds[1])
				if err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					fmt.Println()
				}

			case "/save":
				if len(fileds) != 2 {
					fmt.Println(render.GetTheme().Error.Sprintf("Usage: /save <filename>"))
					fmt.Println()
					continue
				}
				err := cfg.SaveKVCache(fileds[1])
				if err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					fmt.Println()
				}

			case "/mic":
				fmt.Println(render.GetTheme().Info.Sprint("Recording is going on, press Ctrl-C to stop"))

				t := strconv.Itoa(int(time.Now().Unix()))
				outputFile := filepath.Join(os.TempDir(), "nexa-cli", t+".wav")
				rec, err := record.NewRecorder(outputFile)
				if err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					fmt.Println()
					continue
				}
				if err = rec.Run(); err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Failed to start recording: %s", err))
					fmt.Println()
					continue
				}

				recordAudios = append(recordAudios, rec.GetOutputFile())
				fmt.Println()

			default:
				fmt.Println(render.GetTheme().Error.Sprintf("Unknown command: %s", fileds[0]))
				fmt.Println()
			}

			continue
		}

		// run async
		ctx, cancelFunc := context.WithCancel(context.Background())
		cancel = cancelFunc

		firstToken := true
		spin := render.NewSpinner("encoding...")
		spin.Start()

		audios = append(audios, recordAudios...)
		recordAudios = nil // clear after use

		_, profileData, err := cfg.Run(line, images, audios, func(token string) bool {
			if firstToken {
				spin.Stop()
				firstToken = false
			}

			switch token {
			case "<think>":
				render.GetTheme().Set(render.GetTheme().ThinkOutput)
				fmt.Print(token)
			case "</think>":
				fmt.Print(token)
				render.GetTheme().Set(render.GetTheme().ModelOutput)
			case string("\x20\xef\xbf\xbd"):
				fmt.Print("\x20")
			case string("\xef\xbf\xbd"):
			default:
				fmt.Print(token)
			}

			return ctx.Err() == nil
		})

		render.GetTheme().Reset()
		fmt.Println()
		fmt.Println()
		printProfile(profileData)

		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("Error: %s\n", err))
			fmt.Println()
			return
		}
	}
}

// =============== file name parse ===============

var fileRegex = regexp.MustCompile(`(?:[a-zA-Z]:)?(?:\./|/|\\)[\S\\ ]+?\.(?i:jpg|jpeg|png|webp|mp3|wav)\b`)
var partRegex = regexp.MustCompile(`-\d+-of-\d+\.gguf$`)

func parseFiles(prompt string) (string, []string, []string) {
	files := fileRegex.FindAllString(prompt, -1)
	images := make([]string, 0, len(files))
	audios := make([]string, 0, len(files))

	for _, file := range files {
		realFile := strings.NewReplacer(
			"\\ ", " ",
			"\\(", "(",
			"\\)", ")",
			"\\[", "[",
			"\\]", "]",
			"\\{", "{",
			"\\}", "}",
			"\\$", "$",
			"\\&", "&",
			"\\;", ";",
			"\\'", "'",
			"\\\\", "\\",
			"\\*", "*",
			"\\?", "?",
			"\\~", "~",
		).Replace(file)

		_, err := os.Stat(realFile)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("parse file error: [%s] %s", realFile, err))
			continue
		}
		switch realFile[len(realFile)-3:] {
		case "mp3", "wav":
			audios = append(audios, realFile)
			slog.Debug("add audio", "file", realFile)
		default:
			images = append(images, realFile)
			slog.Debug("add image", "file", realFile)
		}

		prompt = strings.ReplaceAll(prompt, "'"+realFile+"'", "")
		prompt = strings.ReplaceAll(prompt, "'"+file+"'", "")
		prompt = strings.ReplaceAll(prompt, file, "")
	}
	return strings.TrimSpace(prompt), images, audios
}

// =============== quant name parse ===============
var quantRegix = regexp.MustCompile(`(` + strings.Join([]string{
	"[fF][pP][0-9]+",                 // FP32, FP16, FP64
	"[fF][0-9]+",                     // F64, F32, F16
	"[iI][0-9]+",                     // I64, I32, I16, I8
	"[qQ][0-9]+(_[A-Za-z0-9]+)*",     // Q8_0, Q8_1, Q8_K, Q6_K, Q5_0, Q5_1, Q5_K, Q4_0, Q4_1, Q4_K, Q3_K, Q2_K
	"[iI][qQ][0-9]+(_[A-Za-z0-9]+)*", // IQ4_NL, IQ4_XS, IQ3_S, IQ3_XXS, IQ2_XXS, IQ2_S, IQ2_XS, IQ1_S, IQ1_M
	"[bB][fF][0-9]+",                 // BF16
	"[0-9]+[bB][iI][tT]",             // 1bit, 2bit, 3bit, 4bit, 16bit, 1BIT, 16Bit, etc.
}, "|") + `)`)

// order big to small
func quantGreaterThan(a, b string, order []string) bool {
	// empty
	if a == "" || b == "" {
		return a != ""
	}

	a = strings.ToUpper(a)
	b = strings.ToUpper(b)

	// same
	if a == b {
		return false
	}

	// order
	ca := slices.Index(order, a)
	cb := slices.Index(order, b)
	if ca >= 0 && cb >= 0 {
		return ca < cb
	} else if ca >= 0 || cb >= 0 {
		return ca >= 0
	}

	// normal
	if a[0] == b[0] {
		return a > b
	} else {
		return a[0] == 'F'
	}
}

// getFileSizesConcurrent fetches file sizes concurrently with a limit of 8 concurrent requests
func getFileSizesConcurrent(name string, files []string) (map[string]int64, error) {
	fileSizes := make(map[string]int64, len(files))
	if len(files) == 0 {
		return fileSizes, nil
	}

	// Create semaphore to limit concurrent requests to 8
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup
	var firstError error
	var errorMutex sync.Mutex

	for i, file := range files {
		wg.Add(1)
		// Acquire semaphore
		sem <- struct{}{}

		go func(index int, filename string) {
			defer wg.Done()
			defer func() { <-sem }()

			size, err := store.Get().HFFileSize(context.TODO(), name, filename)

			errorMutex.Lock()
			fileSizes[filename] = size
			if err != nil && firstError == nil {
				firstError = err
			}
			errorMutex.Unlock()
		}(i, file)
	}

	wg.Wait()

	return fileSizes, firstError
}

func chooseModelType() (types.ModelType, error) {
	var modelType types.ModelType
	if err := huh.NewSelect[types.ModelType]().
		Title("Choose Model Type").
		Options(huh.NewOptions(
			types.ModelTypeLLM, types.ModelTypeVLM, types.ModelTypeEmbedder, types.ModelTypeReranker,
			types.ModelTypeASR, types.ModelTypeTTS, types.ModelTypeCV, types.ModelTypeImageGen)...).
		Value(&modelType).
		Run(); err != nil {
		return "", err
	}
	return modelType, nil
}

func chooseModelTypeByName(modelName string) (types.ModelType, error) {
	// Hardcoded model type mapping for specific models
	switch modelName {
	case "nexaml/qwen3-npu", "nexaml/qwen3-4B-npu", "nexaml/qwen3-1.7B-npu-encrypt", "nexaml/qwen3-4B-npu-encrypt", "NexaAI/qwen3-1.7B-npu", "NexaAI/qwen3-4B-npu":
		return types.ModelTypeLLM, nil
	case "nexaml/omni-neural", "nexaml/omni-neural-npu-encrypt", "NexaAI/OmniNeural-4B":
		return types.ModelTypeVLM, nil
	case "nexaml/paddleocr-npu", "nexaml/yolov12-npu", "nexaml/paddleocr-npu-encrypt", "nexaml/yolov12-npu-encrypt", "NexaAI/paddleocr-npu", "NexaAI/yolov12-npu":
		return types.ModelTypeCV, nil
	default:
		// Fallback to interactive selection for unknown models
		return chooseModelType()
	}
}

func chooseFiles(name string, files []string) (res types.ModelManifest, err error) {
	if len(files) == 0 {
		err = fmt.Errorf("repo is empty")
		return
	}

	spin := render.NewSpinner("loading model size...")
	res.Name = name
	res.ModelFile = make(map[string]types.ModeFileInfo)

	// TODO: refactor
	// check gguf
	var mmprojs []string
	ggufGroups := make(map[string][]string)
	// qwen2.5-7b-instruct-q8_0-00003-of-00003.gguf original name is qwen2.5-7b-instruct-q8_0
	// *d-of-*d like this
	for _, file := range files {
		lower := strings.ToLower(file)
		if strings.HasSuffix(lower, ".gguf") {
			if strings.HasPrefix(lower, "mmproj") {
				mmprojs = append(mmprojs, file)
			} else {
				name := partRegex.ReplaceAllString(file, "")
				ggufGroups[name] = append(ggufGroups[name], file)
			}
		}
	}

	ggufs := make([]string, 0, len(ggufGroups))
	for gguf := range ggufGroups {
		ggufs = append(ggufs, gguf)
	}

	// choose model file
	if len(ggufs) > 0 {
		// detect gguf
		if len(ggufs) == 1 {
			// single quant
			fileInfo := types.ModeFileInfo{}
			fileInfo.Name = ggufs[0]
			fileInfo.Downloaded = true
			spin.Start()
			fileSizes, err := getFileSizesConcurrent(name, ggufGroups[ggufs[0]])
			spin.Stop()
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: [%s] %s", ggufs[0], err))
				return res, err
			}
			for _, size := range fileSizes {
				fileInfo.Size += size
			}

			quant := strings.ToUpper(quantRegix.FindString(ggufs[0]))
			if quant == "" {
				quant = "N/A"
			}
			res.ModelFile[quant] = fileInfo

		} else {
			// interactive choose
			// Get file sizes for display
			spin.Start()
			// key is gguf file name, value is file size total containts part file
			fileSizes := make(map[string]int64)
			for _, gguf := range ggufs {
				sizes, err := getFileSizesConcurrent(name, ggufGroups[gguf])
				if err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: [%s] %s", gguf, err))
					return res, err
				}
				for _, size := range sizes {
					fileSizes[gguf] += size
				}
			}
			spin.Stop()

			// select default gguf
			var file, quant string
			for _, gguf := range ggufs {
				ggufQuant := quantRegix.FindString(gguf)
				if quantGreaterThan(ggufQuant, quant, []string{"Q4_K_M", "Q4_0", "Q8_0"}) {
					quant = ggufQuant
					file = gguf
				}
			}

			// Find the longest quant name for alignment
			options := make([]huh.Option[string], 0, len(ggufs)+1)
			if file != "" {
				sizeStr := humanize.IBytes(uint64(fileSizes[file]))
				options = append(options, huh.NewOption(
					fmt.Sprintf("%-10s [%7s] (default)", strings.ToUpper(quant), sizeStr), file,
				))
			}
			for i := range ggufs {
				quant := strings.ToUpper(quantRegix.FindString(ggufs[i]))
				if quant != "" && file != ggufs[i] {
					sizeStr := humanize.IBytes(uint64(fileSizes[ggufs[i]]))
					options = append(options, huh.NewOption(
						fmt.Sprintf("%-10s [%7s]", quant, sizeStr), ggufs[i],
					))
				}
			}

			if len(options) == 0 {
				err = fmt.Errorf("no valid gguf found")
				return res, err
			}

			if err = huh.NewSelect[string]().
				Title("Choose a quant version to download").
				Options(options...).
				Value(&file).
				Run(); err != nil {
				return res, err
			}

			for k := range ggufGroups {
				downloaded := k == file
				quant := strings.ToUpper(quantRegix.FindString(k))
				// sort files by name
				files := ggufGroups[k]
				slices.Sort(files)
				res.ModelFile[quant] = types.ModeFileInfo{
					Name:       files[0],
					Downloaded: downloaded,
					Size:       fileSizes[k],
				}
				for _, file := range files[1:] {
					res.ExtraFiles = append(res.ExtraFiles, types.ModeFileInfo{
						Name:       file,
						Downloaded: downloaded,
					})
				}
			}
		}

		// detect mmproj
		switch len(mmprojs) {
		case 0:
		case 1:
			res.MMProjFile.Name = mmprojs[0]
			spin.Start()
			size, err := store.Get().HFFileSize(context.TODO(), name, mmprojs[0])
			spin.Stop()
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: [%s] %s", mmprojs[0], err))
				return res, err
			}
			res.MMProjFile.Size = size
			res.MMProjFile.Downloaded = true

		default:
			// Get mmproj file sizes for display
			spin.Start()
			mmprojSizes, err := getFileSizesConcurrent(name, mmprojs)
			spin.Stop()
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: %s", err))
				return res, err
			}

			// match biggest
			var file string
			var size int64
			for _, mmproj := range mmprojs {
				if mmprojSizes[mmproj] > size {
					file = mmproj
				}
			}

			res.MMProjFile.Name = file
			res.MMProjFile.Size = mmprojSizes[file]
			res.MMProjFile.Downloaded = true
		}
	} else {
		// other format

		// quant
		var quant string
		if q := strings.ToUpper(quantRegix.FindString(name)); q != "" {
			quant = q
		} else if q, err := store.Get().GetQuantInfo(context.TODO(), name); err == nil && q != 0 {
			quant = fmt.Sprintf("%dBIT", q)
		} else {
			quant = "N/A"
		}

		// detect main model file
		// add other files
		for _, file := range files {
			if res.ModelFile[quant].Name == "" {
				lower := strings.ToLower(file)
				if strings.HasSuffix(lower, "safetensors") || strings.HasSuffix(lower, "npz") {
					res.ModelFile[quant] = types.ModeFileInfo{Name: file}
					continue
				}
			}
			res.ExtraFiles = append(res.ExtraFiles, types.ModeFileInfo{Name: file})
		}

		// fallback to first file
		if res.ModelFile[quant].Name == "" {
			res.ModelFile[quant] = types.ModeFileInfo{Name: files[0]}
			res.ExtraFiles = res.ExtraFiles[1:]
		}

		spin.Start()
		sizes, err := getFileSizesConcurrent(name, files)
		spin.Stop()
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("get filesize error: %s", err))
			return res, err
		}

		res.ModelFile[quant] = types.ModeFileInfo{
			Name:       res.ModelFile[quant].Name,
			Downloaded: true,
			Size:       sizes[res.ModelFile[quant].Name],
		}
		for i, v := range res.ExtraFiles {
			res.ExtraFiles[i] = types.ModeFileInfo{
				Name:       v.Name,
				Downloaded: true,
				Size:       sizes[v.Name],
			}
		}
	}

	return
}

func chooseQuantFiles(old types.ModelManifest) (*types.ModelManifest, error) {
	var mf types.ModelManifest
	d, _ := sonic.Marshal(old)
	sonic.Unmarshal(d, &mf)
	// Find the longest quant name for alignment
	options := make([]huh.Option[string], 0, len(mf.ModelFile))
	for q, m := range mf.ModelFile {
		if !m.Downloaded {
			options = append(options, huh.NewOption(
				fmt.Sprintf("%-10s [%7s]", q, humanize.IBytes(uint64(m.Size))), q,
			))
		}
	}

	var quant string
	if err := huh.NewSelect[string]().
		Title("Choose a quant version to download").
		Options(options...).
		Value(&quant).
		Run(); err != nil {
		return nil, err
	}

	mf.ModelFile[quant] = types.ModeFileInfo{
		Name:       mf.ModelFile[quant].Name,
		Downloaded: true,
		Size:       mf.ModelFile[quant].Size,
	}

	file := mf.ModelFile[quant].Name
	ggufName := partRegex.ReplaceAllString(file, "")
	for i, f := range mf.ExtraFiles {
		if ggufName == partRegex.ReplaceAllString(file, "") {
			mf.ExtraFiles[i] = types.ModeFileInfo{
				Name:       f.Name,
				Downloaded: true,
			}
		}

	}

	return &mf, nil
}
