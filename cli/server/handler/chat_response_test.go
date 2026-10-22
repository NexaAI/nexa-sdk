// Copyright (c) 2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
)

// Covers both bodies writeBlockingResponse can produce: openai's own struct, and the
// local one that carries reasoning_content.
type blockingBody struct {
	Choices []struct {
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
	Timings timings `json:"timings"`
}

func TestWriteBlockingResponseGemma4EmptyThought(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var content, reasoning strings.Builder
	onToken, finish := sinkWithToolCalls(reasoningClass(), &content, &reasoning, "")
	for _, token := range []string{"<|channel>", "thought", "<channel|>", `<|tool_call>call:go{}<tool_call|>`} {
		onToken(token)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	writeBlockingResponse(c, &content, &reasoning, geniex_sdk.ProfileData{StopReason: "eos"}, finish)

	var got blockingBody
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	choice := got.Choices[0]
	if choice.Message.Content != "" || choice.Message.ReasoningContent != "" {
		t.Errorf("content = %q, reasoning = %q, want both empty", choice.Message.Content, choice.Message.ReasoningContent)
	}
	if choice.FinishReason != "tool_calls" || len(choice.Message.ToolCalls) != 1 ||
		choice.Message.ToolCalls[0].Function.Name != "go" || choice.Message.ToolCalls[0].Function.Arguments != `{}` {
		t.Errorf("choice = %+v, want one go tool call", choice)
	}
}

// A tool call must not cost the text around it, nor the reasoning: both used to be
// dropped whenever one was found.
func TestWriteBlockingResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const call = `{"name":"get_weather","arguments":{"city":"Beijing"}}`
	weather := [2]string{"get_weather", `{"city":"Beijing"}`}

	tests := []struct {
		name          string
		content       string
		reasoning     string
		parseTool     bool
		tools         string
		wantContent   string
		wantReasoning string
		wantFinish    string
		wantCalls     [][2]string
	}{
		{
			name: "text only", content: "hello",
			wantContent: "hello", wantFinish: "stop",
		},
		{
			name: "tools on but nothing to match", content: "hello", parseTool: true,
			wantContent: "hello", wantFinish: "stop",
		},
		{
			name: "a call keeps the prose around it", content: "sure " + call + " done", parseTool: true,
			wantContent: "sure  done", wantFinish: "tool_calls", wantCalls: [][2]string{weather},
		},
		{
			name: "parallel calls", content: call + call, parseTool: true,
			wantFinish: "tool_calls", wantCalls: [][2]string{weather, weather},
		},
		{
			name: "gemma4 syntax", content: `x <|tool_call>call:f{a:1}<tool_call|>`, parseTool: true,
			wantContent: "x ", wantFinish: "tool_calls", wantCalls: [][2]string{{"f", `{"a":1}`}},
		},
		{
			name: "MiniCPM5 typed arguments", content: `<function name="set"><param name="count">5</param><param name="label">5</param></function>`,
			parseTool: true, tools: `[{"type":"function","function":{"name":"set","parameters":{"type":"object","properties":{"count":{"type":"integer"},"label":{"type":"string"}}}}}]`,
			wantFinish: "tool_calls", wantCalls: [][2]string{{"set", `{"count":5,"label":"5"}`}},
		},
		{
			name: "MiniCPM5 unsupported schema leaves valid types intact", content: `<function name="set"><param name="count">5</param><param name="disabled">false</param></function>`,
			parseTool: true, tools: `[{"type":"function","function":{"name":"set","parameters":{"type":"object","properties":{"count":{"type":"integer"},"disabled":false}}}},
				{"type":"function","function":{"name":"other","parameters":{"type":"object","properties":{"disabled":false}}}}]`,
			wantFinish: "tool_calls", wantCalls: [][2]string{{"set", `{"count":5,"disabled":"false"}`}},
		},
		{
			name: "reasoning with a call", content: "sure " + call, reasoning: "let me check",
			parseTool: true, wantContent: "sure ", wantReasoning: "let me check",
			wantFinish: "tool_calls", wantCalls: [][2]string{weather},
		},
		{
			name: "reasoning without a call", content: "hello", reasoning: "let me check",
			parseTool: true, wantContent: "hello", wantReasoning: "let me check", wantFinish: "stop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tools := tt.tools
			if tools != "" {
				_, request := bindRequest(t, `{"model":"m","messages":[{"role":"user","content":"test"}],"tools":`+tools+`}`)
				_, encoded, err := parseTools(request)
				if err != nil {
					t.Fatal(err)
				}
				tools = encoded
			}
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			profile := geniex_sdk.ProfileData{StopReason: "eos", PromptTokens: 5, PrefillSpeed: 10, GeneratedTokens: 3, DecodingSpeed: 20}
			var content, reasoning strings.Builder
			onToken := sink(plainClass, &content, &reasoning)
			finish := func() []openai.ChatCompletionMessageFunctionToolCallFunction { return nil }
			if tt.parseTool {
				onToken, finish = sinkWithToolCalls(plainClass, &content, &reasoning, tools)
			}
			onToken(tt.content)
			reasoning.WriteString(tt.reasoning)
			writeBlockingResponse(c, &content, &reasoning, profile, finish)

			var got blockingBody
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("body %s: %v", w.Body.Bytes(), err)
			}
			if len(got.Choices) != 1 {
				t.Fatalf("choices = %d, want 1", len(got.Choices))
			}
			if got.Timings.PromptPerSecond != 10 || got.Timings.PredictedPerSecond != 20 {
				t.Errorf("timings = %+v, want prompt_per_second=10 predicted_per_second=20", got.Timings)
			}
			choice := got.Choices[0]
			if choice.FinishReason != tt.wantFinish {
				t.Errorf("finish_reason = %q, want %q", choice.FinishReason, tt.wantFinish)
			}
			if choice.Message.Content != tt.wantContent {
				t.Errorf("content = %q, want %q", choice.Message.Content, tt.wantContent)
			}
			if choice.Message.ReasoningContent != tt.wantReasoning {
				t.Errorf("reasoning_content = %q, want %q", choice.Message.ReasoningContent, tt.wantReasoning)
			}
			if len(choice.Message.ToolCalls) != len(tt.wantCalls) {
				t.Fatalf("tool_calls = %+v, want %v", choice.Message.ToolCalls, tt.wantCalls)
			}
			for i, want := range tt.wantCalls {
				fn := choice.Message.ToolCalls[i].Function
				if fn.Name != want[0] || fn.Arguments != want[1] {
					t.Errorf("tool_calls[%d] = (%q, %q), want (%q, %q)",
						i, fn.Name, fn.Arguments, want[0], want[1])
				}
			}
		})
	}
}

