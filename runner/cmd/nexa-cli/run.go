package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/internal/config"
	"github.com/NexaAI/nexa-sdk/internal/render"
	"github.com/NexaAI/nexa-sdk/internal/store"
	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
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

	// check
	_, err := client.Models.Get(context.TODO(), model)
	if err != nil {
		if _, ok := err.(net.Error); ok {
			fmt.Println(text.FgRed.Sprintf("Is server running? Please check your network. \n\t%s", err))
			return
		}
		if e, ok := err.(*openai.Error); ok && e.StatusCode == http.StatusNotFound {
			// pull model
			fmt.Println(text.FgBlue.Sprintf("model not found, start download"))

			// download manifest
			spin := render.NewSpinner("download manifest from: " + model)
			spin.Start()
			files, err := store.Get().HFModelInfo(context.TODO(), model)
			spin.Stop()
			if err != nil {
				fmt.Println(text.FgRed.Sprintf("Get manifest from huggingface error: %s", err))
				return
			}

			manifest, err := chooseFiles(model, files)
			if err != nil {
				return
			}

			var raw *http.Response
			err = client.Post(context.TODO(), "/models", nil, &raw,
				option.WithJSONSet("Name", manifest.Name),
				option.WithJSONSet("Size", manifest.GetSize()),
				option.WithJSONSet("ModelFile", manifest.ModelFile),
				option.WithJSONSet("MMProjFile", manifest.MMProjFile),
				option.WithJSONSet("ExtraFiles", manifest.ExtraFiles),
			)
			stream := ssestream.NewStream[types.DownloadInfo](ssestream.NewDecoder(raw), err)
			bar := render.NewProgressBar(manifest.GetSize(), "downloading")
			for stream.Next() {
				event := stream.Current()
				bar.Set(event.TotalDownloaded)
			}
			bar.Exit()

			if stream.Err() != nil {
				bar.Clear()
				fmt.Println(text.FgRed.Sprintf("pull model error: %s", stream.Err().Error()))
				return
			}
		} else {
			fmt.Println(text.FgRed.Sprintf("get model error: %s", err.Error()))
			return
		}
	}

	// warm up
	spin := render.NewSpinner("loading model...")
	spin.Start()
	_, err = client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Messages: nil,
		Model:    model,
	})
	spin.Stop()

	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	// repl
	var history []openai.ChatCompletionMessageParamUnion
	var profileData *nexa_sdk.ProfilingData
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
			profileData = nil
			start := time.Now()

			history = append(history, openai.UserMessage(prompt))

			chatCompletion, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
				Messages: history,
				Model:    model,
			})
			if err != nil {
				return "", err
			}

			if len(chatCompletion.Choices) == 0 {
				return "", fmt.Errorf("response empty")
			}

			content := chatCompletion.Choices[0].Message.Content

			history = append(history, openai.AssistantMessage(content))

			profileData = &nexa_sdk.ProfilingData{
				TotalTimeUs: int64(time.Since(start).Microseconds()),
			}

			return content, err
		},

		RunStream: func(ctx context.Context, prompt string, images, audios []string, dataCh chan<- string, errCh chan<- error) {
			defer close(errCh)
			defer close(dataCh)

			profileData = nil
			start := time.Now()

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

			if len(acc.Choices) > 0 {
				history = append(history, openai.AssistantMessage(acc.Choices[0].Message.Content))
			}

			profileData = &nexa_sdk.ProfilingData{
				TotalTimeUs: int64(time.Since(start).Microseconds()),
			}
		},

		GetProfilingData: func() (*nexa_sdk.ProfilingData, error) {
			if profileData == nil {
				return nil, fmt.Errorf("do not have profiling data")
			}
			return profileData, nil
		},
	})
}
