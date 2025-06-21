package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/config"
)

func run() *cobra.Command {
	runCmd := &cobra.Command{}
	runCmd.Use = "run"

	runCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	stream := runCmd.Flags().BoolP("stream", "s", false, "enable stream mode")
	runCmd.Run = func(cmd *cobra.Command, args []string) {
		client := openai.NewClient(
			option.WithBaseURL(fmt.Sprintf("http://%s/v1", config.Get().Host)),
		)

		// warm up
		spin := spinner.New(spinner.CharSets[39], 100*time.Millisecond, spinner.WithSuffix("loading model..."))
		spin.Start()
		_, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
			Messages: nil,
			Model:    args[0],
		})
		spin.Stop()

		if err != nil {
			fmt.Printf("Request Error: %s\n", err)
			return
		}

		// repl
		var history []openai.ChatCompletionMessageParamUnion
		r := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("\nEnter text: ")
			txt, e := r.ReadString('\n')
			if e != nil {
				fmt.Printf("ReadString Error: %s\n", e)
				break
			}

			history = append(history, openai.UserMessage(txt))

			if *stream {
				start := time.Now()
				var count int

				acc := openai.ChatCompletionAccumulator{}
				stream := client.Chat.Completions.NewStreaming(context.TODO(), openai.ChatCompletionNewParams{
					Messages: history,
					Model:    args[0],
				})

				fmt.Print("\033[33m")
				for stream.Next() {
					chunk := stream.Current()
					acc.AddChunk(chunk)
					if len(chunk.Choices) > 0 {
						fmt.Print(chunk.Choices[0].Delta.Content)
						count += 1
					}
				}
				if stream.Err() != nil {
					fmt.Printf("\nGenerate error: %s\n", stream.Err())
					return
				}
				fmt.Print("\033[0m\n")
				duration := time.Since(start).Seconds()

				fmt.Print("\033[34m")
				fmt.Printf("\nGenerate %d token in %f s, speed is %f token/s\n",
					count,
					duration,
					float64(count)/duration)
				fmt.Print("\033[0m")

				history = append(history, openai.AssistantMessage(acc.Choices[0].Message.Content))

			} else {

				start := time.Now()
				chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
					Messages: history,
					Model:    args[0],
				})
				if err != nil {
					fmt.Printf("Request Error: %s\n", e)
					break
				}
				content := chatCompletion.Choices[0].Message.Content
				fmt.Print("\033[33m")
				println(content)
				fmt.Print("\033[0m\n")

				duration := time.Since(start).Seconds()
				fmt.Print("\033[34m")
				fmt.Printf("\nGenerate in %f s\n", duration)
				fmt.Print("\033[0m")

				history = append(history, openai.AssistantMessage(content))
			}

		}
	}
	return runCmd
}
