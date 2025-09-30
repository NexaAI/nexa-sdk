package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/bytedance/sonic"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/model_hub"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

func run() *cobra.Command {
	runCmd := &cobra.Command{
		GroupID: "inference",
		Use:     "run <model-name>",
		Short:   "Run a model in REPL mode",
		Long:    "Run a model in REPL mode. The server must be running and the model should be downloaded and cached locally.",
	}

	runCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	runCmd.Flags().AddFlagSet(samplerFlags)

	runCmd.SetUsageFunc(func(c *cobra.Command) error {
		flagGroups := []*pflag.FlagSet{
			samplerFlags,
		}
		w := c.OutOrStdout()
		fmt.Fprint(w, "Usage:")
		if c.Runnable() {
			fmt.Fprintf(w, "\n  %s", c.UseLine())
		}
		if len(c.Aliases) > 0 {
			fmt.Fprintf(w, "\n\nAliases:\n")
			fmt.Fprintf(w, "  %s", c.NameAndAliases())
		}

		for _, flags := range flagGroups {
			fmt.Fprintf(w, "\n\n%s Flags:\n", flags.Name())
			fmt.Fprint(w, strings.TrimRightFunc(flags.FlagUsages(), unicode.IsSpace))
		}

		if c.HasAvailableInheritedFlags() {
			fmt.Fprintf(w, "\n\nGlobal Flags:\n")
			fmt.Fprint(w, strings.TrimRightFunc(c.InheritedFlags().FlagUsages(), unicode.IsSpace))
		}
		fmt.Fprintln(w)
		return nil
	})

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
			files, hmf, err := model_hub.ModelInfo(context.TODO(), model)
			spin.Stop()
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Get manifest from huggingface error: %s", err))
				return
			}

			if hmf != nil && !isValidVersion(hmf.MinSDKVersion) {
				fmt.Println(render.GetTheme().Error.Sprintf("Model requires NexaSDK CLI version %s or higher. Please upgrade your NexaSDK CLI.", hmf.MinSDKVersion))
				return
			}

			var manifest types.ModelManifest

			if hmf != nil {
				manifest.ModelName = hmf.ModelName
				manifest.PluginId = hmf.PluginId
				manifest.DeviceId = hmf.DeviceId
				manifest.ModelType = hmf.ModelType
				manifest.MinSDKVersion = hmf.MinSDKVersion
			}

			if manifest.ModelName == "" {
				manifest.ModelName = model
			}
			if manifest.PluginId == "" {
				manifest.PluginId = choosePluginId(model)
			}
			if manifest.DeviceId == "" {
				manifest.DeviceId = hmf.DeviceId
			}
			if manifest.ModelType == "" {
				if ctype, err := chooseModelType(); err != nil {
					fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
					return
				} else {
					manifest.ModelType = ctype
				}
			}

			err = chooseFiles(model, files, &manifest)
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("Error: %s", err))
				return
			}

			var raw *http.Response
			err = client.Post(context.TODO(), "/models", nil, &raw,
				option.WithJSONSet("Name", manifest.Name),
				option.WithJSONSet("ModelName", manifest.ModelName),
				option.WithJSONSet("PluginId", manifest.PluginId),
				option.WithJSONSet("DeviceID", manifest.DeviceId),
				option.WithJSONSet("ModelType", manifest.ModelType),
				option.WithJSONSet("MinSDKVersion", manifest.MinSDKVersion),
				option.WithJSONSet("ModelFile", manifest.ModelFile),
				option.WithJSONSet("MMProjFile", manifest.MMProjFile),
				option.WithJSONSet("TokenizerFile", manifest.TokenizerFile),
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

			// check again
			modelInfo, err = client.Models.Get(context.TODO(), model)
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("get model error: %s", "download is incorrect"))
				return
			}
		} else {
			fmt.Println(render.GetTheme().Error.Sprintf("get model error: %s", err.Error()))
			return
		}
	}

	var manifest types.ModelManifest
	sonic.UnmarshalString(modelInfo.RawJSON(), &manifest)

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

			start := time.Now()
			acc := openai.ChatCompletionAccumulator{}
			stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
				Messages:         history,
				Model:            model,
				StreamOptions:    openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Opt(true)},
				Temperature:      openai.Float(float64(temperature)),
				TopP:             openai.Float(float64(topP)),
				PresencePenalty:  openai.Float(float64(presencePenalty)),
				FrequencyPenalty: openai.Float(float64(frequencyPenalty)),
				Seed:             openai.Int(int64(seed)),
			},
				option.WithJSONSet("enable_json", enableJson),
				option.WithHeaderAdd("Nexa-KeepCache", "false"))

			var firstToken time.Time
			var profileData nexa_sdk.ProfileData
			for stream.Next() {
				if firstToken.IsZero() {
					firstToken = time.Now()
				}

				chunk := stream.Current()
				acc.AddChunk(chunk)
				if len(chunk.Choices) > 0 {
					if !on_token(chunk.Choices[0].Delta.Content) {
						stream.Close()
						break
					}
					acc.AddChunk(chunk)
				}
				if chunk.Usage.PromptTokens > 0 {
					profileData.PromptTokens = chunk.Usage.PromptTokens
					profileData.GeneratedTokens = chunk.Usage.CompletionTokens
				}
			}
			end := time.Now()
			profileData.TTFT = firstToken.Sub(start).Microseconds()
			profileData.PromptTime = 0
			profileData.DecodeTime = end.Sub(firstToken).Microseconds()
			profileData.DecodingSpeed = float64(profileData.GeneratedTokens) / float64(end.Sub(firstToken).Seconds())

			if len(acc.Choices) > 0 {
				history = append(history, openai.AssistantMessage(acc.Choices[0].Message.Content))
				return acc.Choices[0].Message.Content, profileData, nil
			}

			return "", profileData, nil
		},
	})
}
