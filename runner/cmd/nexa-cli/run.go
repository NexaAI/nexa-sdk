package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

func run() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run <model-name>",
		Short: "Run a model in REPL mode",
		Long:  "Run a model in REPL mode. The server must be running and the model should be downloaded and cached locally.",
	}

	runCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	runCmd.Flags().SortFlags = false

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
	modelInfo, err := client.Models.Get(context.TODO(), model)
	var manifest types.ModelManifest
	sonic.UnmarshalString(modelInfo.RawJSON(), &manifest)
	if err != nil {
		if _, ok := err.(net.Error); ok {
			fmt.Println(render.GetTheme().Error.Sprintf("Is server running? Please check your network. \n\t%s", err))
			return
		}
		if e, ok := err.(*openai.Error); ok && e.StatusCode == http.StatusNotFound {
			// pull model
			fmt.Println(render.GetTheme().Info.Sprintf("model not found, start download"))

			// download manifest
			spin := render.NewSpinner("download manifest from: " + model)
			spin.Start()
			files, err := store.Get().HFModelInfo(context.TODO(), model)
			spin.Stop()
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Get manifest from huggingface error: %s", err))
				return
			}

			modelType, err := chooseModelType()
			if err != nil {
				return
			}

			manifest, err := chooseFiles(model, files)
			if err != nil {
				return
			}

			var raw *http.Response
			err = client.Post(context.TODO(), "/models", nil, &raw,
				option.WithJSONSet("Name", manifest.Name),
				option.WithJSONSet("ModelType", modelType),
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
				fmt.Println(render.GetTheme().Error.Sprintf("pull model error: %s", stream.Err().Error()))
				return
			}
		} else {
			fmt.Println(render.GetTheme().Error.Sprintf("get model error: %s", err.Error()))
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
	repl(ReplConfig{
		ParseFile: manifest.ModelType == types.ModelTypeVLM,

		Reset: func() error {
			history = nil
			_, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
				Messages: nil,
				Model:    model,
			})
			return err
		},

		Run: func(prompt string, images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
			if len(images) > 0 || len(audios) > 0 {
				contents := make([]openai.ChatCompletionContentPartUnionParam, 0)
				contents = append(contents, openai.ChatCompletionContentPartUnionParam{
					OfText: &openai.ChatCompletionContentPartTextParam{
						Text: prompt,
					},
				})
				for _, image := range images {
					contents = append(contents, openai.ChatCompletionContentPartUnionParam{
						OfImageURL: &openai.ChatCompletionContentPartImageParam{
							ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
								URL: image,
							},
						},
					})
				}
				for _, audio := range audios {
					contents = append(contents, openai.ChatCompletionContentPartUnionParam{
						OfInputAudio: &openai.ChatCompletionContentPartInputAudioParam{
							InputAudio: openai.ChatCompletionContentPartInputAudioInputAudioParam{
								Data: audio,
							},
						},
					})
				}
				history = append(history, openai.UserMessage(contents))
			} else {
				history = append(history, openai.UserMessage(prompt))
			}

			acc := openai.ChatCompletionAccumulator{}
			stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
				Messages: history,
				Model:    model,
			}, option.WithHeaderAdd("Nexa-KeepCache", "false"))

			for stream.Next() {
				chunk := stream.Current()
				acc.AddChunk(chunk)
				if len(chunk.Choices) > 0 {
					if !on_token(chunk.Choices[0].Delta.Content) {
						stream.Close()
						break
					}
					acc.AddChunk(chunk)
				}
			}

			if len(acc.Choices) > 0 {
				history = append(history, openai.AssistantMessage(acc.Choices[0].Message.Content))
				return acc.Choices[0].Message.Content, nexa_sdk.ProfileData{}, nil
			}

			return "", nexa_sdk.ProfileData{}, nil

		},
	})
}
