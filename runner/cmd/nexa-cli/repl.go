package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
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

type ReplConfig struct {
	ParseFile bool

	Reset       func() error
	SaveKVCache func(path string) error
	LoadKVCache func(path string) error

	Record func() (*string, error)
	Run    func(prompt string, images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error)
}

func (cfg *ReplConfig) fill() {
	notSupport := fmt.Errorf("notSupport")

	if cfg.Reset == nil {
		cfg.Reset = func() error { return nil }
	}
	if cfg.SaveKVCache == nil {
		cfg.SaveKVCache = func(string) error { return nexa_sdk.ErrCommonNotSupport }
	}
	if cfg.LoadKVCache == nil {
		cfg.LoadKVCache = func(string) error { return nexa_sdk.ErrCommonNotSupport }
	}
	if cfg.Record == nil {
		cfg.Record = func() (*string, error) { return nil, notSupport }
	}
	if cfg.Run == nil {
		cfg.Run = func(string, []string, []string, func(string) bool) (string, nexa_sdk.ProfileData, error) {
			return "", nexa_sdk.ProfileData{}, notSupport
		}
	}
}

// ========= repl tool ========

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
		if multiline == MultilineNone && len(recordAudios) > 0 {
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
				continue

			case "/exit":
				return

			case "/clear":
				cfg.Reset()
				recordAudios = nil
				fmt.Print("\033[H\033[2J")
				continue

			case "/load":
				if len(fileds) != 2 {
					fmt.Println(render.GetTheme().Error.Sprintf("Usage: /load <filename>"))
					fmt.Println()
					continue
				}
				cfg.Reset()
				err := cfg.LoadKVCache(fileds[1])
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
				err := cfg.SaveKVCache(fileds[1])
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
				outputFile, err := cfg.Record()
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

		// run async
		ctx, cancelFunc := context.WithCancel(context.Background())
		cancel = cancelFunc

		firstToken := true
		spin := render.NewSpinner("encoding...")
		spin.Start()

		audios = append(audios, recordAudios...)
		recordAudios = nil // clear after use

		state := STATE_ASSISTANT
		_, profileData, err := cfg.Run(line, images, audios, func(token string) bool {
			if firstToken {
				spin.Stop()
				firstToken = false
			}

			fsmEvent(&state, token)

			return ctx.Err() == nil
		})
		slog.Debug("profileData", "profileData", profileData)

		// reset spin when no token received
		if firstToken {
			spin.Stop()
		}

		// reset color
		render.GetTheme().Reset()
		fmt.Println()
		fmt.Println()
		printProfile(profileData)

		switch {
		case err == nil:
		case errors.Is(err, nexa_sdk.ErrLlmTokenizationContextLength):
			fmt.Println(render.GetTheme().Info.Sprintf("Context length exceeded, please start a new conversation"))
			fmt.Println()
			return
		case errors.Is(err, ErrNoAudio):
			fmt.Println(render.GetTheme().Error.Sprintf("No audio file provided, please provide an audio file or use /mic command"))
			fmt.Println()
		default:
			fmt.Println(render.GetTheme().Error.Sprintf("Error: %s\n", err))
			fmt.Println()
			return
		}
	}
}

// print profile data

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

	if text == "" {
		return
	}

	fmt.Println(render.GetTheme().Profile.Sprint(text))
	fmt.Println()
}

// output parse

const (
	STATE_ASSISTANT = iota // init state
	STATE_THINK
	STATE_NORMAL

	STATE_START
	STATE_CHANNEL
	STATE_ANALYSIS
	STATE_FINAL
	STATE_END
)

var fsm = map[[2]any][2]any{
	// normal
	{STATE_ASSISTANT, "<think>"}: {STATE_THINK, thinkStart(false)},
	{STATE_THINK, "</think>"}:    {STATE_NORMAL, thinkEnd(false)},

	// gpt-oss
	{STATE_ASSISTANT, "<|channel|>"}: {STATE_CHANNEL, nil},
	{STATE_CHANNEL, "analysis"}:      {STATE_ANALYSIS, nil},
	{STATE_CHANNEL, "final"}:         {STATE_FINAL, nil},
	{STATE_ANALYSIS, "<|message|>"}:  {STATE_THINK, thinkStart(true)},
	{STATE_FINAL, "<|message|>"}:     {STATE_NORMAL, nil},
	{STATE_THINK, "<|end|>"}:         {STATE_END, thinkEnd(true)},
	{STATE_NORMAL, "<|end|>"}:        {STATE_END, nil},
	{STATE_END, "<|start|>"}:         {STATE_START, nil},
	{STATE_START, "assistant"}:       {STATE_ASSISTANT, nil},
}

var (
	thinkSpin  = render.NewSpinner("thinking...")
	thinkStart = func(extraLine bool) func() {
		return func() {
			if hideThink {
				thinkSpin.Start()
			} else {
				render.GetTheme().Set(render.GetTheme().ThinkOutput)
				if extraLine {
					fmt.Print("<think>\n")
				} else {
					fmt.Print("<think>")
				}
			}
		}
	}
	thinkEnd = func(extraLine bool) func() {
		return func() {
			if hideThink {
				thinkSpin.Stop()
			} else {
				if extraLine {
					fmt.Print("\n</think>\n\n")
				} else {
					fmt.Print("</think>")
				}
				render.GetTheme().Set(render.GetTheme().ModelOutput)
			}
		}
	}
)

func fsmEvent(state *int, token string) {
	next, ok := fsm[[2]any{*state, token}]
	if ok {
		*state = next[0].(int)
		if next[1] != nil {
			next[1].(func())()
		}
		return
	}

	if !(hideThink && *state == STATE_THINK) {
		fmt.Print(token)
	}
}

// file name parse

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
