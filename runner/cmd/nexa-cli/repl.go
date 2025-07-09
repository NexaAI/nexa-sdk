package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/ollama/ollama/readline"

	"github.com/NexaAI/nexa-sdk/internal/render"
	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

//var completer = readline.NewPrefixCompleter(
//	readline.PcItem("/?"),
//	readline.PcItem("/h"),
//	readline.PcItem("/help"),
//
//	readline.PcItem("/exit"),
//
//	readline.PcItem("/clear"),
//	readline.PcItem("/load", readline.PcItemDynamic(listFiles("./"))),
//	readline.PcItem("/save", readline.PcItemDynamic(listFiles("./"))),
//)

var help = [][2]string{
	{"/?, /h, /help", "Show this help message"},
	{"/exit", "Exit the REPL"},
	{"/clear", "Clear the screen and conversation history"},
	{"/load <filename>", "Load conversation history from a file"},
	{"/save <filename>", "Save conversation history to a file"},
}

// TODO: support sub dir
//func listFiles(path string) func(string) []string {
//	return func(line string) []string {
//		names := make([]string, 0)
//		files, _ := os.ReadDir(path)
//		for _, f := range files {
//			names = append(names, f.Name())
//		}
//		return names
//	}
//}

// LLM, VLM
type ReplConfig struct {
	Stream    bool
	ParseFile bool

	Clear       func()
	SaveKVCache func(path string) error
	LoadKVCache func(path string) error

	Run       func(prompt string, images, audios []string) (string, error)
	RunStream func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error)

	GetProfilingData func() (*nexa_sdk.ProfilingData, error)
}

func (cfg *ReplConfig) fill() {
	notSupport := fmt.Errorf("notSupport")

	if cfg.Clear == nil {
		cfg.Clear = func() {}
	}
	if cfg.SaveKVCache == nil {
		cfg.SaveKVCache = func(string) error { return notSupport }
	}
	if cfg.LoadKVCache == nil {
		cfg.LoadKVCache = func(string) error { return notSupport }
	}
	if cfg.Run == nil {
		cfg.Run = func(string, []string, []string) (string, error) { return "", notSupport }
	}
	if cfg.RunStream == nil {
		cfg.RunStream = func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error) {
			close(dataCh)
			errCh <- notSupport
			close(errCh)
		}
	}
}

func printProfiling(profilingData *nexa_sdk.ProfilingData) {
	if profilingData == nil {
		return
	}

	if profilingData.TokensPerSecond != 0 {
		profilingText := fmt.Sprintf("%.2f tok/sec • %d tokens • %.2fs to first token • Stop reason: %s",
			profilingData.TokensPerSecond,
			profilingData.GeneratedTokens,
			profilingData.TTFTMs/1e3,
			strings.ToUpper(profilingData.StopReason))

		fmt.Print(text.FgBlue.Sprint(profilingText))
		fmt.Println(text.Reset.EscapeSeq())
		fmt.Println()
		return
	}

	if profilingData.TotalTimeMs != 0 {
		profilingText := fmt.Sprintf("Total time: %.2f sec",
			profilingData.TotalTimeMs/1e3,
		)

		fmt.Print(text.FgBlue.Sprint(profilingText))
		fmt.Println(text.Reset.EscapeSeq())
		fmt.Println()
		return
	}
}

type MultilineState int

const (
	MultilineNone MultilineState = iota
	MultilinePrompt
)

