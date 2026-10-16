// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/cmd/geniex/common"
	"github.com/qualcomm/GenieX/cli/internal/config"
	decision "github.com/qualcomm/GenieX/cli/internal/decision"
	"github.com/qualcomm/GenieX/cli/internal/render"
	"github.com/qualcomm/GenieX/cli/internal/store"
)

type decideOptions struct {
	labels              []string
	options             []string
	instruction         string
	context             string
	contextFile         string
	images              []string
	confidence          bool
	minSelections       int
	maxSelections       int
	benchmark           int
	threads             int32
	threadsBatch        int32
	batch               int32
	ubatch              int32
	nSeqMax             int32
	chatTemplatePath    string
	chatTemplateContent string
	specType            string
	draftModel          string
	draftTokens         int32
	draftMin            int32
	draftPMin           float32
	powerMode           string
	stop                []string
	slidingWindow       bool
	slidingWindowNKeep  int32
	temperature         float32
	topP                float32
	topK                int32
	minP                float32
	repetitionPenalty   float32
	presencePenalty     float32
	frequencyPenalty    float32
	seed                int32
}

func decideCmd() *cobra.Command {
	options := decideOptions{}
	command := &cobra.Command{
		GroupID: "inference",
		Use:     "decide <model-name>[:<precision>] <classify|choose>",
		Short:   "Run a validated local structured decision",
		Long:    "Run one text or visual structured decision with a local model. Results are locally validated JSON candidates and never execute side effects.",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires a model name and a decision mode: classify or choose")
			}
			switch args[1] {
			case "classify", "choose":
				return nil
			default:
				return fmt.Errorf("unknown decision mode %q (expected classify or choose)", args[1])
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.benchmark < 0 {
				return fmt.Errorf("--benchmark must be at least 0")
			}
			if err := validateDecideRuntimeOptions(options); err != nil {
				return err
			}
			switch args[1] {
			case "classify":
				return runDecideClassify(cmd.Context(), args[0], options)
			case "choose":
				return runDecideChoose(cmd.Context(), args[0], options)
			}
			return nil
		},
	}
	command.Flags().AddFlagSet(decideModelFlags())
	command.Flags().StringArrayVar(&options.labels, "label", nil, "allowed classification label ID (repeatable)")
	command.Flags().StringArrayVar(&options.options, "option", nil, "allowed choice as id=text (repeatable)")
	command.Flags().StringVar(&options.instruction, "instruction", "", "task instruction")
	command.Flags().StringVar(&options.context, "context", "", "untrusted task context")
	command.Flags().StringVar(&options.contextFile, "context-file", "", "file containing untrusted task context")
	command.Flags().StringArrayVar(&options.images, "image", nil, "image path for a visual decision (repeatable)")
	command.Flags().BoolVar(&options.confidence, "confidence", false, "request advisory model-reported confidence")
	command.Flags().IntVar(&options.minSelections, "min-selections", 0, "minimum selected options (choose mode)")
	command.Flags().IntVar(&options.maxSelections, "max-selections", 0, "maximum selected options (choose mode)")
	command.Flags().IntVar(&options.benchmark, "benchmark", 0, "measure decision inference with this many repetitions after one warmup")
	command.Flags().Int32Var(&options.threads, "threads", 0, "model decode thread count; 0 uses the SDK default")
	command.Flags().Int32Var(&options.threadsBatch, "threads-batch", 0, "model prefill thread count; 0 uses the SDK default")
	command.Flags().Int32Var(&options.batch, "batch", 0, "model batch size; 0 uses the SDK default")
	command.Flags().Int32Var(&options.ubatch, "ubatch", 0, "model micro-batch size; 0 uses the SDK default")
	command.Flags().Int32Var(&options.nSeqMax, "n-seq-max", 0, "maximum concurrent model sequences; 0 uses the SDK default")
	command.Flags().StringVar(&options.chatTemplatePath, "chat-template-path", "", "path to a chat template override (llama_cpp only)")
	command.Flags().StringVar(&options.chatTemplateContent, "chat-template-content", "", "chat template override content (llama_cpp only)")
	command.Flags().StringVar(&options.specType, "spec-type", "", "speculative decoding type(s) (llama_cpp only)")
	command.Flags().StringVar(&options.draftModel, "draft-model", "", "draft model catalogue name or GGUF path (llama_cpp only)")
	command.Flags().Int32Var(&options.draftTokens, "draft-tokens", 3, "maximum draft tokens per step (llama_cpp only)")
	command.Flags().Int32Var(&options.draftMin, "draft-min", 0, "minimum draft tokens per step (llama_cpp only)")
	command.Flags().Float32Var(&options.draftPMin, "draft-p-min", 0, "minimum draft probability (llama_cpp only)")
	command.Flags().StringVar(&options.powerMode, "power-mode", "", "HTP power mode (default: burst)")
	command.Flags().StringArrayVar(&options.stop, "stop", nil, "stop sequence (repeatable; llama_cpp only)")
	command.Flags().BoolVar(&options.slidingWindow, "sliding-window", false, "evict oldest context on overflow (qairt only)")
	command.Flags().Int32Var(&options.slidingWindowNKeep, "sliding-window-n-keep", 0, "tokens to retain with sliding window; 0 uses plugin default")
	command.Flags().Float32Var(&options.temperature, "temperature", 0, "sampling temperature; default 0 is deterministic")
	command.Flags().Float32Var(&options.topP, "top-p", 0, "top-p sampling")
	command.Flags().Int32Var(&options.topK, "top-k", 0, "top-k sampling")
	command.Flags().Float32Var(&options.minP, "min-p", 0, "min-p sampling")
	command.Flags().Float32Var(&options.repetitionPenalty, "repetition-penalty", 0, "repetition penalty")
	command.Flags().Float32Var(&options.presencePenalty, "presence-penalty", 0, "presence penalty")
	command.Flags().Float32Var(&options.frequencyPenalty, "frequency-penalty", 0, "frequency penalty")
	command.Flags().Int32Var(&options.seed, "seed", 0, "random seed")
	command.Flags().SortFlags = false
	return command
}

