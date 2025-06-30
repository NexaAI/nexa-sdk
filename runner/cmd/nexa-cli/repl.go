package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/jedib0t/go-pretty/v6/text"
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
	Stream bool

	Clear       func()
	SaveKVCache func(path string) error
	LoadKVCache func(path string) error

	Run       func(prompt string) (string, error)
	RunStream func(ctx context.Context, prompt string, dataCh chan<- string, errCh chan<- error)
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
		cfg.Run = func(string) (string, error) { return "", notSupport }
	}
	if cfg.RunStream == nil {
		cfg.RunStream = func(ctx context.Context, prompt string, dataCh chan<- string, errCh chan<- error) {
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

		if strings.HasPrefix(line, "/") {
			fileds := strings.Fields(strings.TrimSpace(line))

			switch fileds[0] {
			case "/?", "/h", "/help":
				fmt.Println("Commands:")
				fmt.Println(completer.Tree("    ")) // TODO: add description

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
			go cfg.RunStream(context.TODO(), line, dataCh, errCh)

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

			res, err := cfg.Run(line)
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
