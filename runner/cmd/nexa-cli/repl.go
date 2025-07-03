package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
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

		if len(images) == 0 && len(audio) == 0 && strings.HasPrefix(line, "/") {

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
			start := time.Now()
			var count int

			// run async
			dataCh := make(chan string, 10)
			errCh := make(chan error, 1)
			go cfg.RunStream(context.TODO(), line, images, audios, dataCh, errCh)

			// print stream
			fmt.Print(text.FgYellow.EscapeSeq())
			for r := range dataCh {
				fmt.Print(r)
				count++
			}
			fmt.Print(text.Reset.EscapeSeq())
			fmt.Println()

			// print duration
			duration := time.Since(start).Seconds()
			fmt.Println(text.FgBlue.Sprintf(
				"Generate %d token in %f s, speed is %f token/s",
				count, duration, float64(count)/duration,
			))

			// check error
			e, ok := <-errCh
			if ok {
				fmt.Println(text.FgRed.Sprintf("Error: %s", e))
			}
		} else {
			start := time.Now()

			res, err := cfg.Run(line, images, audios)
			fmt.Println(text.FgYellow.Sprint(res))

			// print duration
			duration := time.Since(start).Seconds()
			fmt.Println(text.FgBlue.Sprintf(
				"Generate in %f s",
				duration,
			))

			if err != nil {
				fmt.Println(text.FgRed.Sprintf("Error: %s", err))
			}
		}
	}
}

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

var quantRegix = regexp.MustCompile(``)

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
			var useDefault bool

			var file string
			// select default gguf

			if err = huh.NewSelect[bool]().
				Title("Choose a version to download").
				Options(
					huh.NewOption(fmt.Sprintf("default (%s)", file), true),
					huh.NewOption("choose a quant version", false),
				).
				Value(&useDefault).
				Run(); err != nil {
				return
			}

			if !useDefault {
				options := make([]huh.Option[string], len(ggufs))
				for i := range ggufs {
					options[i] = huh.NewOption(ggufs[i], ggufs[i])
				}

				if err = huh.NewSelect[string]().
					Title("Choose a quant version to download").
					Options(options...).
					Value(&file).
					Run(); err != nil {
					return
				}

			}

		}

		// detect mmproj
		switch len(mmprojs) {
		case 0:
		case 1:
			res.TokenizerFile = mmprojs[0]
		default:
			// match biggest
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