func decideContext(options decideOptions) (string, error) {
	if options.context != "" && options.contextFile != "" {
		return "", fmt.Errorf("--context and --context-file cannot be used together")
	}
	if options.contextFile != "" {
		content, err := os.ReadFile(options.contextFile)
		if err != nil {
			return "", fmt.Errorf("read --context-file: %w", err)
		}
		return string(content), nil
	}
	if strings.TrimSpace(options.context) == "" {
		return "", fmt.Errorf("one of --context or --context-file is required")
	}
	return options.context, nil
}

func runDecideClassify(ctx context.Context, model string, options decideOptions) error {
	contextText, err := decideContext(options)
	if err != nil {
		return err
	}
	if err := validateDecideLabels(options.labels, options.instruction); err != nil {
		return err
	}
	return withDecideModel(ctx, model, options, func(paths *geniex_sdk.ModelPaths, llm *geniex_sdk.LLM, vlm *geniex_sdk.VLM) error {
		spec := decision.ClassificationSpec{Labels: options.labels, Instructions: options.instruction, Context: contextText, ImagePaths: options.images, MaxTokens: maxTokens, RequestConfidence: options.confidence}
		if llm != nil {
			decider := decision.LLMDecider{LLM: llm, RuntimeID: paths.RuntimeID, GenerationOptions: decideGenerationOptions(options)}
			result, timing, err := runDecideBenchmark(options.benchmark, llm.Reset, func() (any, decision.DecisionTiming, error) {
				return decider.ClassifyWithTiming(ctx, spec)
			})
			if err != nil {
				return err
			}
			printDecideBenchmark(options.benchmark, options, timing)
			return printDecideJSON(result)
		}
		decider := decision.VLMDecider{VLM: vlm, RuntimeID: paths.RuntimeID, GenerationOptions: decideGenerationOptions(options)}
		result, timing, err := runDecideBenchmark(options.benchmark, vlm.Reset, func() (any, decision.DecisionTiming, error) {
			return decider.ClassifyWithTiming(ctx, spec)
		})
		if err != nil {
			return err
		}
		printDecideBenchmark(options.benchmark, options, timing)
		return printDecideJSON(result)
	})
}