// Exercises the production-shared finalization and serialization seam, without
// acquiring a model or invoking runChat's generation and retry path.
func TestWriteBlockingResponsePreservesHeldText(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const tools = `[{"type":"function","function":{"name":"f","parameters":{"type":"object","properties":{}}}}]`
	for _, format := range []struct {
		name      string
		plain     bool
		prefix    []string
		reasoning string
	}{
		{name: "none", plain: true},
		{name: "separated empty reasoning"},
		{name: "separated populated reasoning", prefix: []string{"<think>", "reason", "</think>"}, reasoning: "reason"},
	} {
		for _, tail := range []struct {
			name   string
			tokens []string
			want   string
		}{
			{"open object", []string{"{"}, "{"},
			{"partial wrapper", []string{"answer", " <", "tool", "_c"}, "answer <tool_c"},
		} {
			for _, stop := range []struct {
				name   string
				reason string
				err    error
				finish string
			}{
				{"EOF", "eos", nil, "stop"},
				{"context limit", "length", geniex_sdk.ErrLlmTokenizationContextLength, "length"},
			} {
				t.Run(format.name+"/"+tail.name+"/"+stop.name, func(t *testing.T) {
					class := func() tokenClass {
						if format.plain {
							return plainClass
						}
						return reasoningClass()
					}
					tokens := append(append([]string(nil), format.prefix...), tail.tokens...)
					var content, reasoning strings.Builder
					onToken, finish := sinkWithToolCalls(class(), &content, &reasoning, tools)
					for _, token := range tokens {
						onToken(token)
					}
					w := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(w)
					profile := geniex_sdk.ProfileData{StopReason: stop.reason}
					writeBlockingResponse(c, &content, &reasoning, profile, finish)
					var got blockingBody
					if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
						t.Fatalf("body %s: %v", w.Body.Bytes(), err)
					}
					if len(got.Choices) != 1 {
						t.Fatalf("choices = %+v, want one choice", got.Choices)
					}
					choice := got.Choices[0]
					if choice.Message.Content != tail.want || choice.Message.ReasoningContent != format.reasoning || len(choice.Message.ToolCalls) != 0 || choice.FinishReason != stop.finish {
						t.Errorf("choice = %+v, want content %q, reasoning %q, no calls and finish %q", choice, tail.want, format.reasoning, stop.finish)
					}

					sw := &streamRecorder{httptest.NewRecorder(), make(chan bool)}
					sc, _ := gin.CreateTestContext(sw)
					sc.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
					dataCh := make(chan string, len(tokens))
					for _, token := range tokens {
						dataCh <- token
					}
					close(dataCh)
					streamToolCall(sc, dataCh, func() error { return stop.err }, false, &profile, class(), tools)
					deltas := sseDeltas(t, sw.Body.String())
					var streamedContent, streamedReasoning strings.Builder
					for _, delta := range deltas {
						streamedContent.WriteString(delta.Delta.Content)
						streamedReasoning.WriteString(delta.Delta.ReasoningContent)
						if len(delta.Delta.ToolCalls) != 0 {
							t.Errorf("streamed calls = %+v, want none", delta.Delta.ToolCalls)
						}
					}
					if streamedContent.String() != tail.want || streamedReasoning.String() != format.reasoning {
						t.Errorf("streamed text = (%q, %q), want (%q, %q)", streamedContent.String(), streamedReasoning.String(), tail.want, format.reasoning)
					}
					if len(deltas) == 0 {
						t.Fatal("stream emitted no terminal chunk")
					}
					if reason := deltas[len(deltas)-1].FinishReason; reason == nil || *reason != stop.finish {
						t.Errorf("streamed finish = %v, want %q", reason, stop.finish)
					}
				})
			}
		}
	}
}

