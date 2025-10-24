package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"unicode"

	"github.com/bytedance/sonic"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
)

func run() *cobra.Command {
	runCmd := &cobra.Command{
		GroupID: "inference",
		Use:     "run <model-name>",
		Short:   "Infer a model with server",
		Long:    "Infer a model with server. The server must be running and the model should be downloaded and cached locally.",
	}

	runCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)
	for _, flags := range flagGroups {
		runCmd.Flags().AddFlagSet(flags)
	}

	runCmd.SetUsageFunc(func(c *cobra.Command) error {
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

	runCmd.Run = func(cmd *cobra.Command, args []string) {
		name, quant := normalizeModelName(args[0])
		if quant != "" {
			name = name + ":" + quant
		}

		client := openai.NewClient(
			option.WithBaseURL(fmt.Sprintf("http://%s/v1", config.Get().Host)),
			// option.WithRequestTimeout(time.Second*15),
		)

		// check
		modelInfo, err := client.Models.Get(context.TODO(), name)
		if err != nil {
			if _, ok := err.(net.Error); ok {
				fmt.Println(render.GetTheme().Error.Sprintf("Is server running? Please check your network. \n\t%s", err))
				os.Exit(1)
			}
			if e, ok := err.(*openai.Error); ok && e.StatusCode == http.StatusNotFound {
				fmt.Println(render.GetTheme().Error.Sprintf("Model not found: %s, Please download first", name))
				os.Exit(1)
			} else {
				fmt.Println(render.GetTheme().Error.Sprintf("get model error: %s", err.Error()))
				os.Exit(1)
			}
		}

		var manifest types.ModelManifest
		sonic.UnmarshalString(modelInfo.RawJSON(), &manifest)

		switch manifest.ModelType {
		case types.ModelTypeLLM, types.ModelTypeVLM:
			err = runCompletion(manifest, quant)
		// case types.ModelTypeVLM:
		// 	checkDependency()
		// 	err = inferVLM(manifest, quant)
		// case types.ModelTypeEmbedder:
		// 	err = inferEmbedder(manifest, quant)
		// case types.ModelTypeReranker:
		// 	err = inferReranker(manifest, quant)
		// case types.ModelTypeTTS:
		// 	err = inferTTS(manifest, quant)
		// case types.ModelTypeASR:
		// 	checkDependency()
		// 	err = inferASR(manifest, quant)
		// case types.ModelTypeDiarize:
		// 	err = inferDiarize(manifest, quant)
		// case types.ModelTypeCV:
		// 	err = inferCV(manifest, quant)
		// case types.ModelTypeImageGen:
		// 	// ImageGen model is a directory, not a file
		// 	err = inferImageGen(manifest, quant)
		default:
			panic("not support model type")
		}

		if err != nil {
			os.Exit(1)
		}
	}
	return runCmd
}

func runCompletion(manifest types.ModelManifest, quant string) error {
	// // warm up
	// spin := render.NewSpinner("loading model...")
	// spin.Start()
	// warmUpRequest := openai.ChatCompletionNewParams{
	// 	Messages: nil,
	// 	Model:    name,
	// }
	// if systemPrompt != "" {
	// 	warmUpRequest.Messages = append(warmUpRequest.Messages, openai.SystemMessage(systemPrompt))
	// }
	// _, err = client.Chat.Completions.New(context.TODO(), warmUpRequest)
	// spin.Stop()
	//
	// if err != nil {
	// 	fmt.Printf("%s\n", err)
	// 	os.Exit(1)
	// }
	//
	// // repl
	// var history []openai.ChatCompletionMessageParamUnion
	// if systemPrompt != "" {
	// 	history = append(history, openai.SystemMessage(systemPrompt))
	// }
	//
	// processor := &common.Processor{
	// 	ParseFile: manifest.ModelType == types.ModelTypeVLM,
	// 	TestMode:  testMode,
	// 	Run: func(prompt string, images, audios []string, on_token func(string) bool) (string, nexa_sdk.ProfileData, error) {
	// 		if len(images) > 0 || len(audios) > 0 {
	// 			contents := make([]openai.ChatCompletionContentPartUnionParam, 0)
	// 			contents = append(contents, openai.ChatCompletionContentPartUnionParam{
	// 				OfText: &openai.ChatCompletionContentPartTextParam{
	// 					Text: prompt,
	// 				},
	// 			})
	// 			for _, image := range images {
	// 				contents = append(contents, openai.ChatCompletionContentPartUnionParam{
	// 					OfImageURL: &openai.ChatCompletionContentPartImageParam{
	// 						ImageURL: openai.ChatCompletionContentPartImageImageURLParam{
	// 							URL: image,
	// 						},
	// 					},
	// 				})
	// 			}
	// 			for _, audio := range audios {
	// 				contents = append(contents, openai.ChatCompletionContentPartUnionParam{
	// 					OfInputAudio: &openai.ChatCompletionContentPartInputAudioParam{
	// 						InputAudio: openai.ChatCompletionContentPartInputAudioInputAudioParam{
	// 							Data: audio,
	// 						},
	// 					},
	// 				})
	// 			}
	// 			history = append(history, openai.UserMessage(contents))
	// 		} else {
	// 			history = append(history, openai.UserMessage(prompt))
	// 		}
	//
	// 		start := time.Now()
	// 		acc := openai.ChatCompletionAccumulator{}
	// 		stream := client.Chat.Completions.NewStreaming(context.Background(), openai.ChatCompletionNewParams{
	// 			Messages:         history,
	// 			Model:            name,
	// 			StreamOptions:    openai.ChatCompletionStreamOptionsParam{IncludeUsage: openai.Opt(true)},
	// 			Temperature:      openai.Float(float64(temperature)),
	// 			TopP:             openai.Float(float64(topP)),
	// 			PresencePenalty:  openai.Float(float64(presencePenalty)),
	// 			FrequencyPenalty: openai.Float(float64(frequencyPenalty)),
	// 			Seed:             openai.Int(int64(seed)),
	// 		},
	//
	// 			option.WithJSONSet("enable_think", enableThink),
	// 			option.WithJSONSet("top_k", topK),
	// 			option.WithJSONSet("min_p", minP),
	// 			option.WithJSONSet("repetition_penalty", repetitionPenalty),
	// 			option.WithJSONSet("grammar_path", grammarPath),
	// 			option.WithJSONSet("grammar_string", grammarString),
	// 			option.WithJSONSet("enable_json", enableJson),
	// 			option.WithHeaderAdd("Nexa-KeepCache", "true"))
	//
	// 		var firstToken time.Time
	// 		var profileData nexa_sdk.ProfileData
	// 		for stream.Next() {
	// 			if firstToken.IsZero() {
	// 				firstToken = time.Now()
	// 			}
	//
	// 			chunk := stream.Current()
	// 			acc.AddChunk(chunk)
	// 			if len(chunk.Choices) > 0 {
	// 				if !on_token(chunk.Choices[0].Delta.Content) {
	// 					stream.Close()
	// 					break
	// 				}
	// 			}
	// 			if chunk.Usage.PromptTokens > 0 {
	// 				profileData.PromptTokens = chunk.Usage.PromptTokens
	// 				profileData.GeneratedTokens = chunk.Usage.CompletionTokens
	// 			}
	// 		}
	// 		if stream.Err() != nil {
	// 			return "", profileData, stream.Err()
	// 		}
	//
	// 		// zero token generated
	// 		if firstToken.IsZero() {
	// 			firstToken = time.Now()
	// 		}
	//
	// 		end := time.Now()
	// 		profileData.TTFT = firstToken.Sub(start).Microseconds()
	// 		profileData.PromptTime = profileData.TTFT
	// 		profileData.DecodeTime = end.Sub(firstToken).Microseconds()
	// 		profileData.DecodingSpeed = float64(profileData.GeneratedTokens) / float64(end.Sub(firstToken).Seconds())
	//
	// 		if len(acc.Choices) > 0 {
	// 			history = append(history, openai.AssistantMessage(acc.Choices[0].Message.Content))
	// 			return acc.Choices[0].Message.Content, profileData, nil
	// 		}
	//
	// 		return "", profileData, nil
	// 	},
	// }
	// repl := common.Repl{
	// 	Reset: func() error {
	// 		history = nil
	// 		_, err := client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
	// 			Messages: nil,
	// 			Model:    name,
	// 		})
	// 		return err
	// 	},
	// }
	// defer repl.Close()
	// processor.GetPrompt = repl.GetPrompt
	// processor.Process()
	return nil
}