func runDecideChoose(ctx context.Context, model string, options decideOptions) error {
	contextText, err := decideContext(options)
	if err != nil {
		return err
	}
	parsed, err := parseDecideOptions(options.options)
	if err != nil {
		return err
	}
	if err := validateDecideOptions(parsed, options.instruction, options.minSelections, options.maxSelections); err != nil {
		return err
	}
	return withDecideModel(ctx, model, options, func(paths *geniex_sdk.ModelPaths, llm *geniex_sdk.LLM, vlm *geniex_sdk.VLM) error {
		spec := decision.MultipleChoiceSpec{Options: parsed, Instructions: options.instruction, Context: contextText, ImagePaths: options.images, MaxTokens: maxTokens, MinSelections: options.minSelections, MaxSelections: options.maxSelections, RequestConfidence: options.confidence}
		if llm != nil {
			decider := decision.LLMDecider{LLM: llm, RuntimeID: paths.RuntimeID, GenerationOptions: decideGenerationOptions(options)}
			result, timing, err := runDecideBenchmark(options.benchmark, llm.Reset, func() (any, decision.DecisionTiming, error) {
				return decider.ChooseWithTiming(ctx, spec)
			})
			if err != nil {
				return err
			}
			printDecideBenchmark(options.benchmark, options, timing)
			return printDecideJSON(result)
		}
		decider := decision.VLMDecider{VLM: vlm, RuntimeID: paths.RuntimeID, GenerationOptions: decideGenerationOptions(options)}
		result, timing, err := runDecideBenchmark(options.benchmark, vlm.Reset, func() (any, decision.DecisionTiming, error) {
			return decider.ChooseWithTiming(ctx, spec)
		})
		if err != nil {
			return err
		}
		printDecideBenchmark(options.benchmark, options, timing)
		return printDecideJSON(result)
	})
}

func parseDecideOptions(values []string) ([]decision.Option, error) {
	options := make([]decision.Option, 0, len(values))
	for _, value := range values {
		id, text, ok := strings.Cut(value, "=")
		if !ok || strings.TrimSpace(id) == "" || strings.TrimSpace(text) == "" {
			return nil, fmt.Errorf("--option must use id=text")
		}
		options = append(options, decision.Option{ID: id, Text: text})
	}
	return options, nil
}

func validateDecideRuntimeOptions(options decideOptions) error {
	for _, option := range []struct {
		name  string
		value int32
	}{
		{"--threads", options.threads},
		{"--threads-batch", options.threadsBatch},
		{"--batch", options.batch},
		{"--ubatch", options.ubatch},
		{"--n-seq-max", options.nSeqMax},
		{"--sliding-window-n-keep", options.slidingWindowNKeep},
		{"--draft-tokens", options.draftTokens},
		{"--draft-min", options.draftMin},
	} {
		if option.value < 0 {
			return fmt.Errorf("%s must be at least 0", option.name)
		}
	}
	if options.chatTemplatePath != "" && options.chatTemplateContent != "" {
		return fmt.Errorf("--chat-template-path and --chat-template-content cannot be used together")
	}
	if options.draftModel != "" && options.specType == "" {
		return fmt.Errorf("--draft-model requires --spec-type")
	}
	return nil
}

