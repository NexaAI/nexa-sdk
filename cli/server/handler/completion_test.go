// Copyright (c) 2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"math"
	"reflect"
	"testing"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
)

func TestCompletionPrompt(t *testing.T) {
	tests := []struct {
		name    string
		prompt  openai.CompletionNewParamsPromptUnion
		want    string
		wantErr bool
	}{
		{"string", openai.CompletionNewParamsPromptUnion{OfString: param.NewOpt("<|fim_prefix|>a<|fim_suffix|>b<|fim_middle|>")}, "<|fim_prefix|>a<|fim_suffix|>b<|fim_middle|>", false},
		{"empty string", openai.CompletionNewParamsPromptUnion{OfString: param.NewOpt("")}, "", true},
		{"single-element array", openai.CompletionNewParamsPromptUnion{OfArrayOfStrings: []string{"abc"}}, "abc", false},
		{"multi-element array", openai.CompletionNewParamsPromptUnion{OfArrayOfStrings: []string{"a", "b"}}, "", true},
		{"empty array", openai.CompletionNewParamsPromptUnion{OfArrayOfStrings: []string{}}, "", true},
		{"token array", openai.CompletionNewParamsPromptUnion{OfArrayOfTokens: []int64{1, 2}}, "", true},
		{"omitted", openai.CompletionNewParamsPromptUnion{}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := completionPrompt(tt.prompt)
			if (err != nil) != tt.wantErr {
				t.Fatalf("completionPrompt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("completionPrompt() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCompletionStop(t *testing.T) {
	tests := []struct {
		name string
		stop openai.CompletionNewParamsStopUnion
		want []string
	}{
		{"omitted", openai.CompletionNewParamsStopUnion{}, nil},
		{"string", openai.CompletionNewParamsStopUnion{OfString: param.NewOpt("\n\n")}, []string{"\n\n"}},
		{"empty string", openai.CompletionNewParamsStopUnion{OfString: param.NewOpt("")}, nil},
		{"array", openai.CompletionNewParamsStopUnion{OfStringArray: []string{"<|endoftext|>", "\n"}}, []string{"<|endoftext|>", "\n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := completionStop(tt.stop); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("completionStop() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestCompletionUnsupported(t *testing.T) {
	tests := []struct {
		name    string
		params  CompletionNewParams
		wantErr bool
	}{
		{"defaults", CompletionNewParams{}, false},
		{"n 1", CompletionNewParams{N: param.NewOpt[int64](1)}, false},
		{"best_of 1", CompletionNewParams{BestOf: param.NewOpt[int64](1)}, false},
		{"empty suffix", CompletionNewParams{Suffix: param.NewOpt("")}, false},
		{"suffix", CompletionNewParams{Suffix: param.NewOpt("tail")}, true},
		{"n 2", CompletionNewParams{N: param.NewOpt[int64](2)}, true},
		{"best_of 2", CompletionNewParams{BestOf: param.NewOpt[int64](2)}, true},
		{"logprobs", CompletionNewParams{Logprobs: param.NewOpt[int64](0)}, true},
		{"logit_bias", CompletionNewParams{LogitBias: map[string]int64{"50256": -100}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := completionUnsupported(tt.params); (err != nil) != tt.wantErr {
				t.Errorf("completionUnsupported() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCompletionRequestFieldsAfterSDKUpgrade(t *testing.T) {
	req := defaultCompletionRequest()
	body := `{
		"model":"local/example:Q4_K_M", "prompt":["Hello"], "max_tokens":12,
		"suffix":"", "echo":true, "stop":["STOP","END"],
		"temperature":0.3, "top_p":0.9, "seed":31,
		"presence_penalty":0.4, "frequency_penalty":0.5,
		"stream":true, "stream_options":{"include_usage":true},
		"top_k":7, "min_p":0.1, "repetition_penalty":1.2,
		"nctx":2048, "ngl":-1, "compute":"hybrid", "vit_compute":"CPU", "power_mode":"burst"
	}`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	prompt, err := completionPrompt(req.Prompt)
	if err != nil || prompt != "Hello" || !reflect.DeepEqual(completionStop(req.Stop), []string{"STOP", "END"}) || completionUnsupported(req.CompletionNewParams) != nil {
		t.Fatalf("prompt/stop/support fields changed: %+v, %v", req, err)
	}
	if req.Model != "local/example:Q4_K_M" || req.MaxTokens.Value != 12 || req.Suffix.Value != "" || !req.Echo.Value || !req.Stream || !req.StreamOptions.IncludeUsage.Value {
		t.Fatalf("core completion fields changed: %+v", req)
	}
	if req.Temperature.Value != 0.3 || req.TopP.Value != 0.9 || req.Seed.Value != 31 || req.PresencePenalty.Value != 0.4 || req.FrequencyPenalty.Value != 0.5 || req.TopK != 7 || req.MinP != 0.1 || req.RepetitionPenalty != 1.2 {
		t.Fatalf("sampler fields changed: %+v", req)
	}
	if req.NCtx != 2048 || req.Ngl != -1 || req.Compute != "hybrid" || req.VitCompute != "CPU" || req.PowerMode != "burst" {
		t.Fatalf("device fields changed: %+v", req)
	}
}

func embedsB64(vals ...float32) string {
	raw := make([]byte, 4*len(vals))
	for i, v := range vals {
		binary.LittleEndian.PutUint32(raw[i*4:], math.Float32bits(v))
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func TestCompletionPromptEmbeds(t *testing.T) {
	withPrompt := openai.CompletionNewParamsPromptUnion{OfString: param.NewOpt("hi")}
	tests := []struct {
		name    string
		mutate  func(*CompletionRequest)
		want    []float32
		wantErr bool
	}{
		{"unset", func(r *CompletionRequest) {}, nil, false},
		{"two rows", func(r *CompletionRequest) { r.PromptEmbeds, r.PromptEmbedsDim = embedsB64(1, 2, 3, 4), 2 }, []float32{1, 2, 3, 4}, false},
		{"prompt conflict", func(r *CompletionRequest) {
			r.Prompt, r.PromptEmbeds, r.PromptEmbedsDim = withPrompt, embedsB64(1, 2), 2
		}, nil, true},
		{"echo", func(r *CompletionRequest) {
			r.Echo, r.PromptEmbeds, r.PromptEmbedsDim = param.NewOpt(true), embedsB64(1, 2), 2
		}, nil, true},
		{"missing dim", func(r *CompletionRequest) { r.PromptEmbeds = embedsB64(1, 2) }, nil, true},
		{"bad base64", func(r *CompletionRequest) { r.PromptEmbeds, r.PromptEmbedsDim = "!!", 2 }, nil, true},
		{"not float32 sized", func(r *CompletionRequest) {
			r.PromptEmbeds, r.PromptEmbedsDim = base64.StdEncoding.EncodeToString([]byte{1, 2, 3}), 1
		}, nil, true},
		{"not a multiple of dim", func(r *CompletionRequest) { r.PromptEmbeds, r.PromptEmbedsDim = embedsB64(1, 2, 3), 2 }, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := defaultCompletionRequest()
			tt.mutate(&req)
			got, err := completionPromptEmbeds(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("completionPromptEmbeds() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("completionPromptEmbeds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompletionRequestPromptEmbedsFields(t *testing.T) {
	req := defaultCompletionRequest()
	body := `{"model":"m", "prompt_embeds":"` + embedsB64(1, 2) + `", "prompt_embeds_dim":2, "max_tokens":4}`
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatal(err)
	}
	got, err := completionPromptEmbeds(req)
	if err != nil || !reflect.DeepEqual(got, []float32{1, 2}) || req.PromptEmbedsDim != 2 || req.MaxTokens.Value != 4 {
		t.Fatalf("prompt_embeds fields not bound: %+v, %v, %v", req, got, err)
	}
}
