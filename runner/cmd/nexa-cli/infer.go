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

var (
	//disableStream *bool
	tool         []string
	image, audio string
	prompt       []string
	query        string
	document     []string
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{}
	inferCmd.Use = "infer <model-name>"

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	inferCmd.Flags().BoolVarP(&disableStream, "disable-stream", "s", false, "disable stream mode in llm/vlm")
	inferCmd.Flags().StringSliceVarP(&tool, "tool", "t", nil, "add tool to make function call")
	inferCmd.Flags().StringVarP(&image, "image", "i", "", "pass image file to vlm")
	inferCmd.Flags().StringVarP(&audio, "audio", "a", "", "pass audio file to vlm")
	inferCmd.Flags().StringSliceVarP(&prompt, "prompt", "p", nil, "pass prompt to vlm/embedder")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "rerank only")
	inferCmd.Flags().StringSliceVarP(&document, "document", "d", nil, "rerank only")

	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.NewStore()
		model := args[0]
		manifest, err := s.GetManifest(model)
		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
			return
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
			if len(tool) == 0 {
				inferLLM(modelfile, tokenizer)
			} else {
			}
		case types.ModelTypeVLM:
			inferVLM(modelfile, tokenizer)
		case types.ModelTypeEmbedder:
			inferEmbed(modelfile, tokenizer)
		case types.ModelTypeReranker:
			inferRerank(modelfile, tokenizer)
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
		Stream: !disableStream,

		Clear: p.Reset,

		SaveKVCache: func(path string) error {
			return p.SaveKVCache(path)
		},

		LoadKVCache: func(path string) error {
			return p.LoadKVCache(path)
		},

		Run: func(prompt string) (string, error) {
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})

			formatted, err := p.ApplyChatTemplate(history)
			if err != nil {
				return "", err
			}

			res, err := p.Generate(prompt)
			if err != nil {
				return "", err
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res})

			formatted, err = p.ApplyChatTemplate(history)
			if err != nil {
				return "", err
			}
			lastLen = len(formatted)

			return res, nil
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

func inferVLM(model string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()
	p := nexa_sdk.NewVLM(model, tokenizer, 4096, nil)
	defer p.Destroy()

	spin.Stop()

	if len(prompt) != 1 {
		fmt.Println(text.FgRed.Sprintf("only 1 text prompt is accept"))
		return
	}

	formatted, err := p.ApplyChatTemplate([]nexa_sdk.ChatMessage{
		{Role: nexa_sdk.LLMRoleUser, Content: prompt[0]},
	})
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("apply chat template: %s", err))
		return
	}

	var file *string
	if image != "" {
		file = &image
	}
	if audio != "" {
		file = &audio
	}

	if !disableStream {
		fmt.Print(text.FgYellow.EscapeSeq())
		dCh, eCh := p.GenerateStream(context.TODO(), formatted, file)
		for r := range dCh {
			fmt.Print(r)
		}
		fmt.Print(text.Reset.EscapeSeq())
		fmt.Println()

		for e := range eCh {
			fmt.Println(text.FgRed.Sprintf("Error: %s", e))
		}
	} else {
		res, err := p.Generate(formatted, file)
		fmt.Println(text.FgYellow.Sprint(res))

		if err != nil {
			fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		}

	}
}

func inferEmbed(modelfile string, tokenizer *string) {
	p := nexa_sdk.NewEmbedder(modelfile, tokenizer, nil)
	defer p.Destroy()

	if len(prompt) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 text prompt is accept"))
		return
	}

	res, err := p.Embed(prompt)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	} else {
		nEmbed := len(res) / len(prompt)
		for i := range res {
			if i%nEmbed == 0 {
				fmt.Print(text.FgYellow.Sprintf("\n===> %d\n", i/nEmbed))
			}
			fmt.Print(text.FgYellow.Sprintf("%f ", res[i]))
		}
	}
}

func inferRerank(modelfile string, tokenizer *string) {
	p := nexa_sdk.NewReranker(modelfile, tokenizer, nil)
	defer p.Destroy()

	if len(query) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 query is accept"))
		return
	}
	if len(document) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 document is accept"))
		return
	}

	res, err := p.Rerank(query, document)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	} else {
		for i := range res {
			fmt.Println(text.FgYellow.Sprintf("%d => %f", i, res[i]))
		}
	}
}
