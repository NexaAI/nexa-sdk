// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/cmd/geniex/common"
	"github.com/qualcomm/GenieX/cli/internal/config"
	"github.com/qualcomm/GenieX/cli/internal/decision"
	"github.com/qualcomm/GenieX/cli/internal/render"
	"github.com/qualcomm/GenieX/cli/internal/store"
)

var (
	decisionKind     string
	decisionJSON     bool
	decisionInput    string
	decisionQuestion string
	decisionChoices  []string
	decisionLevels   []string
	decisionImages   []string
	decisionAudios   []string
)

func decisionFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("LLM/VLM Decision", pflag.ExitOnError)
	flags.SortFlags = false
	flags.StringVar(&decisionKind, "kind", "", "decision kind: predicate, choice, or score")
	flags.BoolVar(&decisionJSON, "json", false, "output the full decision result as JSON")
	flags.StringVarP(&decisionInput, "input", "i", "", "input text to evaluate")
	flags.StringVarP(&decisionQuestion, "question", "q", "", "question or instruction for the decision")
	flags.StringArrayVar(&decisionChoices, "choice", nil, "candidate VALUE; repeat for choice; VALUE=DESCRIPTION adds a description for VALUE")
	flags.StringArrayVar(&decisionLevels, "level", nil, "level LABEL; repeat for score; LABEL=DESCRIPTION adds a description for LABEL")
	return flags
}

func decisionVLMFlagSet() *pflag.FlagSet {
	flags := pflag.NewFlagSet("VLM Specific", pflag.ExitOnError)
	flags.SortFlags = false
	flags.StringVar(&vitComputeUnit, "vit-compute", "", "compute unit for the VLM vision encoder, e.g. CPU or HTP2 (llama_cpp only)")
	flags.StringArrayVar(&decisionImages, "image", nil, "image path, repeated for visual decisions")
	flags.StringArrayVar(&decisionAudios, "audio", nil, "audio path, repeated for audio decisions")
	return flags
}

func decide() *cobra.Command {
	cmd := &cobra.Command{
		GroupID: "inference",
		Use:     "decide <model-name>[:<precision>]",
		Short:   "Make a local structured decision",
		Long:    "Score a predicate, choice, or ordinal decision locally with a llama_cpp LLM or VLM.",
		Args:    cobra.ExactArgs(1),
	}
	groups := []*pflag.FlagSet{modelLoadFlags, decisionFlagSet(), decisionVLMFlagSet()}
	for _, flags := range groups {
		cmd.Flags().AddFlagSet(flags)
	}
	cmd.SetUsageFunc(func(c *cobra.Command) error { return groupedFlagUsage(c, groups) })
	cmd.MarkFlagRequired("kind")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		question, err := validateDecision(
			decision.Kind(decisionKind), decisionInput, decisionQuestion, decisionChoices, decisionLevels)
		if err != nil {
			return err
		}

		name, precision := geniex_sdk.SplitNamePrecision(args[0])
		if precision == "" {
			chosen, err := pickCachedPrecision(name)
			if err != nil {
				return err
			}
			precision = chosen
		}
		paths, err := ensureModelAvailable(cmd.Context(), name, precision)
		if err != nil {
			return err
		}
		if paths.RuntimeID != geniex_sdk.RuntimeLlamaCpp {
			return errors.New("decide requires a llama_cpp model")
		}
		images, err := decisionMediaPaths("--image", decisionImages)
		if err != nil {
			return err
		}
		audios, err := decisionMediaPaths("--audio", decisionAudios)
		if err != nil {
			return err
		}
		if (len(images) > 0 || len(audios) > 0) && paths.ModelType != geniex_sdk.ModelTypeVLM {
			return errors.New("--image and --audio require a VLM model")
		}
		if err := common.InitSDK(); err != nil {
			return err
		}
		defer geniex_sdk.DeInit()

		resolvedPower, err := geniex_sdk.ResolvePowerMode(powerMode)
		if err != nil {
			return err
		}
		computeUnit, ubatch = config.ChipsetDefaults(computeUnit, ubatch, store.Get().ResolveChipset(true))
		deviceID, resolvedNgl, resolvedNCtx, err := resolveModelParams(paths.RuntimeID, paths.ModelName)
		if err != nil {
			return err
		}
		if paths.ModelType == geniex_sdk.ModelTypeVLM {
			model, err := geniex_sdk.NewVLM(geniex_sdk.VlmCreateInput{
				ModelPath: paths.ModelPath, MmprojPath: paths.MmprojPath,
				RuntimeID: paths.RuntimeID, DeviceID: deviceID, VitDeviceID: vitComputeUnit,
				Config: geniex_sdk.ModelConfig{
					NCtx: resolvedNCtx, NUbatch: ubatch, NGpuLayers: resolvedNgl, PowerMode: resolvedPower,
				},
			})
			if err != nil {
				return err
			}
			defer model.Destroy()
			if len(images) > 0 || len(audios) > 0 {
				caps, err := model.Capabilities()
				if err != nil {
					return err
				}
				if len(images) > 0 && !caps.SupportsVision {
					return errors.New("--image requires a VLM with vision support")
				}
				if len(audios) > 0 && !caps.SupportsAudio {
					return errors.New("--audio requires a VLM with audio support")
				}
			}
			result, err := evaluateDecision(model, decisionInput, question, images, audios)
			if err != nil {
				return err
			}
			return writeDecisionResult(cmd.OutOrStdout(), result, decisionJSON)
		}
		if paths.ModelType != geniex_sdk.ModelTypeLLM {
			return fmt.Errorf("decide does not support model type %q", paths.ModelType)
		}

		model, err := geniex_sdk.NewLLM(geniex_sdk.LlmCreateInput{
			ModelPath: paths.ModelPath, RuntimeID: paths.RuntimeID, DeviceID: deviceID,
			Config: geniex_sdk.ModelConfig{
				NCtx: resolvedNCtx, NUbatch: ubatch, NGpuLayers: resolvedNgl, PowerMode: resolvedPower,
			},
		})
		if err != nil {
			return err
		}
		defer model.Destroy()
		result, err := evaluateDecision(model, decisionInput, question, nil, nil)
		if err != nil {
			return err
		}
		return writeDecisionResult(cmd.OutOrStdout(), result, decisionJSON)
	}
	return cmd
}