func decideGenerationOptions(options decideOptions) decision.GenerationOptions {
	return decision.GenerationOptions{
		Temperature: options.temperature, TopP: options.topP, TopK: options.topK, MinP: options.minP,
		RepetitionPenalty: options.repetitionPenalty, PresencePenalty: options.presencePenalty,
		FrequencyPenalty: options.frequencyPenalty, Seed: options.seed, Stop: options.stop,
		SlidingWindow: options.slidingWindow, SlidingWindowNKeep: options.slidingWindowNKeep,
	}
}

func decideModelConfig(nctx, ngl int32, options decideOptions, powerMode geniex_sdk.PowerMode, specDraftModel string) geniex_sdk.ModelConfig {
	return geniex_sdk.ModelConfig{
		NCtx: nctx, NThreads: options.threads, NThreadsBatch: options.threadsBatch,
		NBatch: options.batch, NUbatch: options.ubatch, NSeqMax: options.nSeqMax, NGpuLayers: ngl,
		ChatTemplatePath: options.chatTemplatePath, ChatTemplateContent: options.chatTemplateContent,
		SpecType: options.specType, SpecDraftModel: specDraftModel, SpecNMax: options.draftTokens,
		SpecNMin: options.draftMin, SpecPMin: options.draftPMin, PowerMode: powerMode,
	}
}

func validateDecideLabels(labels []string, instruction string) error {
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("--instruction is required")
	}
	if len(labels) == 0 {
		return fmt.Errorf("at least one --label is required")
	}
	seen := make(map[string]struct{}, len(labels))
	for _, label := range labels {
		if strings.TrimSpace(label) == "" {
			return fmt.Errorf("--label cannot be empty")
		}
		if _, duplicate := seen[label]; duplicate {
			return fmt.Errorf("duplicate --label %q", label)
		}
		seen[label] = struct{}{}
	}
	return nil
}

func validateDecideOptions(options []decision.Option, instruction string, min, max int) error {
	if strings.TrimSpace(instruction) == "" {
		return fmt.Errorf("--instruction is required")
	}
	if len(options) == 0 {
		return fmt.Errorf("at least one --option is required")
	}
	seen := make(map[string]struct{}, len(options))
	for _, option := range options {
		if _, duplicate := seen[option.ID]; duplicate {
			return fmt.Errorf("duplicate --option ID %q", option.ID)
		}
		seen[option.ID] = struct{}{}
	}
	if min == 0 && max == 0 {
		return nil
	}
	if min < 0 || max < min || max > len(options) {
		return fmt.Errorf("invalid --min-selections/--max-selections bounds")
	}
	return nil
}

func printDecideJSON(value any) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
}

type decideTimingSummary struct {
	Samples          int     `json:"samples"`
	PromptHash       string  `json:"prompt_sha256"`
	PromptVersion    string  `json:"prompt_version"`
	SDKTotalMS       float64 `json:"sdk_total_ms"`
	EndToEndMS       float64 `json:"end_to_end_ms"`
	HostOverheadMS   float64 `json:"host_overhead_ms"`
	TTFTMS           float64 `json:"ttft_ms"`
	PromptTokens     int64   `json:"prompt_tokens"`
	GeneratedTokens  int64   `json:"generated_tokens"`
	PrefillTokensSec float64 `json:"prefill_tokens_per_second"`
	DecodeTokensSec  float64 `json:"decode_tokens_per_second"`
	EndToEndStdDevMS float64 `json:"end_to_end_stddev_ms"`
	Threads          int32   `json:"threads"`
	ThreadsBatch     int32   `json:"threads_batch"`
	Batch            int32   `json:"batch"`
	UBatch           int32   `json:"ubatch"`
}

