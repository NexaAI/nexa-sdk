package common

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ollama/ollama/readline"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
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

type Repl struct {
	Reset       func() error
	SaveKVCache func(path string) error
	LoadKVCache func(path string) error

	Record func() (*string, error)

	init bool
	rl   *readline.Instance
}

// ========= repl tool ========

type MultilineState int

const (
	MultilineNone MultilineState = iota
	MultilinePrompt
)

func (r *Repl) GetPrompt() (string, error) {
	if !r.init {
		// fill default functions
		notSupport := fmt.Errorf("NotSupport")
		if r.Reset == nil {
			r.Reset = func() error { return nil }
		}
		if r.SaveKVCache == nil {
			r.SaveKVCache = func(path string) error { return notSupport }
		}
		if r.LoadKVCache == nil {
			r.LoadKVCache = func(path string) error { return notSupport }
		}
		if r.Record == nil {
			r.Record = func() (*string, error) { return nil, nexa_sdk.ErrCommonNotSupport }
		}

		// init readline
		rl, err := readline.New(readline.Prompt{
			Prompt:         render.GetTheme().Prompt.Sprint("> "),
			AltPrompt:      render.GetTheme().Prompt.Sprint(". "),
			Placeholder:    "Send a message, press /? for help",
			AltPlaceholder: `Use """ to end multi-line input`,
		})
		if err != nil {
			panic(err)
		}
		r.rl = rl
		// TODO: graceful shutdown
		fmt.Print(readline.StartBracketedPaste)

		r.init = true
	}

	var sb strings.Builder
	var multiline MultilineState
	var recordAudios []string

	for {
		// print stashed content
		if multiline == MultilineNone && len(recordAudios) > 0 {
			fmt.Println(render.GetTheme().Info.Sprintf("Current stash audios: %s", strings.Join(recordAudios, ", ")))
		}

		line, err := r.rl.Readline()

		// check err or exit
		switch {
		case errors.Is(err, io.EOF):
			fmt.Println()
			return "", io.EOF
		case errors.Is(err, readline.ErrInterrupt):
			if line == "" {
				fmt.Println("\nUse Ctrl + d or /exit to exit.")
				fmt.Println()
			}
			r.rl.Prompt.UseAlt = false
			sb.Reset()
			continue
		case err != nil:
			return "", err
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
			r.rl.Prompt.UseAlt = false

		case strings.HasPrefix(line, `"""`):
			line := strings.TrimPrefix(line, `"""`)
			line, ok := strings.CutSuffix(line, `"""`)
			sb.WriteString(line)
			if !ok {
				// no multiline terminating string; need more input
				fmt.Fprintln(&sb)
				multiline = MultilinePrompt
				r.rl.Prompt.UseAlt = true
			}

		case r.rl.Pasting:
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

		// check if it's a command
		var fileds []string
		if !strings.HasPrefix(line, "/") {
			line += strings.Join(recordAudios, " ")
			recordAudios = nil // clear stashed audios after use
			return line, nil
		}
		fileds = strings.Fields(strings.TrimSpace(line))
		if strings.Contains(fileds[0][1:], "/") || strings.Contains(fileds[0], ".") {
			line += strings.Join(recordAudios, " ")
			recordAudios = nil // clear stashed audios after use
			return line, nil
		}

		switch fileds[0] {
		case "/?", "/h", "/help":
			fmt.Println("Commands:")
			for _, h := range help {
				fmt.Printf("  %-25s %s\n", h[0], h[1])
			}
			fmt.Println()
			continue

		case "/exit":
			return "", io.EOF

		case "/clear":
			r.Reset()
			recordAudios = nil
			fmt.Print("\033[H\033[2J")
			continue

		case "/load":
			if len(fileds) != 2 {
				fmt.Println(render.GetTheme().Error.Sprintf("Usage: /load <filename>"))
				fmt.Println()
				continue
			}
			r.Reset()
			err := r.LoadKVCache(fileds[1])
			if err != nil {
				if errors.Is(err, nexa_sdk.ErrCommonNotSupport) {
					fmt.Println(render.GetTheme().Warning.Sprintf("Load conversation history is not supported for this model yet"))
					fmt.Println()
				} else {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					fmt.Println()
				}
			}
			continue

		case "/save":
			if len(fileds) != 2 {
				fmt.Println(render.GetTheme().Error.Sprintf("Usage: /save <filename>"))
				fmt.Println()
				continue
			}
			err := r.SaveKVCache(fileds[1])
			if err != nil {
				if errors.Is(err, nexa_sdk.ErrCommonNotSupport) {
					fmt.Println(render.GetTheme().Warning.Sprintf("Save conversation history is not supported for this model yet"))
					fmt.Println()
				} else {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					fmt.Println()
				}
			}
			continue

		case "/mic":
			outputFile, err := r.Record()
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
				fmt.Println()
			}
			if outputFile != nil {
				recordAudios = append(recordAudios, *outputFile)
			}
			continue

		default:
			fmt.Println(render.GetTheme().Error.Sprintf("Unknown command: %s", fileds[0]))
			fmt.Println()
			continue
		}
	}
}

func (r *Repl) Close() {
	if r.init {
		fmt.Printf(readline.EndBracketedPaste)
	}
}