func decisionMediaPaths(flag string, raw []string) ([]string, error) {
	if len(raw) > 8 {
		return nil, fmt.Errorf("%s supports at most 8 files", flag)
	}
	paths := make([]string, 0, len(raw))
	for _, path := range raw {
		path = strings.TrimSpace(path)
		if path == "" {
			return nil, fmt.Errorf("%s paths must not be empty", flag)
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolve %s %q: %w", flag, path, err)
		}
		info, err := os.Stat(absolute)
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", flag, path, err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("%s %q is not a regular file", flag, path)
		}
		paths = append(paths, absolute)
	}
	return paths, nil
}

func validateDecision(kind decision.Kind, input, instructions string, rawChoices, rawLevels []string) (decision.Question, error) {
	switch kind {
	case decision.KindPredicate, decision.KindChoice, decision.KindScore:
	default:
		return decision.Question{}, errors.New("--kind must be predicate, choice, or score")
	}
	if strings.TrimSpace(input) == "" {
		return decision.Question{}, errors.New("--input must not be empty")
	}
	if strings.TrimSpace(instructions) == "" {
		return decision.Question{}, errors.New("--question must not be empty")
	}

	question := decision.Question{Kind: kind, Instructions: instructions}
	switch kind {
	case decision.KindPredicate:
		if len(rawChoices) != 0 || len(rawLevels) != 0 {
			return decision.Question{}, errors.New("predicate does not accept --choice or --level")
		}
		question.Choices = []decision.Choice{{Value: true}, {Value: false}}
	case decision.KindChoice:
		if len(rawLevels) != 0 {
			return decision.Question{}, errors.New("--level is only valid with --kind score")
		}
		choices, err := parseDecisionChoices(rawChoices)
		if err != nil {
			return decision.Question{}, err
		}
		if len(choices) < 2 || len(choices) > 255 {
			return decision.Question{}, errors.New("choice requires 2–255 --choice values")
		}
		question.Choices = choices
	case decision.KindScore:
		if len(rawChoices) != 0 {
			return decision.Question{}, errors.New("--choice is only valid with --kind choice")
		}
		levels, err := parseDecisionLevels(rawLevels)
		if err != nil {
			return decision.Question{}, err
		}
		if len(levels) < 2 || len(levels) > 10 {
			return decision.Question{}, errors.New("score requires 2–10 --level values")
		}
		question.Levels = levels
	}
	return question, nil
}

func parseDecisionChoices(raw []string) ([]decision.Choice, error) {
	choices := make([]decision.Choice, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		parts := strings.SplitN(value, "=", 2)
		choice := strings.TrimSpace(parts[0])
		if choice == "" {
			return nil, errors.New("--choice values must not be empty")
		}
		description := ""
		if len(parts) == 2 {
			description = strings.TrimSpace(parts[1])
			if description == "" {
				return nil, fmt.Errorf("description for choice %q must not be empty", choice)
			}
		}
		if _, exists := seen[choice]; exists {
			return nil, fmt.Errorf("duplicate choice %q", choice)
		}
		seen[choice] = struct{}{}
		choices = append(choices, decision.Choice{Value: choice, Description: description})
	}
	return choices, nil
}

