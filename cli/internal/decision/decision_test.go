// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package decision

import (
	"math"
	"testing"
)

func TestQuestionsPreserveSystemoneVariables(t *testing.T) {
	questions := []Question{
		{
			ID: "route", Kind: KindChoice, Instructions: "Choose a route",
			Choices: []Choice{{Value: "retry", Description: "try again"}, {Value: "replace", Description: "new battery"}},
		},
		{Kind: KindPredicate, Instructions: "Is it safe?", Choices: []Choice{{Value: true}, {Value: false}}},
		{Kind: KindScore, Instructions: "Rate it", Levels: []Level{{Label: "low", Description: "bad"}, {Label: "high", Description: "good"}}},
	}
	got, err := Questions(questions)
	if err != nil {
		t.Fatal(err)
	}
	if got[0].ID != "route" || got[0].Type != "choice" || got[0].Instructions != "Choose a route" ||
		got[0].Options[0].Key != "retry" || *got[0].Options[1].Description != "new battery" {
		t.Fatalf("choice = %+v", got[0])
	}
	if got[1].ID != "" || got[1].Type != "predicate" || got[1].Options[0].Key != "true" ||
		got[1].Options[0].KeyIsBool || got[1].Options[1].Key != "false" {
		t.Fatalf("predicate = %+v", got[1])
	}
	if got[2].Type != "score" || got[2].Options[1].Key != "1" || got[2].Options[1].Description == nil {
		t.Fatalf("score = %+v", got[2])
	}
}

func TestAnswerComputesPredicateAndChoiceOutputs(t *testing.T) {
	predicateTemperature := 0.85 * 1.829074
	predicate, err := Answer(
		Question{Kind: KindPredicate, Choices: []Choice{{Value: false}, {Value: true}}},
		[]float32{0, float32(math.Log(3))}, predicateTemperature)
	wantPredicate := math.Exp(math.Log(3)/predicateTemperature) / (1 + math.Exp(math.Log(3)/predicateTemperature))
	if err != nil || predicate.Probability == nil || math.Abs(*predicate.Probability-wantPredicate) > 1e-6 {
		t.Fatalf("predicate = %+v, want %v, err = %v", predicate, wantPredicate, err)
	}
	choice, err := Answer(Question{Kind: KindChoice, Choices: []Choice{{Value: "a"}, {Value: "b"}}}, []float32{0, 2}, 0.85)
	wantChoiceConfidence := (1/(1+math.Exp(-2/0.85)) - 0.5) / 0.5
	if err != nil || choice.Choice != "b" || choice.Confidence == nil || math.Abs(*choice.Confidence-wantChoiceConfidence) > 1e-6 || len(choice.Probabilities) != 2 {
		t.Fatalf("choice = %+v, want confidence %v, err = %v", choice, wantChoiceConfidence, err)
	}
}

func TestAnswerComputesScoreOutput(t *testing.T) {
	question := Question{Kind: KindScore, Levels: []Level{{Label: "low", Description: "low"}, {Label: "medium", Description: "medium"}, {Label: "high", Description: "high"}}}
	answer, err := Answer(question, []float32{0, 0, 0}, 1)
	if err != nil || answer.Score == nil || *answer.Score != 1 || answer.Confidence == nil || *answer.Confidence != 0 {
		t.Fatalf("uniform score = %+v, err = %v", answer, err)
	}
	answer, err = Answer(question, []float32{-1000, 1000, -1000}, 1)
	if err != nil || answer.Score == nil || *answer.Score != 1 || answer.Confidence == nil || *answer.Confidence != 1 {
		t.Fatalf("certain score = %+v, err = %v", answer, err)
	}
}

func TestAnswerProbabilities(t *testing.T) {
	predicate := Question{Kind: KindPredicate, Choices: []Choice{{Value: false}, {Value: true}}}
	answer, err := AnswerProbabilities(predicate, []float32{0.25, 0.75})
	if err != nil || answer.Probability == nil || math.Abs(*answer.Probability-0.75) > 1e-6 {
		t.Fatalf("predicate = %+v, err = %v", answer, err)
	}
	choice := Question{Kind: KindChoice, Choices: []Choice{{Value: "first"}, {Value: "second"}}}
	answer, err = AnswerProbabilities(choice, []float32{0.1, 0.9})
	if err != nil || answer.Choice != "second" || len(answer.Probabilities) != 2 {
		t.Fatalf("choice = %+v, err = %v", answer, err)
	}
	for _, values := range [][]float32{{0}, {0, 0}, {-1, 2}, {float32(math.NaN()), 1}} {
		if _, err := AnswerProbabilities(choice, values); err == nil {
			t.Errorf("accepted probabilities %v", values)
		}
	}
}

func TestAnswerRejectsWrongAndNonFiniteValues(t *testing.T) {
	question := Question{Kind: KindChoice, Choices: []Choice{{Value: "a"}, {Value: "b"}}}
	for _, test := range []struct {
		logits      []float32
		temperature float64
	}{
		{[]float32{0}, 1},
		{[]float32{float32(math.NaN()), 0}, 1},
		{[]float32{float32(math.Inf(1)), 0}, 1},
		{[]float32{0, 1}, 0},
		{[]float32{0, 1}, math.NaN()},
	} {
		if _, err := Answer(question, test.logits, test.temperature); err == nil {
			t.Errorf("accepted logits %v temperature %v", test.logits, test.temperature)
		}
	}
}

func TestRejectsInvalidQuestions(t *testing.T) {
	if _, err := Questions([]Question{{Kind: KindPredicate, Instructions: "test", Choices: []Choice{{Value: true}, {Value: true}}}}); err == nil {
		t.Fatal("accepted duplicate predicate candidates")
	}
	question := Question{Kind: Kind("unknown")}
	if _, err := Questions([]Question{question}); err == nil {
		t.Fatal("accepted an unsupported kind")
	}
	if _, err := Answer(question, nil, 1); err == nil {
		t.Fatal("Answer accepted an unsupported kind")
	}
}