func repl(cfg ReplConfig) {
	// fmt.Println(text.FgBlue.Sprintf("Send a message, press /? for help"))
	cfg.fill()

	//l, err := readline.NewEx(&readline.Config{
	//	Prompt:          text.Colors{text.FgGreen, text.Bold}.Sprint("> "),
	//	AutoComplete:    completer,
	//	InterruptPrompt: "^C",
	//	EOFPrompt:       "^D",
	//	HistoryFile:     store.Get().HistoryFilePath(),
	//})
	l, err := readline.New(readline.Prompt{
		Prompt:         text.Colors{text.FgGreen, text.Bold}.Sprint("> "),
		AltPrompt:      text.Colors{text.FgGreen, text.Bold}.Sprint(". "),
		Placeholder:    "Send a message, press /? for help",
		AltPlaceholder: `Use """ to end multi-line input`,
	})
	if err != nil {
		panic(err)
	}
	// defer l.Close()

	var cancel func()
	cSignal := make(chan os.Signal, 1)
	signal.Notify(cSignal, os.Interrupt)
	go func() {
		for range cSignal {
			if cancel != nil {
				cancel()
				fmt.Print(text.Reset.EscapeSeq())
				os.Exit(0)
			}
		}
	}()

	var sb strings.Builder
	var multiline MultilineState
	for {
		line, err := l.Readline()

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

		if sb.Len() == 0 || multiline != MultilineNone {
			continue
		}

		line = sb.String()
		sb.Reset()

		// paser file
		var images, audios []string
		if cfg.ParseFile {
			line, images, audios = parseFiles(line)
		}

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
				cfg.Clear()
				fmt.Print("\033[H\033[2J")

			case "/load":
				if len(fileds) != 2 {
					fmt.Println(text.FgRed.Sprintf("Usage: /load <filename>"))
					fmt.Println()
				}
				cfg.Clear()
				err := cfg.LoadKVCache(fileds[1])
				if err != nil {
					fmt.Println(text.FgRed.Sprintf("Error: %s", err))
					fmt.Println()
				}

			case "/save":
				if len(fileds) != 2 {
					fmt.Println(text.FgRed.Sprintf("Usage: /save <filename>"))
					fmt.Println()
					continue
				}
				err := cfg.SaveKVCache(fileds[1])
				if err != nil {
					fmt.Println(text.FgRed.Sprintf("Error: %s", err))
					fmt.Println()
				}

			default:
				fmt.Println(text.FgRed.Sprintf("Unknown command: %s", fileds[0]))
				fmt.Println()
			}

			continue
		}

		// chat
		if cfg.Stream {
			// run async
			dataCh := make(chan string, 10)
			errCh := make(chan error, 1)

			fmt.Print(text.FgMagenta.EscapeSeq())
			ctx, cancelFunc := context.WithCancel(context.Background())
			cancel = cancelFunc
			go cfg.RunStream(ctx, line, images, audios, dataCh, errCh)

			// print stream
			firstToken := true
			for r := range dataCh {

				if firstToken {
					fmt.Print(text.FgYellow.EscapeSeq())
					firstToken = false
				}

				switch r {
				case "<think>":
					fmt.Print(text.FgHiBlack.EscapeSeq())
					fmt.Print(r)
				case "</think>":
					fmt.Print(r)
					fmt.Print(text.FgYellow.EscapeSeq())
				default:
					fmt.Print(r)
				}

			}
			fmt.Println(text.Reset.EscapeSeq())
			fmt.Println()

			if data, err := cfg.GetProfilingData(); err == nil {
				printProfiling(data)
			}

			// check error
			e, ok := <-errCh
			if ok {
				fmt.Println(text.FgRed.Sprintf("Error: %s\n", e))
				fmt.Println()
				return
			}
		} else {
			fmt.Print(text.FgMagenta.EscapeSeq())
			res, err := cfg.Run(line, images, audios)
			// append color to think
			res = strings.ReplaceAll(res, "<think>", text.FgBlack.EscapeSeq()+"<think>")
			res = strings.ReplaceAll(res, "</think>", "</think>"+text.FgYellow.EscapeSeq())
			fmt.Println(text.FgYellow.Sprint(res))
			fmt.Println()

			if data, err := cfg.GetProfilingData(); err == nil {
				printProfiling(data)
			}

			fmt.Print(text.Reset.EscapeSeq())
			if err != nil {
				fmt.Println(text.FgRed.Sprintf("Error: %s\n", err))
				fmt.Println()
				return
			}
		}

	}
}

// =============== file name parse ===============

var fileRegex = regexp.MustCompile(`(?:[a-zA-Z]:)?(?:\./|/|\\)[\S\\ ]+?\.(?i:jpg|jpeg|png|webp|mp3|wav)\b`)

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
			fmt.Println(text.FgRed.Sprintf("parse file error: [%s] %s", realFile, err))
			continue
		}
		switch realFile[len(realFile)-3:] {
		case "mp3", "wav":
			audios = append(audios, realFile)
			fmt.Println(text.FgBlue.Sprintf("add audio: %s", realFile))
		default:
			images = append(images, realFile)
			fmt.Println(text.FgBlue.Sprintf("add image: %s", realFile))
		}

		prompt = strings.ReplaceAll(prompt, "'"+realFile+"'", "")
		prompt = strings.ReplaceAll(prompt, "'"+file+"'", "")
		prompt = strings.ReplaceAll(prompt, file, "")
	}
	return strings.TrimSpace(prompt), images, audios
}

