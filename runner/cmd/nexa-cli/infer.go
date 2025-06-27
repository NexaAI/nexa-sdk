package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer"

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	multiModal := inferCmd.Flags().BoolP("multi-model", "m", false, "enable multi model mode")
	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		file, err := s.ModelfilePath(args[0])

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
		}

		if !*multiModal {
			inferLLM(file)
		} else {
			inferVLM(file)
		}
	}
	return inferCmd
}

func inferLLM(file string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()

	p := nexa_sdk.NewLLM(file, nil, 4096, nil)
	defer p.Destroy()

	time.Sleep(time.Second) // TODO: remove test code
	spin.Stop()

	var history []nexa_sdk.ChatMessage
	var lastLen int

	repl(ReplConfig{
		Stream: true,

		Clear: p.Reset,

		SaveKVCache: func(path string) error {
			return p.SaveKVCache(path)
		},

		LoadKVCache: func(path string) error {
			return p.LoadKVCache(path)
		},

		RunStream: func(ctx context.Context, prompt string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}

			var full strings.Builder
			dCh, eCh := p.GenerateStream(ctx, formatted[lastLen:])
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})

			formatted, e = p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}
			lastLen = len(formatted)
		},
	})
}

func inferVLM(file string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()
	p := nexa_sdk.NewVLM(file, nil, 4096, nil)
	defer p.Destroy()

	time.Sleep(time.Second) // TODO: remove test code
	spin.Stop()

	var history []nexa_sdk.ChatMessage
	var lastImage *string
	var lastLen int

	repl(ReplConfig{
		Stream: true,

		Clear: p.Reset,

		RunStream: func(ctx context.Context, prompt string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}

			var full strings.Builder
			dCh, eCh := p.GenerateStream(ctx, formatted[lastLen:], lastImage)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}
			lastImage = nil

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})

			formatted, e = p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}
			lastLen = len(formatted)
		},

		Image: func(file string) error {
			lastImage = &file
			return nil
		},
	})
}
