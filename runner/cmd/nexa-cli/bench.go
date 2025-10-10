package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/internal/render"
	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/spf13/cobra"
)

func bench() *cobra.Command {
	benchCmd := &cobra.Command{
		Use:    "bench",
		Short:  "Run benchmark tests",
		Hidden: true,
		Run:    benchFunc,
	}
	return benchCmd
}

type model struct {
	PluginID   string
	ModelType  string
	ModelName  string
	TestParams []param
}

type param struct {
	TestName string

	modelConfig   nexa_sdk.ModelConfig
	samplerConfig nexa_sdk.SamplerConfig

	// llm only
	llmChatRounds     [][]nexa_sdk.LlmChatMessage
	llmGenerateConfig nexa_sdk.GenerationConfig
}

type environment struct {
	OS            string `json:"os"`
	Arch          string `json:"arch"`
	SDKVersion    string `json:"sdk_version"`
	BridgeVersion string `json:"bridge_version"`
}

type roundResult struct {
	Err         error
	RoundInput  []nexa_sdk.LlmChatMessage
	RoundOutput nexa_sdk.LlmChatMessage
	ProfileData nexa_sdk.ProfileData
}

type result struct {
	Env environment
	Err error

	PluginID    string
	DeviceID    string
	ModelName   string
	RoundResult []roundResult
}

var llmDefaultTestParams = param{
	TestName: "logic-heavy-16rounds",
	modelConfig: nexa_sdk.ModelConfig{
		NCtx:       4096,
		NGpuLayers: 999,
	},
	samplerConfig: nexa_sdk.SamplerConfig{},
	llmGenerateConfig: nexa_sdk.GenerationConfig{
		MaxTokens: 2048,
	},
	llmChatRounds: [][]nexa_sdk.LlmChatMessage{
		{{Role: "user", Content: "You are a technical assistant. I have an array of integers: [3,1,4,1,5,9,2,6,5]. Describe the step-by-step approach to sort it without using built-in sort."}},
		{{Role: "user", Content: "Now provide the fully sorted array and show intermediate states after each swap (brief)."}},
		{{Role: "user", Content: "Return the final array as a JSON list only, no extra text."}},
		{{Role: "user", Content: "Using the same original array, compute the median and explain how you derive it in two sentences."}},
		{{Role: "user", Content: "Suppose these numbers arrive as a stream. Describe an O(1) memory approach to approximate median (one paragraph)."}},
		{{Role: "user", Content: "Summarize all previous algorithm steps in a single bullet-point list."}},
		{{Role: "user", Content: "Convert that bullet list into a one-line log message suitable for automated CI logs."}},
		{{Role: "user", Content: "Generate a Go unit test (table-driven) that asserts the sorted output equals [1,1,2,3,4,5,5,6,9]. Provide code only."}},
		{{Role: "user", Content: "Change the input to include negatives: [-2,0,3,-1]. Give the sorted result JSON."}},
		{{Role: "user", Content: "Explain how time complexity scales with N for the algorithm you described."}},
		{{Role: "user", Content: "Present pseudocode for a parallelized version of the algorithm (concise)."}},
		{{Role: "user", Content: "List edge cases that must be tested for the sorter (max 6 bullet points)."}},
		{{Role: "user", Content: "If the input is nearly sorted (each element at most 2 positions off), which algorithm is optimal and why? One paragraph."}},
		{{Role: "user", Content: "Provide a minimal failing test vector that would detect a stability bug (return as JSON object {input:..., expected:...})."}},
		{{Role: "user", Content: "Assume version A produced JSON [1,1,2,3,4,5,5,6,9] and version B produced [1,1,2,3,4,5,6,5,9]. Produce a one-line human-readable diff summary."}},
		{{Role: "user", Content: "Return only the word 'PASS' if the two arrays are identical, else return 'FAIL'."}},
	},
}

var models = []model{
	{"cpu_gpu", "llm", "Qwen/Qwen3-0.6B-GGUF", []param{llmDefaultTestParams}},
	{"cpu_gpu", "llm", "unsloth/Qwen3-1.7B-GGUF", []param{llmDefaultTestParams}},
	{"npu", "llm", "NexaAI/Llama3.2-3B-NPU-Turbo", []param{llmDefaultTestParams}},
}

