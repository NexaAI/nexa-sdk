package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/chzyer/readline"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/NexaAI/nexa-sdk/internal/types"
)

var completer = readline.NewPrefixCompleter(
	readline.PcItem("/?"),
	readline.PcItem("/h"),
	readline.PcItem("/help"),

	readline.PcItem("/exit"),

	readline.PcItem("/clear"),
	readline.PcItem("/load", readline.PcItemDynamic(listFiles("./"))),
	readline.PcItem("/save", readline.PcItemDynamic(listFiles("./"))),
)

var help = [][2]string{
	{"/?, /h, /help", "Show this help message"},
	{"/exit", "Exit the REPL"},
	{"/clear", "Clear the screen and conversation history"},
	{"/load <filename>", "Load conversation history from a file"},
	{"/save <filename>", "Save conversation history to a file"},
}

// TODO: support sub dir
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := os.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

// LLM, VLM
type ReplConfig struct {
	Stream    bool
	ParseFile bool

	Clear       func()
	SaveKVCache func(path string) error
	LoadKVCache func(path string) error

	Run       func(prompt string, images, audios []string) (string, error)
	RunStream func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error)
}

func (cfg *ReplConfig) fill() {
	var notSupport = fmt.Errorf("notSupport")

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

func repl(cfg ReplConfig) {
	fmt.Println(text.FgBlue.Sprintf("Send a message, press /? for help"))
	cfg.fill()

	l, err := readline.NewEx(&readline.Config{
		Prompt:          text.Colors{text.FgGreen, text.Bold}.Sprint("> "),
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "^D",
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()

	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				return
			} else {
				continue
			}
		} else if err == io.EOF {
			return
		}

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

			case "/exit":
				return

			case "/clear":
				cfg.Clear()
				fmt.Print("\033[H\033[2J")

			case "/load":
				if len(fileds) != 2 {
					fmt.Println(text.FgRed.Sprintf("Usage: /load <filename>"))
				}
				cfg.Clear()
				err := cfg.LoadKVCache(fileds[1])
				if err != nil {
					fmt.Println(text.FgRed.Sprintf("Error: %s", err))
				}

			case "/save":
				if len(fileds) != 2 {
					fmt.Println(text.FgRed.Sprintf("Usage: /save <filename>"))
					continue
				}
				err := cfg.SaveKVCache(fileds[1])
				if err != nil {
					fmt.Println(text.FgRed.Sprintf("Error: %s", err))
				}

			default:
				fmt.Println(text.FgRed.Sprintf("Unknown command: %s", fileds[0]))
			}

			continue
		}

		// chat
		if cfg.Stream {
			var count int
			var tokenStart time.Time
			var firstToken bool

			// track RunStream start time for TTFT calculation
			runStreamStart := time.Now()

			// run async
			dataCh := make(chan string, 10)
			errCh := make(chan error, 1)
			go cfg.RunStream(context.TODO(), line, images, audios, dataCh, errCh)

			// print stream
			fmt.Print(text.FgYellow.EscapeSeq())
			for r := range dataCh {
				if !firstToken {
					tokenStart = time.Now()
					firstToken = true
				}
				fmt.Print(r)
				count++
			}
			fmt.Print(text.Reset.EscapeSeq())
			fmt.Println()
      
			// print metrics
			if firstToken {
				ttft := tokenStart.Sub(runStreamStart).Seconds()
				tokenDuration := time.Since(tokenStart).Seconds()
				tokensPerSecond := float64(count) / tokenDuration

				fmt.Println(text.FgBlue.Sprintf(
					"TTFT: %f s, Generated %d tokens at %f token/s",
					ttft, count, tokensPerSecond,
				))
			} else {
				fmt.Println(text.FgBlue.Sprintf("(no tokens generated)"))
			}

			// check error
			e, ok := <-errCh
			if ok {
				fmt.Println(text.FgRed.Sprintf("Error: %s\n", e))
			}
		} else {
			start := time.Now()

			res, err := cfg.Run(line, images, audios)
			fmt.Println(text.FgYellow.Sprint(res))

			// print duration
			duration := time.Since(start).Seconds()
			fmt.Println(text.FgBlue.Sprintf(
				"Generate in %f s\n",
				duration,
			))

			if err != nil {
				fmt.Println(text.FgRed.Sprintf("Error: %s\n", err))
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

// (f32|f16|q4_k_m|q4_1).gguf
var quantRegix = regexp.MustCompile(`\b([qQ][0-9]+(_[A-Z0-9]+)*|[fF][0-9]+)`)

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

func chooseFiles(name string, files []string) (res types.ModelManifest, err error) {
	if len(files) == 0 {
		err = fmt.Errorf("repo is empty")
		return
	}

	res.Name = name

	// choose model type
	var modelTypeString string
	if err = huh.NewSelect[string]().
		Title("Choose model type").
		Options(
			huh.NewOption(types.ModelTypeLLM, types.ModelTypeLLM),
			huh.NewOption(types.ModelTypeVLM, types.ModelTypeVLM),
			huh.NewOption(types.ModelTypeEmbedder, types.ModelTypeEmbedder),
			huh.NewOption(types.ModelTypeReranker, types.ModelTypeReranker),
		).
		Value(&modelTypeString).
		Run(); err != nil {
		return
	}
	res.ModelType = types.ModelType(modelTypeString)

	// check gguf
	var ggufs, mmprojs []string
	for _, file := range files {
		lower := strings.ToLower(file)
		if strings.HasSuffix(lower, ".gguf") {
			if strings.HasPrefix(lower, "mmproj") {
				mmprojs = append(mmprojs, file)
			} else {
				ggufs = append(ggufs, file)
			}
		}
	}

	// choose model file
	if len(ggufs) > 0 || len(mmprojs) > 0 {
		// detect gguf
		switch len(ggufs) {
		case 0:
			err = fmt.Errorf("can no detect model file in repo")
			return
		case 1:
			res.ModelFile = ggufs[0]
		default:
			// interactive choose

			var file, quant string
			// select default gguf
			for _, gguf := range ggufs {
				ggufQuant := quantRegix.FindString(gguf)
				if quantGreaterThan(ggufQuant, quant, []string{"Q4_K_M", "Q4_0", "Q8_0"}) {
					quant = ggufQuant
					file = gguf
				}
			}

			options := make([]huh.Option[string], 0, len(ggufs)+1)
			if file != "" {
				options = append(options, huh.NewOption(fmt.Sprintf("default (%s)", quant), file))
			}
			for i := range ggufs {
				quant := quantRegix.FindString(ggufs[i])
				if quant != "" && file != ggufs[i] {
					options = append(options, huh.NewOption(quant, ggufs[i]))
				}
			}

			if len(options) == 0 {
				err = fmt.Errorf("no valid gguf found")
				return
			}

			if err = huh.NewSelect[string]().
				Title("Choose a quant version to download").
				Options(options...).
				Value(&file).
				Run(); err != nil {
				return
			}

			res.ModelFile = file
		}

		if res.ModelType == types.ModelTypeVLM {
			// detect mmproj
			switch len(mmprojs) {
			case 0:
			case 1:
				res.TokenizerFile = mmprojs[0]
			default:
				// match biggest
				var file, quant string

				for _, mmproj := range mmprojs {
					mmprojQuant := quantRegix.FindString(mmproj)
					if quantGreaterThan(mmprojQuant, quant, nil) {
						quant = mmprojQuant
						file = mmproj
					}
				}

				res.TokenizerFile = file
			}
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
		// fallback to first file
		if res.ModelFile == "" {
			res.ModelFile = files[0]
			res.ExtraFiles = res.ExtraFiles[1:]
		}
	}

	return
}
