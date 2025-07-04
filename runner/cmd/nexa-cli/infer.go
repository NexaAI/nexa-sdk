package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

func infer() *cobra.Command {
	var disableStream bool
	var tool []string

	inferCmd := &cobra.Command{
		Use:   "infer <model-name>",
		Short: "Infer with a model",
		Long:  "Run inference with a specified model. The model must be downloaded and cached locally.",
	}

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	inferCmd.Flags().BoolVarP(&disableStream, "disable-stream", "s", false, "disable stream mode in llm/vlm")
	inferCmd.Flags().StringSliceVarP(&tool, "tool", "t", nil, "add tool to make function call")

	inferCmd.Run = func(cmd *cobra.Command, args []string) {
		model := normalizeModelName(args[0])

		s := store.Get()

		manifest, err := s.GetManifest(model)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(text.FgBlue.Sprintf("model not found, start download"))

			pull().Run(cmd, args)
			// check agin
			manifest, err = s.GetManifest(model)
		}

		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
			return
		}

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile)

		if manifest.MMProjFile != "" {
			t := s.ModelfilePath(manifest.Name, manifest.MMProjFile)
			inferVLM(modelfile, &t)
		} else {
			inferLLM(modelfile, nil)
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
		Stream:    !disableStream,
		ParseFile: false,

		Clear: p.Reset,

		SaveKVCache: func(path string) error {
			return p.SaveKVCache(path)
		},

		LoadKVCache: func(path string) error {
			return p.LoadKVCache(path)
		},

		Run: func(prompt string, _, _ []string) (string, error) {
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
			lastLen = len(formatted) + len(res)

			return res, nil
		},

		RunStream: func(ctx context.Context, prompt string, _, _ []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}

			var full strings.Builder
			//fmt.Printf(text.FgBlack.Sprint(formatted[:lastLen]))
			//fmt.Printf(text.FgCyan.Sprint(formatted[lastLen:]))
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
			lastLen = len(formatted) + len(full.String())
		},
	})
}

func inferVLM(model string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()

	p := nexa_sdk.NewVLM(model, tokenizer, 4096, nil)
	defer p.Destroy()

	spin.Stop()

	var history []nexa_sdk.ChatMessage
	var lastLen int

	repl(ReplConfig{
		Stream:    !disableStream,
		ParseFile: true,

		Clear: p.Reset,

		Run: func(prompt string, images, audios []string) (string, error) {
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, err := p.ApplyChatTemplate(history)
			if err != nil {
				return "", err
			}

			res, err := p.Generate(prompt, images, audios)
			if err != nil {
				return "", err
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res})
			lastLen = len(formatted) + len(res)

			return res, nil
		},

		RunStream: func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			//fmt.Println(text.FgBlack.Sprint(prompt))

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				errCh <- e
				return
			}

			var full strings.Builder
			dCh, eCh := p.GenerateStream(ctx, formatted[lastLen:], images, audios)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})
			lastLen = len(formatted) + len(full.String())
		},
	})
}

func embed() *cobra.Command {
	var prompt []string

	embedCmd := &cobra.Command{
		Use:   "embed <model-name>",
		Short: "infer a embed model",
	}

	embedCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	embedCmd.Flags().StringSliceVarP(&prompt, "prompt", "p", nil, "pass prompt to embedder")

	embedCmd.Run = func(cmd *cobra.Command, args []string) {
		model := normalizeModelName(args[0])

		s := store.Get()

		manifest, err := s.GetManifest(model)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(text.FgBlue.Sprintf("model not found, start download"))

			pull().Run(cmd, args)
			// check agin
			manifest, err = s.GetManifest(model)
		}

		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
			return
		}

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile)
		p := nexa_sdk.NewEmbedder(modelfile, nil, nil)
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
	return embedCmd
}

func rerank() *cobra.Command {
	var query string
	var document []string

	rerankCmd := &cobra.Command{
		Use:   "rerank <model-name>",
		Short: "infer a rerank model",
	}

	rerankCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	rerankCmd.Flags().StringVarP(&query, "query", "q", "", "rerank only")
	rerankCmd.Flags().StringSliceVarP(&document, "document", "d", nil, "rerank only")

	rerankCmd.Run = func(cmd *cobra.Command, args []string) {
		model := normalizeModelName(args[0])

		s := store.Get()

		manifest, err := s.GetManifest(model)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println(text.FgBlue.Sprintf("model not found, start download"))

			pull().Run(cmd, args)
			// check agin
			manifest, err = s.GetManifest(model)
		}

		if err != nil {
			fmt.Println(text.FgRed.Sprintf("parse manifest error: %s", err))
			return
		}

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile)
		p := nexa_sdk.NewReranker(modelfile, nil, nil)
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
	return rerankCmd
}
