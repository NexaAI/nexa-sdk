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
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

var (
	// disableStream *bool // reuse in run.go
	modelType string
	tool      []string
	prompt    []string
	query     string
	document  []string
)

func infer() *cobra.Command {
	inferCmd := &cobra.Command{
		Use:   "infer <model-name>",
		Short: "Infer with a model",
		Long:  "Run inference with a specified model. The model must be downloaded and cached locally.",
	}

	inferCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	inferCmd.Flags().SortFlags = false
	inferCmd.Flags().StringVarP(&modelType, "model-type", "m", "llm", "specify model type [llm/vlm/embedder/reranker]")
	inferCmd.Flags().BoolVarP(&disableStream, "disable-stream", "s", false, "[llm|vlm] disable stream mode")
	inferCmd.Flags().StringSliceVarP(&tool, "tool", "t", nil, "[llm|vlm] add tool to make function call")
	inferCmd.Flags().StringSliceVarP(&prompt, "prompt", "p", nil, "[embedder] pass prompt")
	inferCmd.Flags().StringVarP(&query, "query", "q", "", "[reranker] query")
	inferCmd.Flags().StringSliceVarP(&document, "document", "d", nil, "[reranker] documents")

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

		switch modelType {
		case types.ModelTypeLLM:
			if manifest.MMProjFile == "" {
				if len(tool) == 0 {
					inferLLM(modelfile, nil)
					return
				} else {
					panic("TODO")
				}
			} else {
				// compat vlm
				t := s.ModelfilePath(manifest.Name, manifest.MMProjFile)
				inferVLM(modelfile, &t)
			}
		case types.ModelTypeVLM:
			t := s.ModelfilePath(manifest.Name, manifest.MMProjFile)
			inferVLM(modelfile, &t)
		case types.ModelTypeEmbedder:
			inferEmbed(modelfile, nil)
		case types.ModelTypeReranker:
			inferRerank(modelfile, nil)
		default:
			panic("not support model type")
		}
	}
	return inferCmd
}

func inferLLM(model string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))

	spin.Start()
	p, err := nexa_sdk.NewLLM(model, tokenizer, 8192, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.ChatMessage

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

		GetProfilingData: func() (*nexa_sdk.ProfilingData, error) {
			return p.GetProfilingData()
		},

		Run: func(prompt string, _, _ []string) (string, error) {
			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})

			formatted, err := p.ApplyChatTemplate(history)
			if err != nil {
				if errors.Is(err, nexa_sdk.ErrChatTemplateNotFound) {
					// Chat template can be not found for some non-instruct-tuned models, we directly use the original prompt in those cases.
					formatted = prompt
					err = nil
				} else {
					return "", err
				}
			}

			res, err := p.Generate(formatted)
			if err != nil {
				return "", err
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: res})

			return res, nil
		},

		RunStream: func(ctx context.Context, prompt string, _, _ []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleUser, Content: prompt})
			formatted, e := p.ApplyChatTemplate(history)
			if e != nil {
				if errors.Is(e, nexa_sdk.ErrChatTemplateNotFound) {
					// Chat template can be not found for some non-instruct-tuned models, we directly use the original prompt in those cases.
					formatted = prompt
					e = nil
				} else {
					errCh <- e
					return
				}
			}

			var full strings.Builder
			// fmt.Printf(text.FgBlack.Sprint(formatted[:lastLen]))
			// fmt.Printf(text.FgCyan.Sprint(formatted[lastLen:]))
			dCh, eCh := p.GenerateStream(ctx, formatted)
			for r := range dCh {
				full.WriteString(r)
				dataCh <- r
			}
			for e := range eCh {
				errCh <- e
				return
			}

			history = append(history, nexa_sdk.ChatMessage{Role: nexa_sdk.LLMRoleAssistant, Content: full.String()})
		},
	})
}

func inferVLM(model string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))

	spin.Start()
	p, err := nexa_sdk.NewVLM(model, tokenizer, 8192, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	var history []nexa_sdk.ChatMessage
	var lastLen int

	repl(ReplConfig{
		Stream:    !disableStream,
		ParseFile: true,

		Clear: p.Reset,

		GetProfilingData: func() (*nexa_sdk.ProfilingData, error) {
			return p.GetProfilingData()
		},

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

			// fmt.Println(text.FgBlack.Sprint(prompt))

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

func inferEmbed(modelfile string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))

	spin.Start()
	p, err := nexa_sdk.NewEmbedder(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	}
	defer p.Destroy()

	if len(prompt) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 text prompt is accept"))
		fmt.Println()
		return
	}

	res, err := p.Embed(prompt)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	} else {
		nEmbed := len(res) / len(prompt)
		for i := range res {
			if i%nEmbed == 0 {
				fmt.Print(text.FgYellow.Sprintf("\n===> %d\n", i/nEmbed))
			}
			fmt.Print(text.FgYellow.Sprintf("%f ", res[i]))
		}
		fmt.Println()
	}
	fmt.Println()

	if data, err := p.GetProfilingData(); err == nil {
		printProfiling(data)
	}
}

func inferRerank(modelfile string, tokenizer *string) {
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))

	spin.Start()
	p, err := nexa_sdk.NewReranker(modelfile, tokenizer, nil)
	spin.Stop()
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		return
	}
	defer p.Destroy()

	if len(query) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 query is accept"))
		fmt.Println()
		return
	}
	if len(document) == 0 {
		fmt.Println(text.FgRed.Sprintf("at least 1 document is accept"))
		fmt.Println()
		return
	}

	res, err := p.Rerank(query, document)
	if err != nil {
		fmt.Println(text.FgRed.Sprintf("Error: %s", err))
		fmt.Println()
		return
	} else {
		fmt.Println()
		for i := range res {
			fmt.Println(text.FgYellow.Sprintf("%d => %f", i, res[i]))
		}
		fmt.Println()
	}

	if data, err := p.GetProfilingData(); err == nil {
		printProfiling(data)
	}
}
