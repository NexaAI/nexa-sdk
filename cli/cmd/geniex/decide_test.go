// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	decision "github.com/qualcomm/GenieX/cli/internal/decision"
)

func TestDecideCommandModes(t *testing.T) {
	command := decideCmd()
	for _, test := range []struct {
		name string
		args []string
		ok   bool
	}{
		{"classify", []string{"model", "classify"}, true},
		{"choose", []string{"model", "choose"}, true},
		{"triage removed", []string{"model", "triage"}, false},
		{"browser removed", []string{"model", "browser"}, false},
		{"missing mode", []string{"model"}, false},
		{"unknown mode", []string{"model", "answer"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := command.Args(command, test.args)
			if (err == nil) != test.ok {
				t.Fatalf("Args(%q) error = %v, want success=%v", test.args, err, test.ok)
			}
		})
	}
}

func TestRootRegistersDecideNotJev(t *testing.T) {
	var decide, jev bool
	for _, command := range RootCmd().Commands() {
		switch command.Name() {
		case "decide":
			decide = true
		case "jev":
			jev = true
		}
	}
	if !decide {
		t.Fatal("RootCmd does not register decide")
	}
	if jev {
		t.Fatal("RootCmd still registers obsolete command")
	}
}

func TestDecideContext(t *testing.T) {
	if _, err := decideContext(decideOptions{}); err == nil {
		t.Fatal("decideContext accepted missing context")
	}
	if _, err := decideContext(decideOptions{context: "a", contextFile: "b"}); err == nil {
		t.Fatal("decideContext accepted both context sources")
	}
	value, err := decideContext(decideOptions{context: "question"})
	if err != nil || value != "question" {
		t.Fatalf("decideContext() = (%q, %v)", value, err)
	}
}

func TestParseDecideOptions(t *testing.T) {
	options, err := parseDecideOptions([]string{"a=Earth", "b=Mars"})
	if err != nil || len(options) != 2 || options[1].ID != "b" || options[1].Text != "Mars" {
		t.Fatalf("parseDecideOptions() = (%#v, %v)", options, err)
	}
	for _, value := range [][]string{{"invalid"}, {"=empty"}, {"id="}} {
		if _, err := parseDecideOptions(value); err == nil {
			t.Fatalf("parseDecideOptions(%q) succeeded", value)
		}
	}
}

func TestDecideInputValidation(t *testing.T) {
	if err := validateDecideLabels([]string{"yes", "yes"}, "Classify."); err == nil {
		t.Fatal("validateDecideLabels accepted duplicate labels")
	}
	if err := validateDecideLabels([]string{"yes"}, ""); err == nil {
		t.Fatal("validateDecideLabels accepted missing instruction")
	}
	if err := validateDecideOptions([]decision.Option{{ID: "a", Text: "One"}, {ID: "a", Text: "Duplicate"}}, "Choose.", 0, 0); err == nil {
		t.Fatal("validateDecideOptions accepted duplicate IDs")
	}
	if err := validateDecideOptions([]decision.Option{{ID: "a", Text: "One"}}, "Choose.", 0, 2); err == nil {
		t.Fatal("validateDecideOptions accepted invalid bounds")
	}
}

func TestDecideOnlyAcceptsConsumedModelFlags(t *testing.T) {
	command := decideCmd()
	for _, name := range []string{"compute", "vit-compute", "ngl", "nctx", "max-tokens", "threads", "threads-batch", "batch", "ubatch", "n-seq-max", "chat-template-path", "chat-template-content", "spec-type", "draft-model", "draft-tokens", "draft-min", "draft-p-min", "power-mode", "stop", "sliding-window", "sliding-window-n-keep", "temperature", "top-p", "top-k", "min-p", "repetition-penalty", "presence-penalty", "frequency-penalty", "seed"} {
		if command.Flags().Lookup(name) == nil {
			t.Fatalf("decide is missing consumed flag --%s", name)
		}
	}
	for _, name := range []string{"image-max-length", "grammar-path", "grammar-string", "enable-json", "stop-file", "think", "system-prompt", "input", "prompt", "token-file", "message", "task", "url", "browser", "browser-path", "attach", "profile", "trace-dir", "headless", "dry-run", "max-steps", "auto"} {
		if command.Flags().Lookup(name) != nil {
			t.Fatalf("decide accepts removed or unsupported flag --%s", name)
		}
	}

	inferCommand := infer()
	for _, name := range []string{"temperature", "grammar-string", "think", "spec-type", "draft-model"} {
		if inferCommand.Flags().Lookup(name) == nil {
			t.Fatalf("infer lost flag --%s", name)
		}
	}
}

