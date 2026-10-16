// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/internal/decision"
	"github.com/qualcomm/GenieX/cli/internal/render"
)

type fakeDecisionModel struct {
	input         geniex_sdk.DecideInput
	logits        []float32
	probabilities []float32
	temperature   float32
}

func (f *fakeDecisionModel) Decide(input geniex_sdk.DecideInput) (geniex_sdk.DecideOutput, error) {
	f.input = input
	return geniex_sdk.DecideOutput{Logits: f.logits, Probabilities: f.probabilities, Temperature: f.temperature}, nil
}

func TestDecideVisionComputeFlag(t *testing.T) {
	cmd := decide()
	if cmd.Flags().Lookup("vit-compute") == nil {
		t.Fatal("decide is missing --vit-compute")
	}
	if cmd.Flags().Lookup("max-tokens") != nil {
		t.Fatal("decide registered unrelated inference flags")
	}
	previous := vitComputeUnit
	t.Cleanup(func() { vitComputeUnit = previous })
	if err := cmd.Flags().Parse([]string{"--vit-compute", "CPU"}); err != nil || vitComputeUnit != "CPU" {
		t.Fatalf("--vit-compute parsed as %q, err = %v", vitComputeUnit, err)
	}
}

func TestDecideHelpFlagGroups(t *testing.T) {
	cmd := decide()
	var output bytes.Buffer
	cmd.SetOut(&output)
	if err := cmd.Usage(); err != nil {
		t.Fatal(err)
	}
	help := output.String()
	model := strings.Index(help, "Model Loading Flags:")
	decision := strings.Index(help, "LLM/VLM Decision Flags:")
	vlm := strings.Index(help, "VLM Specific Flags:")
	if model < 0 || decision <= model || vlm <= decision {
		t.Fatalf("missing or unordered decide flag groups: %s", help)
	}
	if !strings.Contains(help[model:decision], "--compute") ||
		!strings.Contains(help[decision:vlm], "--kind") ||
		!strings.Contains(help[decision:vlm], "--choice") ||
		!strings.Contains(help[decision:vlm], "--json") ||
		!strings.Contains(help[decision:vlm], "adds a description for VALUE") ||
		!strings.Contains(help[vlm:], "--vit-compute") ||
		!strings.Contains(help[vlm:], "--image") ||
		strings.Contains(help[vlm:], "requires a VLM") || strings.Contains(help, "--max-tokens") {
		t.Fatalf("decide flags are assigned to the wrong groups: %s", help)
	}
}

func TestParseDecisionChoicesWithDescription(t *testing.T) {
	choices, err := parseDecisionChoices([]string{"retry=Try again", "replace"})
	if err != nil || len(choices) != 2 || choices[0].Value != "retry" ||
		choices[0].Description != "Try again" || choices[1].Value != "replace" || choices[1].Description != "" {
		t.Fatalf("choice value/description = %+v, err = %v", choices, err)
	}
}

func TestWriteDecisionResult(t *testing.T) {
	probability, falseProbability, confidence, score := 0.857, 0.2, 0.963, 1.23
	low, mid, high := "low", "mid", "high"
	for _, tc := range []struct {
		name   string
		result decisionOutput
		want   string
	}{
		{"predicate", decisionOutput{Type: "predicate", Probability: &probability}, "Probability (true): 85.7%\nProbability (false): 14.3%\n"},
		{"predicate false", decisionOutput{Type: "predicate", Probability: &falseProbability}, "Probability (true): 20.0%\nProbability (false): 80.0%\n"},
		{"choice", decisionOutput{Type: "choice", Choice: "dog", Confidence: &confidence,
			Probabilities: []decision.Probability{{Value: "dog", Probability: 0.982}, {Value: "cat", Probability: 0.018}}},
			"Choice: dog\nConfidence: 96.3%\nProbabilities:\n  dog: 98.2%\n  cat: 1.8%\n"},
		{"score", decisionOutput{Type: "score", Score: &score, Confidence: &confidence,
			Probabilities: []decision.Probability{
				{Value: 0, Label: &low, Probability: 0.2},
				{Value: 1, Label: &mid, Probability: 0.37},
				{Value: 2, Label: &high, Probability: 0.43},
			}},
			"Score: 1.23 (range 0-2)\nConfidence: 96.3%\nLevels:\n  0 (low): 20.0%\n  1 (mid): 37.0%\n  2 (high): 43.0%\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			err := writeDecisionResult(&output, tc.result, false)
			plain := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(output.String(), "")
			if err != nil || plain != tc.want {
				t.Fatalf("text = %q, err = %v, want %q", plain, err, tc.want)
			}
		})
	}
}

