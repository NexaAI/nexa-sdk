// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"strings"
	"testing"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/internal/decision"
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