// =============== quant name parse ===============

// (fp32|fp16|f32|f16|q4_k_m|q4_1|i64|i32|i16|i8|iq4_nl|iq4_xs|iq3_s|iq3_xxs|iq2_xxs|iq2_s|iq2_xs|iq1_s|iq1_m|bf16).gguf
var quantRegix = regexp.MustCompile(`\b([fF][pP][0-9]+|[qQ][0-9]+(_[A-Z0-9]+)*|[fF][0-9]+|[iI][0-9]+|[iI][qQ][0-9]+(_[A-Z0-9]+)*|[bB][fF][0-9]+)`)

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

func chooseFiles(name string, files []string) (res types.ModelManifest, err error) {
	spin := render.NewSpinner("loading model size...")
	if len(files) == 0 {
		err = fmt.Errorf("repo is empty")
		return
	}

	res.Name = name

	// TODO: refactor
	// check gguf
	var mmprojs []string
	ggufGroups := make(map[string][]string)
	// qwen2.5-7b-instruct-q8_0-00003-of-00003.gguf original name is qwen2.5-7b-instruct-q8_0
	// *d-of-*d like this
	partRegix := regexp.MustCompile(`-\d+-of-\d+\.gguf$`)
	for _, file := range files {
		lower := strings.ToLower(file)
		if strings.HasSuffix(lower, ".gguf") {
			if strings.HasPrefix(lower, "mmproj") {
				mmprojs = append(mmprojs, file)
			} else {
				name := partRegix.ReplaceAllString(file, "")
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
			res.ModelFile = ggufs[0]
			spin.Start()
			fileSizes, err := getFileSizesConcurrent(name, ggufGroups[ggufs[0]])
			spin.Stop()
			if err != nil {
				fmt.Println(text.FgRed.Sprintf("get filesize error: [%s] %s", ggufs[0], err))
				return res, err
			}
			for _, size := range fileSizes {
				res.Size += size
			}
		} else {
			// interactive choose
			// Get file sizes for display
			spin.Start()
			// key is gguf file name, value is file size total containts part file
			fileSizes := make(map[string]int64)
			for _, gguf := range ggufs {
				sizes, err := getFileSizesConcurrent(name, ggufGroups[gguf])
				if err != nil {
					fmt.Println(text.FgRed.Sprintf("get filesize error: [%s] %s", gguf, err))
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
			// sort files by name
			files := ggufGroups[file]
			slices.Sort(files)

			res.Size += fileSizes[file]
			res.ModelFile = files[0]
			if len(files) > 1 {
				res.ExtraFiles = append(res.ExtraFiles, files[1:]...)
			}
		}
		res.Quant = strings.ToUpper(quantRegix.FindString(res.ModelFile))

		// detect mmproj
		switch len(mmprojs) {
		case 0:
		case 1:
			res.MMProjFile = mmprojs[0]
			spin.Start()
			size, err := store.Get().HFFileSize(context.TODO(), name, mmprojs[0])
			spin.Stop()
			if err != nil {
				fmt.Println(text.FgRed.Sprintf("get filesize error: [%s] %s", mmprojs[0], err))
				return res, err
			}
			res.Size += size
		default:
			// Get mmproj file sizes for display
			spin.Start()
			mmprojSizes, err := getFileSizesConcurrent(name, mmprojs)
			spin.Stop()
			if err != nil {
				fmt.Println(text.FgRed.Sprintf("get filesize error: %s", err))
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

			res.MMProjFile = file
			res.Size += mmprojSizes[file]
		}
	} else {
		// other format

		// detect main model file
		// add other files
		for _, file := range files {
			if res.ModelFile == "" {
				lower := strings.ToLower(file)
				if strings.HasSuffix(lower, "safetensors") || strings.HasSuffix(lower, "npz") {
					res.ModelFile = file
					continue
				}
			}
			res.ExtraFiles = append(res.ExtraFiles, file)
		}

		spin.Start()
		sizes, err := getFileSizesConcurrent(name, files)
		spin.Stop()
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("get filesize error: %s", err))
			return res, err
		}
		for _, size := range sizes {
			res.Size += size
		}

		// fallback to first file
		if res.ModelFile == "" {
			res.ModelFile = files[0]
			res.ExtraFiles = res.ExtraFiles[1:]
		}
	}

	return
}