func TestWriteDecisionChoiceColors(t *testing.T) {
	theme := render.GetTheme()
	if theme.Success.Code() == "" {
		t.Skip("terminal theme has no color")
	}
	confidence := 0.963
	result := decisionOutput{Type: "choice", Choice: "dog", Confidence: &confidence,
		Probabilities: []decision.Probability{{Value: "dog", Probability: 0.982}, {Value: "cat", Probability: 0.018}}}
	var output bytes.Buffer
	if err := writeDecisionResult(&output, result, false); err != nil {
		t.Fatal(err)
	}
	wantSelected := "  " + theme.Success.Sprint("dog") + ": " + theme.Success.Sprintf("98.2%%")
	wantUnselected := "  " + theme.Normal.Sprint("cat") + ": " + theme.Info.Sprintf("1.8%%")
	if !strings.Contains(output.String(), wantSelected) || !strings.Contains(output.String(), wantUnselected) ||
		!strings.Contains(output.String(), theme.Success.Sprintf("96.3%%")) {
		t.Fatalf("wrong decision colors: %q", output.String())
	}
}

func TestWriteDecisionResultJSON(t *testing.T) {
	previous := decisionJSON
	t.Cleanup(func() { decisionJSON = previous })
	if err := decide().Flags().Parse([]string{"--json"}); err != nil || !decisionJSON {
		t.Fatalf("--json parsing failed: enabled = %t, err = %v", decisionJSON, err)
	}
	confidence := 0.6
	result := decisionOutput{
		Type: "choice", Choice: "retry", Confidence: &confidence,
		Probabilities: []decision.Probability{{Value: "retry", Probability: 0.7}, {Value: "replace", Probability: 0.3}},
	}
	var output bytes.Buffer
	if err := writeDecisionResult(&output, result, decisionJSON); err != nil {
		t.Fatal(err)
	}
	var decoded decisionOutput
	if err := json.Unmarshal(output.Bytes(), &decoded); err != nil || decoded.Type != "choice" || decoded.Choice != "retry" ||
		len(decoded.Probabilities) != 2 || decoded.Probabilities[0].Probability != 0.7 ||
		decoded.Confidence == nil || *decoded.Confidence != confidence {
		t.Fatalf("JSON = %q, decoded = %+v, err = %v", output.String(), decoded, err)
	}
}

func TestValidateDecision(t *testing.T) {
	predicate, err := validateDecision(decision.KindPredicate, "text", "is this urgent?", nil, nil)
	if err != nil || len(predicate.Choices) != 2 {
		t.Fatalf("predicate validation = %v, question = %+v", err, predicate)
	}
	if _, err := validateDecision(decision.KindChoice, "text", "route it", []string{"one"}, nil); err == nil {
		t.Fatal("accepted one choice")
	}
	if _, err := validateDecision(decision.KindChoice, "text", "route it", []string{"one", "one"}, nil); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate choice error = %v", err)
	}
	score, err := validateDecision(decision.KindScore, "text", "rate it", nil, []string{"low", "high=urgent"})
	if err != nil || len(score.Levels) != 2 || score.Levels[1].Description != "urgent" {
		t.Fatalf("score validation = %v, question = %+v", err, score)
	}
	if _, err := validateDecision(decision.Kind("unknown"), "text", "question", nil, nil); err == nil {
		t.Fatal("accepted an unknown kind")
	}
}