func runDecideBenchmark(repetitions int, reset func() error, run func() (any, decision.DecisionTiming, error)) (any, []decision.DecisionTiming, error) {
	if repetitions == 0 {
		result, timing, err := run()
		return result, []decision.DecisionTiming{timing}, err
	}
	if _, _, err := run(); err != nil {
		return nil, nil, fmt.Errorf("decision benchmark warmup: %w", err)
	}
	if err := reset(); err != nil {
		return nil, nil, fmt.Errorf("reset model after decision benchmark warmup: %w", err)
	}

	timings := make([]decision.DecisionTiming, 0, repetitions)
	var result any
	for attempt := 0; attempt < repetitions; attempt++ {
		var timing decision.DecisionTiming
		var err error
		result, timing, err = run()
		if err != nil {
			return nil, nil, fmt.Errorf("decision benchmark repetition %d: %w", attempt+1, err)
		}
		timings = append(timings, timing)
		if attempt+1 < repetitions {
			if err := reset(); err != nil {
				return nil, nil, fmt.Errorf("reset model after decision benchmark repetition %d: %w", attempt+1, err)
			}
		}
	}
	return result, timings, nil
}

func printDecideBenchmark(repetitions int, options decideOptions, timings []decision.DecisionTiming) {
	if repetitions == 0 || len(timings) == 0 {
		return
	}
	endToEnd := make([]float64, len(timings))
	sdkTotal := make([]float64, len(timings))
	promptHash := timings[0].PromptHash
	promptVersion := timings[0].PromptVersion
	for _, timing := range timings[1:] {
		if timing.PromptHash != promptHash || timing.PromptVersion != promptVersion {
			fmt.Fprintln(os.Stderr, "decision timing provenance differs between repetitions")
			promptHash, promptVersion = "", ""
			break
		}
	}
	ttft := make([]float64, len(timings))
	prefill := make([]float64, len(timings))
	decode := make([]float64, len(timings))
	var promptTokens, generatedTokens float64
	for index, timing := range timings {
		profile := timing.ProfileData
		endToEnd[index] = float64(timing.TotalTime) / float64(time.Millisecond)
		sdkTotal[index] = float64(profile.TotalTimeUs()) / 1000
		ttft[index] = float64(profile.TTFT) / 1000
		prefill[index] = profile.PrefillSpeed
		decode[index] = profile.DecodingSpeed
		promptTokens += float64(profile.PromptTokens)
		generatedTokens += float64(profile.GeneratedTokens)
	}
	medianEndToEnd := medianDecideTiming(endToEnd)
	medianSDKTotal := medianDecideTiming(sdkTotal)
	mean := 0.0
	for _, value := range endToEnd {
		mean += value
	}
	mean /= float64(len(endToEnd))
	variance := 0.0
	for _, value := range endToEnd {
		variance += (value - mean) * (value - mean)
	}
	summary := decideTimingSummary{
		Samples:          len(timings),
		PromptHash:       promptHash,
		PromptVersion:    promptVersion,
		SDKTotalMS:       medianSDKTotal,
		EndToEndMS:       medianEndToEnd,
		HostOverheadMS:   medianEndToEnd - medianSDKTotal,
		TTFTMS:           medianDecideTiming(ttft),
		PromptTokens:     int64(promptTokens / float64(len(timings))),
		GeneratedTokens:  int64(generatedTokens / float64(len(timings))),
		PrefillTokensSec: medianDecideTiming(prefill),
		DecodeTokensSec:  medianDecideTiming(decode),
		EndToEndStdDevMS: math.Sqrt(variance / float64(len(endToEnd))),
		Threads:          options.threads,
		ThreadsBatch:     options.threadsBatch,
		Batch:            options.batch,
		UBatch:           options.ubatch,
	}
	encoded, err := json.Marshal(summary)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decision timing encoding failed: %v\n", err)
		return
	}
	fmt.Fprintf(os.Stderr, "decision timing (one warmup, %d measured): %s\n", repetitions, encoded)
}

func medianDecideTiming(values []float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	middle := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[middle-1] + sorted[middle]) / 2
	}
	return sorted[middle]
}