func benchFunc(cmd *cobra.Command, args []string) {
	plugins := getHostPlugins()

	plugin_models := make(map[string][]*model)
	for _, plugin := range plugins {
		i := sort.Search(len(models), func(i int) bool {
			return models[i].PluginID == plugin
		})
		plugin_models[plugin] = append(plugin_models[plugin], &models[i])
	}

	needDownload := []string{}
	s := store.Get()
	for _, model := range models {
		_, err := s.GetManifest(model.ModelName)
		if err != nil {
			needDownload = append(needDownload, model.ModelName)
		}
	}

	if len(needDownload) > 0 {
		fmt.Println(render.GetTheme().Warning.Sprintf("models not downloaded, run:"))
		for _, model := range needDownload {
			fmt.Println(render.GetTheme().Warning.Sprintf("  nexa pull %s", model))
		}
		return
	}

	nexa_sdk.Init()
	defer nexa_sdk.DeInit()

	for plugin, models := range plugin_models {
		slog.Info("bench plugin", "plugin", plugin)

		for _, model := range models {
			manifest, err := s.GetManifest(model.ModelName)
			if err != nil {
				fmt.Println(render.GetTheme().Error.Sprintf("get manifest error: %s", err))
				return
			}
			for i, testCase := range model.TestParams {
				slog.Info("bench model", "model", model.ModelName, "testCase", i)

				var res result
				switch manifest.ModelType {
				case types.ModelTypeLLM:
					quant, err := selectQuant(manifest)
					if err != nil {
						fmt.Println(render.GetTheme().Error.Sprintf("select quant error: %s", err))
						return
					}
					res, err = benchLLM(manifest, quant, &testCase)
					if err != nil {
						fmt.Println(render.GetTheme().Error.Sprintf("bench LLM error: %s", err))
						return
					}
				case types.ModelTypeVLM:
					panic("not supported")
				default:
					panic("not supported")
				}

				filename := fmt.Sprintf("%s_%s.json",
					strings.ReplaceAll(model.ModelName, "/", "_"),
					testCase.TestName)

				date := time.Now().Format("20060102")
				saveResult(res, path.Join("tests", date, filename))
			}
		}
	}
}

func getHostPlugins() []string {
	os := runtime.GOOS
	arch := runtime.GOARCH
	switch {
	case os == "windows" && arch == "arm64":
		return []string{"cpu_gpu", "npu"}
	case os == "windows" && arch == "amd64":
		return []string{"cpu_gpu"}
	case os == "darwin" && arch == "arm64":
		panic("todo")
	case os == "linux" && arch == "amd64":
		panic("todo")
	}
	return nil
}

func getHostBenchList() map[string][]*model {
	plugins := getHostPlugins()
	plugin_models := make(map[string][]*model)
	for _, plugin := range plugins {
		i := sort.Search(len(models), func(i int) bool {
			return models[i].PluginID == plugin
		})
		plugin_models[plugin] = append(plugin_models[plugin], &models[i])
	}
	return plugin_models
}

func benchLLM(manifest *types.ModelManifest, quant string, param *param) (res result, err error) {
	s := store.Get()

	res = result{
		Env: environment{
			OS:            runtime.GOOS,
			Arch:          runtime.GOARCH,
			SDKVersion:    Version,
			BridgeVersion: nexa_sdk.Version(),
		},
		PluginID:  manifest.PluginId,
		DeviceID:  manifest.DeviceId,
		ModelName: manifest.Name,
	}

	llm, err := nexa_sdk.NewLLM(nexa_sdk.LlmCreateInput{
		ModelName: manifest.ModelName,
		ModelPath: s.ModelfilePath(manifest.Name, manifest.ModelFile[quant].Name),
		Config:    param.modelConfig,
		PluginID:  manifest.PluginId,
		DeviceID:  manifest.DeviceId,
	})
	if err != nil {
		res.Err = fmt.Errorf("create LLM error: %w", err)
		return res, res.Err
	}
	defer llm.Destroy()

	chatHistory := []nexa_sdk.LlmChatMessage{}
	for i, msg := range param.llmChatRounds {
		slog.Info("chat", "i", i, "msg", msg, "chatHistory", chatHistory)
		chatHistory = append(chatHistory, msg...)

		tpl, err := llm.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
			Messages: chatHistory,
		})
		if err != nil {
			roundResult := roundResult{
				Err:        fmt.Errorf("apply chat template error: %w", err),
				RoundInput: msg,
			}
			res.RoundResult = append(res.RoundResult, roundResult)
			continue
		}

		output, err := llm.Generate(nexa_sdk.LlmGenerateInput{
			PromptUTF8: tpl.FormattedText,
			Config:     &param.llmGenerateConfig,
			OnToken:    func(string) bool { return true },
		})
		if err != nil {
			roundResult := roundResult{
				Err:        fmt.Errorf("generate error: %w", err),
				RoundInput: msg,
			}
			res.RoundResult = append(res.RoundResult, roundResult)
			continue
		}

		assistantMsg := nexa_sdk.LlmChatMessage{
			Role:    nexa_sdk.LLMRoleAssistant,
			Content: output.FullText,
		}
		chatHistory = append(chatHistory, assistantMsg)

		roundResult := roundResult{
			Err:         nil,
			RoundInput:  msg,
			RoundOutput: assistantMsg,
			ProfileData: output.ProfileData,
		}
		res.RoundResult = append(res.RoundResult, roundResult)

		p := output.ProfileData
		fmt.Printf("Generated: %s\n", output.FullText)
		fmt.Printf("Profile: TTFT=%dμs, DecodeTime=%dμs, Tokens=%d, Speed=%.2f tokens/s\n",
			p.TTFT, p.DecodeTime, p.GeneratedTokens, p.DecodingSpeed)
	}

	return res, nil
}

func saveResult(res result, filename string) {
	data, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal result: %v\n", err)
		return
	}

	dir := path.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("Failed to create tests directory: %v\n", err)
		return
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		fmt.Printf("Failed to write result file: %v\n", err)
		return
	}
	fmt.Printf("Result saved to: %s\n", filename)
}