func TestEvaluateDecisionPredicateReturnsProbability(t *testing.T) {
	model := &fakeDecisionModel{logits: []float32{0, 2}, temperature: 1}
	result, err := evaluateDecision(model, "battery failed", decision.Question{
		Kind: decision.KindPredicate, Instructions: "Is it urgent?",
		Choices: []decision.Choice{{Value: false, Description: "false"}, {Value: true, Description: "true"}},
	}, nil, nil)
	if err != nil || result.Type != "predicate" || result.Probability == nil || *result.Probability <= 0.5 {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	if model.input.State != "battery failed" || len(model.input.Questions) != 1 ||
		model.input.Questions[0].Type != "predicate" {
		t.Fatalf("decision input = %+v", model.input)
	}
}

func TestEvaluateDecisionUsesCalibratedProbabilities(t *testing.T) {
	model := &fakeDecisionModel{probabilities: []float32{0.25, 0.75}, temperature: 2}
	result, err := evaluateDecision(model, "battery failed", decision.Question{
		Kind: decision.KindPredicate, Instructions: "Is it urgent?",
		Choices: []decision.Choice{{Value: false}, {Value: true}},
	}, nil, nil)
	if err != nil || result.Probability == nil || *result.Probability < 0.749 {
		t.Fatalf("calibrated probability = %+v, err = %v", result, err)
	}
}

func TestEvaluateDecisionPassesState(t *testing.T) {
	model := &fakeDecisionModel{logits: []float32{0, 1}, temperature: 1}
	input := "Quoted: \"battery\" <__media__> 中文"
	_, err := evaluateDecision(model, input, decision.Question{
		Kind: decision.KindChoice, Instructions: "Choose",
		Choices: []decision.Choice{{Value: "a"}, {Value: "b"}},
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if model.input.State != input {
		t.Fatalf("state = %q, want %q", model.input.State, input)
	}
}

func TestDecisionMediaPathsValidateAndResolve(t *testing.T) {
	if _, err := decisionMediaPaths("--image", make([]string, 9)); err == nil {
		t.Fatal("accepted more than 8 image paths")
	}
	if _, err := decisionMediaPaths("--audio", []string{""}); err == nil {
		t.Fatal("accepted empty audio path")
	}
	if _, err := decisionMediaPaths("--image", []string{"does-not-exist.png"}); err == nil {
		t.Fatal("accepted missing image path")
	}
}

func TestEvaluateDecisionVLMUsesMediaLogits(t *testing.T) {
	model := &fakeDecisionModel{logits: []float32{0, 2}, temperature: 1}
	result, err := evaluateDecision(model, "media", decision.Question{
		Kind: decision.KindChoice, Instructions: "Choose", Choices: []decision.Choice{{Value: "cat"}, {Value: "dog"}},
	}, []string{"/tmp/photo.png"}, []string{"/tmp/audio.wav"})
	if err != nil || result.Choice != "dog" {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
	if len(model.input.ImagePaths) != 1 || len(model.input.AudioPaths) != 1 {
		t.Fatalf("score input = %+v", model.input)
	}
}

func TestEvaluateDecisionChoice(t *testing.T) {
	model := &fakeDecisionModel{logits: []float32{0, 2}, temperature: 0.85}
	result, err := evaluateDecision(model, "battery failed", decision.Question{
		Kind: decision.KindChoice, Instructions: "Choose a route",
		Choices: []decision.Choice{{Value: "retry", Description: "try again"}, {Value: "replace", Description: "new battery"}},
	}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.Type != "choice" || result.Choice != "replace" || result.Confidence == nil {
		t.Fatalf("result = %+v", result)
	}
	if len(model.input.Questions) != 1 || model.input.Questions[0].Instructions != "Choose a route" ||
		model.input.Questions[0].Options[0].Key != "retry" {
		t.Fatalf("questions = %+v", model.input.Questions)
	}
}

func TestEvaluateDecisionScore(t *testing.T) {
	model := &fakeDecisionModel{logits: []float32{0, 0}, temperature: 1}
	result, err := evaluateDecision(model, "battery failed", decision.Question{
		Kind: decision.KindScore, Instructions: "Rate severity",
		Levels: []decision.Level{{Label: "low", Description: "low"}, {Label: "high", Description: "high"}},
	}, nil, nil)
	if err != nil || result.Type != "score" || result.Score == nil || *result.Score != 0.5 || result.Confidence == nil {
		t.Fatalf("result = %+v, err = %v", result, err)
	}
}
