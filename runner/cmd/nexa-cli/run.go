package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/briandowns/spinner"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/config"
)

var disableStream bool

func run() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <model-name>",
		Short: "Run a model in REPL mode",
		Long:  "Run a model in REPL mode. The server must be running and the model should be downloaded and cached locally.",
	}

	runCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	runCmd.Flags().SortFlags = false
	runCmd.Flags().BoolVarP(&disableStream, "disable-stream", "s", false, "disable stream mode")

	runCmd.Run = runFunc
	return runCmd
}

func runFunc(cmd *cobra.Command, args []string) {
	model := normalizeModelName(args[0])

	client := openai.NewClient(
		option.WithBaseURL(fmt.Sprintf("http://%s/v1", config.Get().Host)),
		// option.WithRequestTimeout(time.Second*15),
	)

	// warm up
	spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
	spin.Start()
	_, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: nil,
		Model:    model,
	})
	spin.Stop()

	if _, ok := err.(net.Error); ok {
		fmt.Printf("Is server running? Please check your network. \n\t%s\n", err)
		return
	} else if e, ok := err.(*openai.Error); ok {
		if e.StatusCode == http.StatusNotFound {
			fmt.Printf("Model not found, please download first")
		} else {
			fmt.Printf("%s\n", err)
		}
		return
	} else if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// repl
	var history []openai.ChatCompletionMessageParamUnion
	repl(ReplConfig{
		Stream:    !disableStream,
		ParseFile: false,

		Clear: func() {
			history = nil
			client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
				Messages: nil,
				Model:    model,
			})
		},

		Run: func(prompt string, images, audios []string) (string, error) {
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

		RunStream: func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error) {
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