func resolveDecideModelOptions(ctx context.Context, runtimeID string, options decideOptions) (geniex_sdk.PowerMode, string, error) {
	power, err := geniex_sdk.ResolvePowerMode(options.powerMode)
	if err != nil {
		return 0, "", err
	}
	if runtimeID != geniex_sdk.RuntimeLlamaCpp {
		if options.specType != "" || options.draftModel != "" || options.chatTemplatePath != "" || options.chatTemplateContent != "" || len(options.stop) != 0 {
			return 0, "", fmt.Errorf("speculative decoding, chat-template overrides, and stop sequences are only supported by llama_cpp")
		}
		if options.slidingWindowNKeep != 0 && !options.slidingWindow {
			return 0, "", fmt.Errorf("--sliding-window-n-keep requires --sliding-window")
		}
		return power, "", nil
	}
	if options.slidingWindow || options.slidingWindowNKeep != 0 {
		return 0, "", fmt.Errorf("sliding-window options are only supported by qairt")
	}
	if options.draftModel == "" {
		return power, "", nil
	}
	draft, err := resolveDraftModel(ctx, options.draftModel)
	if err != nil {
		return 0, "", err
	}
	return power, draft, nil
}

func withDecideModel(ctx context.Context, model string, options decideOptions, run func(*geniex_sdk.ModelPaths, *geniex_sdk.LLM, *geniex_sdk.VLM) error) error {
	name, precision := geniex_sdk.SplitNamePrecision(model)
	paths, err := ensureModelAvailable(ctx, name, precision)
	if err != nil {
		return err
	}
	if len(options.images) != 0 && paths.ModelType != geniex_sdk.ModelTypeVLM {
		return fmt.Errorf("--image requires a VLM; %s is %s", paths.ModelName, paths.ModelType)
	}
	if err := common.InitSDK(); err != nil {
		return err
	}
	defer geniex_sdk.DeInit()
	applyDecideComputeDefault()
	deviceID, nglResolved, nctxResolved, err := resolveModelParams(paths.RuntimeID, paths.ModelName)
	if err != nil {
		return err
	}
	powerMode, specDraftModel, err := resolveDecideModelOptions(ctx, paths.RuntimeID, options)
	if err != nil {
		return err
	}
	config := decideModelConfig(nctxResolved, nglResolved, options, powerMode, specDraftModel)
	spin := render.NewSpinner("loading decision model...")
	spin.Start()
	switch paths.ModelType {
	case geniex_sdk.ModelTypeLLM:
		llm, err := geniex_sdk.NewLLM(geniex_sdk.LlmCreateInput{ModelPath: paths.ModelPath, RuntimeID: paths.RuntimeID, DeviceID: deviceID, Config: config})
		spin.Stop()
		if err != nil {
			return err
		}
		defer llm.Destroy()
		return run(paths, llm, nil)
	case geniex_sdk.ModelTypeVLM:
		vlm, err := geniex_sdk.NewVLM(geniex_sdk.VlmCreateInput{ModelPath: paths.ModelPath, MmprojPath: paths.MmprojPath, RuntimeID: paths.RuntimeID, DeviceID: deviceID, VitDeviceID: vitComputeUnit, Config: config})
		spin.Stop()
		if err != nil {
			return err
		}
		defer vlm.Destroy()
		if len(options.images) != 0 {
			capabilities, err := vlm.Capabilities()
			if err != nil {
				return err
			}
			if !capabilities.SupportsVision && paths.RuntimeID == geniex_sdk.RuntimeLlamaCpp {
				return fmt.Errorf("model %s does not expose vision support", paths.ModelName)
			}
		}
		return run(paths, nil, vlm)
	default:
		return fmt.Errorf("unsupported model type: %s", paths.ModelType)
	}
}

func applyDecideComputeDefault() {
	// Decide does not use the inference command's package-level micro-batch
	// value. Keep its per-command option untouched while applying the same
	// compute defaulting policy as other local inference entry points.
	computeUnit, _ = config.ChipsetDefaults(computeUnit, 0, store.Get().ResolveChipset(true))
}
