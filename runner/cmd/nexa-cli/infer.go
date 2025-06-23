package main

import (
	"context"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func inferFunc(cmd *cobra.Command, args []string) {
	s := store.NewStore()

	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()
	nexa_sdk.Init()
	p := nexa_sdk.NewLLM(s.ModelfilePath(args[0]), nil, 4096, nil)
	time.Sleep(time.Second) // TODO: remove test code
	spin.Stop()

	var history []nexa_sdk.ChatMessage
	var lastLen int

	repl(ReplConfig{
		Stream:      true,
		Reset:       p.Reset,
		SaveKVCache: nil,
		LoadKVCache: nil,
		runStream: func(ctx context.Context, prompt string, dataCh chan<- string, errCh chan<- error) {
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

	p.Destroy()
	nexa_sdk.DeInit()
}

func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer"

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	inferCmd.Run = inferFunc
	return inferCmd
}
