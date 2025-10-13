package common

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

var ErrNoAudio = errors.New("no audio file provided")

type Processor struct {
	ParseFile bool
	HideThink bool

	GetPrompt func() (string, error)
	Run       func(prompt string, images, audios []string, onToken func(string) bool) (string, nexa_sdk.ProfileData, error)

	fsm      map[[2]any][2]any
	fsmState int
}

func (p *Processor) Process() {
	var stopGen bool
	go func() {
		cSignal := make(chan os.Signal, 1)
		signal.Notify(cSignal, os.Interrupt)
		for range cSignal {
			slog.Warn("interrupt signal received")
			stopGen = true
		}
	}()

	for {
		line, err := p.GetPrompt()
		slog.Debug("GetPrompt", "line", line, "err", err)

		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			fmt.Println(render.GetTheme().Error.Sprintf("Error: %s\n", err))
			return
		}

		var prompt string
		var images, audios []string
		if p.ParseFile {
			prompt, images, audios = p.parseFiles(line)
		} else {
			prompt = line
		}

		// run async
		firstToken := true
		spin := render.NewSpinner("encoding...") // merge into fsm
		spin.Start()

		p.fsmInit()
		stopGen = false
		_, profileData, err := p.Run(prompt, images, audios, func(token string) bool {
			if firstToken {
				spin.Stop()
				firstToken = false
			}

			p.fsmEvent(token)
			return !stopGen
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
		p.printProfile(profileData)

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

// file name parse

var fileRegex = regexp.MustCompile(`(?:[a-zA-Z]:)?(?:\./|/|\\)[\S\\ ]+?\.(?i:jpg|jpeg|png|webp|mp3|wav)\b`)

func (p *Processor) parseFiles(prompt string) (string, []string, []string) {
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

var thinkSpin = render.NewSpinner("thinking...")

func (p *Processor) fsmInit() {
	thinkStart := func(extraLine bool) func() {
		return func() {
			if p.HideThink {
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
	thinkEnd := func(extraLine bool) func() {
		return func() {
			if p.HideThink {
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
	p.fsm = map[[2]any][2]any{
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
	p.fsmState = STATE_ASSISTANT
}

func (p *Processor) fsmEvent(token string) {
	next, ok := p.fsm[[2]any{p.fsmState, token}]
	if ok {
		p.fsmState = next[0].(int)
		if next[1] != nil {
			next[1].(func())()
		}
		return
	}

	if !(p.HideThink && p.fsmState == STATE_THINK) {
		fmt.Print(token)
	}
}

// print profile data

func (p *Processor) printProfile(pd nexa_sdk.ProfileData) {
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
