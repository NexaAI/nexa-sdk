package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/config"
)

func runFunc(cmd *cobra.Command, args []string) {
	model := args[0]

	client := openai.NewClient(
		option.WithBaseURL(fmt.Sprintf("http://%s/v1", config.Get().Host)),
	)

	// warm up
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()
	_, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: nil,
		Model:    model,
	})
	spin.Stop()
	if err != nil {
		fmt.Printf("Request Error: %s\n", err)
		return
	}

	// repl
	var history []openai.ChatCompletionMessageParamUnion
	repl(ReplConfig{
		Stream: *runStream,

		Clear: func() {
			history = nil
			client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
				Messages: nil,
				Model:    model,
			})
		},

		SaveKVCache: func(path string) error {
			return errors.New("not support")
			//return client.Post(context.TODO(), "/saveKVCache", nil, nil,
			//	option.WithJSONSet("model", model),
			//	option.WithJSONSet("name", path),
			//)
		},

		LoadKVCache: func(path string) error {
			return errors.New("not support")
			//return client.Post(context.TODO(), "/loadKVCache", nil, nil,
			//	option.WithJSONSet("model", model),
			//	option.WithJSONSet("name", path),
			//)
		},

		run: func(prompt string) (string, error) {
			history = append(history, openai.UserMessage(prompt))

			chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
				Messages: history,
				Model:    model,
			})
			if err != nil {
				return "", err
			}

			content := chatCompletion.Choices[0].Message.Content
			history = append(history, openai.AssistantMessage(content))
			return content, err
		},

		runStream: func(ctx context.Context, prompt string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			acc := openai.ChatCompletionAccumulator{}
			history = append(history, openai.UserMessage(prompt))

			stream := client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
				Messages: history,
				Model:    model,
			})

			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)
				if len(chunk.Choices) > 0 {
					dataCh <- chunk.Choices[0].Delta.Content
					acc.AddChunk(chunk)
				}
			}

			if stream.Err() != nil {
				errCh <- stream.Err()
			}

			history = append(history, openai.AssistantMessage(acc.Choices[0].Message.Content))
		},
	})
}

var runStream *bool
var runTools *string

func run() *cobra.Command {
	runCmd := &cobra.Command{}
	runCmd.Use = "run"

	runCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	runStream = runCmd.Flags().BoolP("stream", "s", false, "enable stream mode")
	runTools = runCmd.Flags().StringP("tools", "t", "", "function call tools")

	runCmd.Run = runFunc
	return runCmd
}