func TestWriteBlockingResponseGPTOSSAbandonedRecipient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const tools = `[{"type":"function","function":{"name":"f"}}]`
	for _, prefix := range []struct {
		name   string
		tokens []string
	}{
		{"role", []string{"<|start|>", "assistant", " to"}},
		{"analysis", []string{"<|channel|>", "analysis", " to"}},
		{"final", []string{"<|channel|>", "final", " to"}},
		{"commentary", []string{"<|channel|>", "comment", "ary", " to"}},
	} {
		for _, exit := range []struct {
			name   string
			tokens []string
		}{
			{"unmatched then end", []string{"x", "<|end|>"}},
			{"empty ordinary body", []string{"<|message|>", "<|end|>"}},
		} {
			t.Run(prefix.name+"/"+exit.name, func(t *testing.T) {
				tokens := append(append([]string(nil), prefix.tokens...), exit.tokens...)
				tokens = append(tokens, "<|start|>", "assistant", " to", "=", "functions", ".f", "<|channel|>", "analysis", "<|message|>", "{}", "<|call|>")
				var content, reasoning strings.Builder
				onToken, finish := sinkWithToolCalls(reasoningClass(), &content, &reasoning, tools)
				for _, token := range tokens {
					onToken(token)
				}
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				writeBlockingResponse(c, &content, &reasoning, geniex_sdk.ProfileData{StopReason: "eos"}, finish)
				var got blockingBody
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				if len(got.Choices) != 1 {
					t.Fatalf("choices = %+v, want one", got.Choices)
				}
				choice := got.Choices[0]
				if choice.Message.Content != "" || choice.Message.ReasoningContent != "" || choice.FinishReason != "tool_calls" || len(choice.Message.ToolCalls) != 1 {
					t.Errorf("choice = %+v, want one fresh call and no display text", choice)
				} else if call := choice.Message.ToolCalls[0].Function; call.Name != "f" || call.Arguments != "{}" {
					t.Errorf("call = %+v, want f with {}", call)
				}
				deltas := runStreamToolCallWithTools(t, reasoningClass(), tools, tokens...)
				var calls []toolCallDelta
				for _, delta := range deltas {
					if delta.Delta.Content != "" || delta.Delta.ReasoningContent != "" {
						t.Errorf("streamed text = %+v, want none", delta.Delta)
					}
					for _, call := range delta.Delta.ToolCalls {
						calls = append(calls, toolCallDelta{int(call.Index), call.Function.Name, call.Function.Arguments})
					}
				}
				if len(calls) != 1 || calls[0] != (toolCallDelta{0, "f", "{}"}) {
					t.Errorf("streamed calls = %+v, want one fresh f call with {}", calls)
				}
			})
		}
	}
}
