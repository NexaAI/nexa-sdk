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
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer <model-name>"

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	image := inferCmd.Flags().StringP("image file", "i", "", "pass image file to vlm")
	audio := inferCmd.Flags().StringP("audio file", "a", "", "pass audio file to vlm")
	prompts := inferCmd.Flags().StringSliceP("prompt", "p", nil, "embed only")
	query := inferCmd.Flags().StringP("query", "q", "", "rerank only")
	documents := inferCmd.Flags().StringSliceP("document", "d", nil, "rerank only")
	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		model := args[0]
		manifest, err := s.GetManifest(model)
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
		}

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile)
		var tokenizer *string
		if manifest.TokenizerFile != "" {
			t := s.ModelfilePath(manifest.Name, manifest.TokenizerFile)
			tokenizer = &t
		}

		switch manifest.ModelType {
		case types.ModelTypeLLM:
			inferLLM(modelfile, tokenizer)
		case types.ModelTypeVLM:
			inferVLM(modelfile, tokenizer, image, audio)
		case types.ModelTypeEmbed:
			inferEmbed(modelfile, tokenizer, *prompts)
		case types.ModelTypeRerank:
			inferRerank(modelfile, tokenizer, *query, *documents)
		default:
			panic("not support model type")
		}
	}
	return inferCmd
}

func inferLLM(model string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()

	p := nexa_sdk.NewLLM(model, tokenizer, 4096, nil)
	defer p.Destroy()

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

func inferVLM(model string, tokenizer *string, image *string, audio *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()
	p := nexa_sdk.NewVLM(model, tokenizer, 4096, nil)
	defer p.Destroy()

	spin.Stop()

	var history []nexa_sdk.ChatMessage
	var lastFile *string
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
			dCh, eCh := p.GenerateStream(ctx, formatted[lastLen:], lastFile)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}
			lastFile = nil

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})

			formatted, e = p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}
			lastLen = len(formatted)
		},

		Image: func(file string) error {
			lastFile = &file
			return nil
		},

		Audio: func(file string) error {
			lastFile = &file
			return nil
		},
	})
}

func inferEmbed(modelfile string, tokenizer *string, prompts []string) {
	p := nexa_sdk.NewEmbedder(modelfile, tokenizer, nil)
	defer p.Destroy()

	res, err := p.Embed(prompts)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	} else {
		nEmbed := len(res) / len(prompts)
		for i := range res {
			if i%nEmbed == 0 {
				fmt.Print(text.FgYellow.Sprintf("\n===> %d\n", i))
			}
			fmt.Print(text.FgYellow.Sprintf("%f ", res[i]))
		}
	}
}

func inferRerank(modelfile string, tokenizer *string, query string, documents []string) {
	p := nexa_sdk.NewReranker(modelfile, tokenizer, nil)
	defer p.Destroy()

	res, err := p.Rerank(query, documents)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	} else {
		for i := range res {
			fmt.Println(text.FgYellow.Sprintf("%f", res[i]))
		}
	}
}
