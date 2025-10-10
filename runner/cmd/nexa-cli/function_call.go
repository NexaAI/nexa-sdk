package main

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

var tool []string

func functionCall() *cobra.Command {
	fcCmd := &cobra.Command{
		GroupID: "inference",
		Use:     "functioncall <model-name>",
		Aliases: []string{"fc"},
		Short:   "Function call with a model",
		Long:    "Run function call with a specified model. The model must be downloaded and cached locally.",
	}

	fcCmd.Args = cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs)

	fcCmd.Flags().SortFlags = false
	fcCmd.Flags().StringArrayVarP(&tool, "tool", "t", nil, "[llm|vlm] add function name for function call")
	fcCmd.Flags().StringArrayVarP(&prompt, "prompt", "p", nil, "[llm|vlm] pass prompt")

	fcCmd.Run = func(cmd *cobra.Command, args []string) {
		s := store.Get()

		manifest, err := ensureModelAvailable(s, args[0])
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("check model error: %s", err))
			return
		}

		quant, err := selectQuant(manifest)
		if err != nil {
			fmt.Println(render.GetTheme().Error.Sprintf("quant error: %s", err))
			return
		}
		fmt.Println(render.GetTheme().Quant.Sprintf("🔹 Quant=%s", quant))

		nexa_sdk.Init()
		defer nexa_sdk.DeInit()

		modelfile := s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name)

		switch manifest.ModelType {
		case types.ModelTypeLLM:
			fcLLM(manifest.PluginId, modelfile)
		case types.ModelTypeVLM:
			var mmprojfile string
			if manifest.MMProjFile.Name != "" {
				mmprojfile = s.ModelfilePath(manifest.Name, manifest.MMProjFile.Name)
			}
			var tokenizerfile string
			if manifest.TokenizerFile.Name != "" {
				tokenizerfile = s.ModelfilePath(manifest.Name, manifest.TokenizerFile.Name)
			}
			fcVLM(manifest.PluginId, modelfile, mmprojfile, tokenizerfile)
		default:
			panic("not support model type")
		}
	}
	return fcCmd
}

func checkParseTools(tools []string) (string, error) {
	if len(tools) == 0 {
		return "", fmt.Errorf("no tools provided")
	}
	return "[" + strings.Join(tools, ",") + "]", nil
}

func fcLLM(plugin, modelfile string) {
	tools, err := checkParseTools(tool)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprint(err))
		return
	}

	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
		ModelPath: modelfile,
		PluginID:  plugin,
		Config: nexa_sdk.ModelConfig{
			NCtx:       nctx,
			NGpuLayers: ngl,
		},
	})
	spin.Stop()

	if err != nil {
		slog.Error("failed to create LLM", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	messages := make([]nexa_sdk.LlmChatMessage, len(prompt))
	for i, p := range prompt {
		messages[i] = nexa_sdk.LlmChatMessage{Role: nexa_sdk.LLMRoleUser, Content: p}
	}
	templateOutput, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
		Messages:    messages,
		EnableThink: false, // disable thinking mode for function call mode
		Tools:       tools,
	})
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("apply chat template error: %s", err))
		return
	}
	res, err := p.Generate(nexa_sdk.LlmGenerateInput{
		PromptUTF8: templateOutput.FormattedText,
		Config: &nexa_sdk.GenerationConfig{
			MaxTokens: 2048,
		},
	},
	)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("generate error: %s", err))
		return
	}

	fmt.Println()
	fmt.Println(render.GetTheme().Success.Sprintf("%s", res.FullText))
	fmt.Println()
	printProfile(res.ProfileData)
}

func fcVLM(plugin, modelfile, mmprojfile, tokenizerfile string) {
	tools, err := checkParseTools(tool)
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprint(err))
		return
	}
	spin := render.NewSpinner("loading model...")
	spin.Start()
	p, err := nexa_sdk.NewVLM(nexa_sdk.VlmCreateInput{
		ModelPath:     modelfile,
		MmprojPath:    mmprojfile,
		TokenizerPath: tokenizerfile,
		PluginID:      plugin,
		Config: nexa_sdk.ModelConfig{
			NCtx:       nctx,
			NGpuLayers: ngl,
		},
	})
	spin.Stop()

	if err != nil {
		slog.Error("failed to create VLM", "error", err)
		fmt.Println(modelLoadFailMsg)
		return
	}
	defer p.Destroy()

	messages := make([]nexa_sdk.VlmChatMessage, len(prompt))
	for i, p := range prompt {
		messages[i] = nexa_sdk.VlmChatMessage{Role: nexa_sdk.VlmRoleUser, Contents: []nexa_sdk.VlmContent{{Type: nexa_sdk.VlmContentTypeText, Text: p}}}
	}
	templateOutput, err := p.ApplyChatTemplate(nexa_sdk.VlmApplyChatTemplateInput{
		Messages:    messages,
		EnableThink: false, // disable thinking mode for function call mode
		Tools:       tools,
	})
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("apply chat template error: %s", err))
		return
	}
	res, err := p.Generate(nexa_sdk.VlmGenerateInput{
		PromptUTF8: templateOutput.FormattedText,
		Config: &nexa_sdk.GenerationConfig{
			MaxTokens: 2048,
		},
	})
	if err != nil {
		fmt.Println(render.GetTheme().Error.Sprintf("generate error: %s", err))
		return
	}

	fmt.Println()
	fmt.Println(render.GetTheme().Success.Sprintf("%s", res.FullText))
	fmt.Println()
	printProfile(res.ProfileData)
}