func TestDecideRuntimeOptions(t *testing.T) {
	if err := validateDecideRuntimeOptions(decideOptions{}); err != nil {
		t.Fatalf("validateDecideRuntimeOptions() error = %v", err)
	}
	for _, options := range []decideOptions{
		{threads: -1},
		{threadsBatch: -1},
		{batch: -1},
		{ubatch: -1},
	} {
		if err := validateDecideRuntimeOptions(options); err == nil {
			t.Fatalf("validateDecideRuntimeOptions(%#v) succeeded", options)
		}
	}
	config := decideModelConfig(4096, -1, decideOptions{threads: 4, threadsBatch: 6, batch: 512, ubatch: 256, nSeqMax: 2, specType: "ngram-simple"}, geniex_sdk.PowerModeBurst, "")
	if config.NCtx != 4096 || config.NGpuLayers != -1 || config.NThreads != 4 || config.NThreadsBatch != 6 || config.NBatch != 512 || config.NUbatch != 256 || config.NSeqMax != 2 || config.SpecType != "ngram-simple" {
		t.Fatalf("decideModelConfig() = %#v", config)
	}
}

func TestResolveDecideModelOptions(t *testing.T) {
	if _, _, err := resolveDecideModelOptions(context.Background(), geniex_sdk.RuntimeLlamaCpp, decideOptions{slidingWindow: true}); err == nil {
		t.Fatal("resolveDecideModelOptions accepted llama.cpp sliding window")
	}
	if _, _, err := resolveDecideModelOptions(context.Background(), geniex_sdk.RuntimeQairt, decideOptions{specType: "ngram-simple"}); err == nil {
		t.Fatal("resolveDecideModelOptions accepted QAIRT speculative decoding")
	}
	if _, _, err := resolveDecideModelOptions(context.Background(), geniex_sdk.RuntimeQairt, decideOptions{slidingWindowNKeep: 4}); err == nil {
		t.Fatal("resolveDecideModelOptions accepted n-keep without sliding window")
	}
	options := decideOptions{temperature: 0.7, topK: 4, stop: []string{"</s>"}, slidingWindow: true, slidingWindowNKeep: 8}
	generation := decideGenerationOptions(options)
	if generation.Temperature != options.temperature || generation.TopK != options.topK || generation.Stop[0] != "</s>" || !generation.SlidingWindow || generation.SlidingWindowNKeep != 8 {
		t.Fatalf("decideGenerationOptions() = %#v", generation)
	}
}

func TestDecideJSONShape(t *testing.T) {
	encoded, err := json.Marshal(decision.ClassificationResult{Label: "capital_of_france"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(encoded) != `{"label":"capital_of_france"}` {
		t.Fatalf("classification JSON = %s", encoded)
	}
	encoded, err = json.Marshal(decision.MultipleChoiceResult{SelectedIDs: []string{"paris"}})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(encoded) != `{"selected_ids":["paris"]}` {
		t.Fatalf("multiple-choice JSON = %s", encoded)
	}
}

func TestRunDecideBenchmark(t *testing.T) {
	calls := 0
	resets := 0
	result, timings, err := runDecideBenchmark(2, func() error {
		resets++
		return nil
	}, func() (any, decision.DecisionTiming, error) {
		calls++
		return calls, decision.DecisionTiming{TotalTime: time.Millisecond}, nil
	})
	if err != nil || result != 3 || calls != 3 || resets != 2 || len(timings) != 2 {
		t.Fatalf("runDecideBenchmark() = (%#v, %#v, %v), calls=%d resets=%d", result, timings, err, calls, resets)
	}
}

func TestDecideBenchmarkValidation(t *testing.T) {
	command := decideCmd()
	command.SetArgs([]string{"model", "classify", "--benchmark", "-1"})
	if err := command.Execute(); err == nil {
		t.Fatal("negative --benchmark succeeded")
	}
	command = decideCmd()
	command.SetArgs([]string{"model", "classify", "--threads", "-1"})
	if err := command.Execute(); err == nil {
		t.Fatal("negative --threads succeeded")
	}
}