func parseDecisionLevels(raw []string) ([]decision.Level, error) {
	levels := make([]decision.Level, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, value := range raw {
		parts := strings.SplitN(value, "=", 2)
		label := strings.TrimSpace(parts[0])
		if label == "" {
			return nil, errors.New("--level labels must not be empty")
		}
		description := label
		if len(parts) == 2 {
			description = strings.TrimSpace(parts[1])
			if description == "" {
				return nil, fmt.Errorf("description for level %q must not be empty", label)
			}
		}
		if _, exists := seen[label]; exists {
			return nil, fmt.Errorf("duplicate level %q", label)
		}
		seen[label] = struct{}{}
		levels = append(levels, decision.Level{Label: label, Description: description})
	}
	return levels, nil
}

type decisionOutput struct {
	Type          string                 `json:"type"`
	Probability   *float64               `json:"probability,omitempty"`
	Choice        any                    `json:"choice,omitempty"`
	Probabilities []decision.Probability `json:"probabilities,omitempty"`
	Confidence    *float64               `json:"confidence,omitempty"`
	Score         *float64               `json:"score,omitempty"`
}

func writeDecisionResult(w io.Writer, result decisionOutput, asJSON bool) error {
	if asJSON {
		return json.NewEncoder(w).Encode(result)
	}
	theme := render.GetTheme()
	var text strings.Builder
	switch result.Type {
	case string(decision.KindPredicate):
		if result.Probability == nil {
			return errors.New("predicate result has no probability")
		}
		trueStyle, falseStyle := theme.Info, theme.Info
		if *result.Probability >= 0.5 {
			trueStyle = theme.Success
		} else {
			falseStyle = theme.Success
		}
		fmt.Fprintf(&text, "%s %s\n", theme.Heading.Sprint("Probability (true):"), trueStyle.Sprintf("%.1f%%", *result.Probability*100))
		fmt.Fprintf(&text, "%s %s\n", theme.Heading.Sprint("Probability (false):"), falseStyle.Sprintf("%.1f%%", (1-*result.Probability)*100))
	case string(decision.KindChoice):
		fmt.Fprintf(&text, "%s %s\n", theme.Heading.Sprint("Choice:"), theme.Success.Sprint(result.Choice))
	case string(decision.KindScore):
		if result.Score == nil {
			return errors.New("score result has no score")
		}
		fmt.Fprintf(&text, "%s %s\n", theme.Heading.Sprint("Score:"), theme.Success.Sprintf("%.2f (range 0-%d)", *result.Score, len(result.Probabilities)-1))
	default:
		return fmt.Errorf("unknown decision kind %q", result.Type)
	}
	if result.Confidence != nil {
		fmt.Fprintf(&text, "%s %s\n", theme.Heading.Sprint("Confidence:"), theme.Success.Sprintf("%.1f%%", *result.Confidence*100))
	}
	switch result.Type {
	case string(decision.KindChoice):
		text.WriteString(theme.Heading.Sprint("Probabilities:"))
		text.WriteByte('\n')
		for _, option := range result.Probabilities {
			label := theme.Normal.Sprint(option.Value)
			probability := theme.Info.Sprintf("%.1f%%", option.Probability*100)
			if option.Value == result.Choice {
				label = theme.Success.Sprint(option.Value)
				probability = theme.Success.Sprintf("%.1f%%", option.Probability*100)
			}
			fmt.Fprintf(&text, "  %s: %s\n", label, probability)
		}
	case string(decision.KindScore):
		text.WriteString(theme.Heading.Sprint("Levels:"))
		text.WriteByte('\n')
		for _, level := range result.Probabilities {
			label := fmt.Sprint(level.Value)
			if level.Label != nil {
				label += " (" + *level.Label + ")"
			}
			fmt.Fprintf(&text, "  %s: %s\n", label, theme.Info.Sprintf("%.1f%%", level.Probability*100))
		}
	}
	_, err := io.WriteString(w, text.String())
	return err
}

type decisionModel interface {
	Decide(geniex_sdk.DecideInput) (geniex_sdk.DecideOutput, error)
}

func evaluateDecision(model decisionModel, input string, question decision.Question, images, audios []string) (decisionOutput, error) {
	questions, err := decision.Questions([]decision.Question{question})
	if err != nil {
		return decisionOutput{}, err
	}
	result, err := model.Decide(geniex_sdk.DecideInput{
		State: input, Questions: questions, QuestionIndex: 0,
		ImagePaths: images, AudioPaths: audios,
	})
	if err != nil {
		return decisionOutput{}, err
	}
	// LEV has already calibrated multi-pass probabilities in the SDK.
	var answer decision.Result
	if len(result.Probabilities) > 0 {
		answer, err = decision.AnswerProbabilities(question, result.Probabilities)
	} else {
		answer, err = decision.Answer(question, result.Logits, float64(result.Temperature))
	}
	if err != nil {
		return decisionOutput{}, err
	}
	return decisionOutput{
		Type:          string(question.Kind),
		Probability:   answer.Probability,
		Choice:        answer.Choice,
		Probabilities: answer.Probabilities,
		Confidence:    answer.Confidence,
		Score:         answer.Score,
	}, nil
}
