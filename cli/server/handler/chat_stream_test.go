// Copyright (c) 2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/internal/thinkfsm"
)

// The deltas of one SSE stream, in order, with [DONE] dropped.
func sseDeltas(t *testing.T, body string) []streamChoice {
	t.Helper()
	var out []streamChoice
	for _, line := range strings.Split(body, "\n") {
		data, ok := strings.CutPrefix(line, "data:")
		if !ok || strings.TrimSpace(data) == "[DONE]" {
			continue
		}
		var chunk streamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			t.Fatalf("chunk %s: %v", data, err)
		}
		out = append(out, chunk.Choices...)
	}
	return out
}

// gin.Context.Stream waits on CloseNotify, which the plain recorder lacks.
type streamRecorder struct {
	*httptest.ResponseRecorder
	gone chan bool
}

func (r *streamRecorder) CloseNotify() <-chan bool { return r.gone }

func runStreamToolCall(t *testing.T, class tokenClass, tokens ...string) []streamChoice {
	t.Helper()
	return runStreamToolCallWithTools(t, class, "", tokens...)
}

func runStreamToolCallWithTools(t *testing.T, class tokenClass, tools string, tokens ...string) []streamChoice {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := &streamRecorder{httptest.NewRecorder(), make(chan bool)}
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)

	dataCh := make(chan string, len(tokens))
	for _, tok := range tokens {
		dataCh <- tok
	}
	close(dataCh)
	profile := geniex_sdk.ProfileData{StopReason: "eos"}
	streamToolCall(c, dataCh, func() error { return nil }, false, &profile, class, tools)
	return sseDeltas(t, w.Body.String())
}

// One tool_calls delta, flattened for comparison.
type toolCallDelta struct {
	index     int
	name      string
	arguments string
}

// Call-looking reasoning stays literal; the same syntax in Content is actionable.
func TestStreamToolCallSeparatesReasoning(t *testing.T) {
	got := runStreamToolCall(t, reasoningClass(),
		"<think>", "the city ", `{"name":"f","arg`, `uments":{"city":"Beijing"}}`, "</think>", "sure ",
		`{"name":"f","arg`, `uments":{"city":"Beijing"}}`, " done")

	var content, reasoning strings.Builder
	var calls []toolCallDelta
	for _, ch := range got {
		content.WriteString(ch.Delta.Content)
		reasoning.WriteString(ch.Delta.ReasoningContent)
		for _, tc := range ch.Delta.ToolCalls {
			calls = append(calls, toolCallDelta{int(tc.Index), tc.Function.Name, tc.Function.Arguments})
		}
	}
	if reasoning.String() != `the city {"name":"f","arguments":{"city":"Beijing"}}` {
		t.Errorf("reasoning = %q", reasoning.String())
	}
	if content.String() != "sure  done" {
		t.Errorf("content = %q", content.String())
	}
	if len(calls) != 1 || calls[0] != (toolCallDelta{0, "f", `{"city":"Beijing"}`}) {
		t.Errorf("calls = %+v", calls)
	}
	if reason := got[len(got)-1].FinishReason; reason == nil || *reason != "tool_calls" {
		t.Errorf("finish_reason = %v, want tool_calls", reason)
	}
}

func TestStreamToolCallGPTOSSChannels(t *testing.T) {
	cases := []struct {
		name   string
		tokens []string
	}{
		{
			name: "commentary header",
			tokens: []string{
				"<|channel|>", "comment", "ary", " to", "=", "functions", ".", "get_open_tabs", "<|constrain|>", "json",
				"<|message|>", `{"topic":"tabs"}`,
			},
		},
		{
			name: "channel recipient tokens",
			tokens: []string{
				"<|channel|>", "comment", "ary", " to", "=", "functions", ".", "get_open_tabs", "<|message|>", `{"topic":"tabs"}`, "<|call|>",
			},
		},
		{
			name: "role recipient at start",
			tokens: []string{
				"<|start|>", "assistant", " to", "=", "functions", ".", "get_open_tabs", "<|channel|>", "comment", "ary", "<|message|>",
				`{"topic":"tabs"}`, "<|call|>",
			},
		},
		{
			name: "role recipient after analysis",
			tokens: []string{
				"<|channel|>", "analysis", "<|message|>", "thinking", "<|end|>",
				"<|start|>", "assistant", " to", "=", "functions", ".", "get_open_tabs", "<|channel|>", "comment", "ary", "<|message|>",
				`{"topic":"tabs"}`, "<|call|>",
			},
		},
		{
			name: "analysis body",
			tokens: []string{
				"<|channel|>", "analysis", "<|message|>", " to=functions.get_open_tabs",
				"<|constrain|>", "json", "<|message|>", `{"topic":"tabs"}`,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runStreamToolCall(t, reasoningClass(), tc.tokens...)
			var content, reasoning strings.Builder
			var calls []toolCallDelta
			for _, ch := range got {
				content.WriteString(ch.Delta.Content)
				reasoning.WriteString(ch.Delta.ReasoningContent)
				for _, call := range ch.Delta.ToolCalls {
					calls = append(calls, toolCallDelta{int(call.Index), call.Function.Name, call.Function.Arguments})
				}
			}
			wantReasoning := ""
			if tc.name == "role recipient after analysis" {
				wantReasoning = "thinking"
			}
			wantCalls := []toolCallDelta{{0, "get_open_tabs", `{"topic":"tabs"}`}}
			if tc.name == "analysis body" {
				wantReasoning = ` to=functions.get_open_tabs<|constrain|>json<|message|>{"topic":"tabs"}`
				wantCalls = nil
			}
			if content.String() != "" || reasoning.String() != wantReasoning {
				t.Errorf("content = %q, reasoning = %q, want reasoning %q", content.String(), reasoning.String(), wantReasoning)
			}
			if len(calls) != len(wantCalls) {
				t.Fatalf("calls = %+v, want %+v", calls, wantCalls)
			}
			for i, call := range calls {
				if call != wantCalls[i] {
					t.Errorf("call %d = %+v, want %+v", i, call, wantCalls[i])
				}
			}
		})
	}
}

func TestStreamToolCallPreservesUnmatchedReasoningAtChannelEnd(t *testing.T) {
	got := runStreamToolCall(t, reasoningClass(),
		"<|channel|>", "analysis", "<|message|>", "This is to=functions", "<|end|>",
		"<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "answer")
	var content, reasoning strings.Builder
	for _, ch := range got {
		content.WriteString(ch.Delta.Content)
		reasoning.WriteString(ch.Delta.ReasoningContent)
	}
	if content.String() != "answer" || reasoning.String() != "This is to=functions" {
		t.Errorf("content = %q, reasoning = %q", content.String(), reasoning.String())
	}
}

func TestBlockingSinkDoesNotParseReasoningToolCalls(t *testing.T) {
	var content, reasoning strings.Builder
	onToken, finish := sinkWithToolCalls(reasoningClass(), &content, &reasoning, "")
	for _, token := range []string{
		"<|channel|>", "analysis", "<|message|>", "I should check ", " to=functions.get_open_tabs",
		"<|constrain|>", "json", "<|message|>", `{"topic":"tabs"}`,
	} {
		onToken(token)
	}
	calls := finish()
	if content.String() != "" || reasoning.String() != `I should check  to=functions.get_open_tabs<|constrain|>json<|message|>{"topic":"tabs"}` {
		t.Errorf("content = %q, reasoning = %q", content.String(), reasoning.String())
	}
	if len(calls) != 0 {
		t.Errorf("reasoning produced calls = %+v, want none", calls)
	}
}

func TestStreamToolCallGemma4EmptyThought(t *testing.T) {
	got := runStreamToolCall(t, reasoningClass(),
		"<|channel>", "thought", "<channel|>", `<|tool_call>call:go{}<tool_call|>`)

	var content, reasoning strings.Builder
	var calls []toolCallDelta
	for _, ch := range got {
		content.WriteString(ch.Delta.Content)
		reasoning.WriteString(ch.Delta.ReasoningContent)
		for _, tc := range ch.Delta.ToolCalls {
			calls = append(calls, toolCallDelta{int(tc.Index), tc.Function.Name, tc.Function.Arguments})
		}
	}
	if content.String() != "" || reasoning.String() != "" {
		t.Errorf("content = %q, reasoning = %q, want both empty", content.String(), reasoning.String())
	}
	if len(calls) != 1 || calls[0] != (toolCallDelta{0, "go", `{}`}) {
		t.Errorf("calls = %+v", calls)
	}
	if reason := got[len(got)-1].FinishReason; reason == nil || *reason != "tool_calls" {
		t.Errorf("finish_reason = %v, want tool_calls", reason)
	}
}

func TestStreamMiniCPM5TypedArguments(t *testing.T) {
	const tools = `[{"type":"function","function":{"name":"set","parameters":{"type":"object","properties":{"count":{"type":"integer"},"label":{"type":"string"}}}}}]`
	got := runStreamToolCallWithTools(t, plainClass, tools,
		"<function", ` name="set">`, "<param", ` name="count">5`, "</param>",
		"<param", ` name="label">5`, "</param>", "</function>")
	var calls []toolCallDelta
	for _, ch := range got {
		if ch.Delta.Content != "" {
			t.Errorf("XML leaked into content: %q", ch.Delta.Content)
		}
		for _, tc := range ch.Delta.ToolCalls {
			calls = append(calls, toolCallDelta{int(tc.Index), tc.Function.Name, tc.Function.Arguments})
		}
	}
	if len(calls) != 1 || calls[0] != (toolCallDelta{0, "set", `{"count":5,"label":"5"}`}) {
		t.Errorf("calls = %+v", calls)
	}
	if reason := got[len(got)-1].FinishReason; reason == nil || *reason != "tool_calls" {
		t.Errorf("finish_reason = %v, want tool_calls", reason)
	}
}

func TestToolCallEventConsumers(t *testing.T) {
	type consumerCase struct {
		name               string
		plain              bool
		toolOnly           bool
		tokens             []string
		content, reasoning string
		calls              []toolCallDelta
	}
	cases := []consumerCase{
		{
			name:    "plain keeps markers inline",
			plain:   true,
			tokens:  []string{"<think>", "reason", "</think>", "<|channel|>", "analysis", "<|message|>", "answer", "<|end|>"},
			content: "<think>reason</think><|channel|>analysis<|message|>answer<|end|>",
		},
		{
			name:    "separated think transition",
			tokens:  []string{"<think>", "reason", "</think>", "answer"},
			content: "answer", reasoning: "reason",
		},
		{
			name:    "ordinary commentary preamble",
			tokens:  []string{"<|start|>", "assistant", "<|channel|>", "comment", "ary", "<|message|>", "Let me check.", "<|call|>"},
			content: "Let me check.",
		},
		{
			name:     "commentary get_weather actual recipient tokens",
			toolOnly: true,
			tokens:   []string{"<|channel|>", "comment", "ary", " to", "=", "functions", ".get", "_weather", "<|message|>", `{"city":"Paris"}`, "<|call|>"},
			calls:    []toolCallDelta{{0, "get_weather", `{"city":"Paris"}`}},
		},
		{
			name:     "analysis get_weather actual role recipient tokens",
			toolOnly: true,
			tokens:   []string{"<|start|>", "assistant", " to", "=", "functions", ".get", "_weather", "<|channel|>", "analysis", "<|message|>", `{"city":"Paris"}`, "<|end|>"},
			calls:    []toolCallDelta{{0, "get_weather", `{"city":"Paris"}`}},
		},
		{
			name:    "channel recipient ended before ordinary analysis",
			tokens:  []string{"<|channel|>", "final", " to", "=", "functions", ".f", "<|channel|>", "analysis", "<|message|>", "reason", "<|end|>", "<|start|>", "assistant", "<|message|>", "answer"},
			content: "answer", reasoning: "reason",
		},
		{
			name:    "role recipient ended on second channel",
			tokens:  []string{"<|start|>", "assistant", " to", "=", "functions", ".f", "<|channel|>", "analysis", "<|channel|>", "final", "<|message|>", "answer", "<|end|>"},
			content: "answer",
		},
		{
			name:    "role recipient unknown first channel then supported ends candidate",
			tokens:  []string{"<|start|>", "assistant", " to", "=", "functions", ".f", "<|channel|>", "custom", "<|channel|>", "final", "<|message|>", "answer", "<|end|>"},
			content: "answer",
		},
		{
			name:   "fresh recipient after header reset excludes buffered old function",
			tokens: []string{"<|channel|>", "analysis", " to", "=", "functions", ".f", "<|channel|>", "comment", "ary", " to", "=", "functions", ".get", "_weather", "<|message|>", `{"city":"Paris"}`, "<|call|>"},
			calls:  []toolCallDelta{{0, "get_weather", `{"city":"Paris"}`}},
		},
		{
			name:     "commentary raw syntax",
			toolOnly: true,
			tokens:   []string{"<|channel|>", "comment", "ary", " to", "=", "functions", ".f", "<|message|>", `{"x":1}`, "<|call|>"},
			calls:    []toolCallDelta{{0, "f", `{"x":1}`}},
		},
		{
			name:     "analysis tool path through call boundary",
			toolOnly: true,
			tokens:   []string{"<|channel|>", "analysis", " to", "=", "functions", ".f", "<|message|>", `{"x":1}`, "<|call|>"},
			calls:    []toolCallDelta{{0, "f", `{"x":1}`}},
		},
		{
			name:     "analysis tool path through end boundary",
			toolOnly: true,
			tokens:   []string{"<|channel|>", "analysis", " to", "=", "functions", ".f", "<|message|>", `{"x":1}`, "<|end|>"},
			calls:    []toolCallDelta{{0, "f", `{"x":1}`}},
		},
		{
			name: "analysis literal prefix preserved at end",
			tokens: []string{"<|channel|>", "analysis", "<|message|>", "This is to=functions", "<|end|>",
				"<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "answer"},
			content: "answer", reasoning: "This is to=functions",
		},
		{
			name:      "analysis literal prefix preserved at call",
			tokens:    []string{"<|channel|>", "analysis", "<|message|>", "This is to=functions", "<|call|>"},
			reasoning: "This is to=functions",
		},
		{
			name:   "unmatched commentary header tail is discarded",
			tokens: []string{"<|channel|>", "comment", "ary", " to", "=", "functions", "."},
		},
		{
			name:   "builtin body cannot become a generic function call",
			tokens: []string{"<|channel|>", "analysis", " to", "=", "python", "<|message|>", `{"name":"fake","arguments":{}}`, "<|call|>"},
		},
		{
			name:   "builtin body cannot contain a nested function recipient",
			tokens: []string{"<|channel|>", "analysis", " to", "=", "python", "<|message|>", ` to=functions.fake<|message|>{}`, "<|end|>"},
		},
		{
			name:   "damaged JSON body is discarded at EOF",
			tokens: []string{"<|channel|>", "comment", "ary", " to", "=", "functions", ".f", "<|message|>", `{"x":`},
		},
		{
			name:    "complete tool body without end recovers once before fresh final",
			tokens:  []string{"<|channel|>", "comment", "ary", " to", "=", "functions", ".f", "<|message|>", `{"x":1}`, "<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "answer", "<|end|>"},
			content: "answer",
			calls:   []toolCallDelta{{0, "f", `{"x":1}`}},
		},
		{
			name:    "incomplete tool body without end discarded before fresh final",
			tokens:  []string{"<|channel|>", "comment", "ary", " to", "=", "functions", ".f", "<|message|>", `{"x":`, "<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "answer", "<|end|>"},
			content: "answer",
		},
		{
			name:    "non JSON body ends before fresh final",
			tokens:  []string{"<|channel|>", "comment", "ary", " to", "=", "functions", ".f", "<|message|>", "not JSON", "<|end|>", "<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "answer"},
			content: "answer",
		},
		{
			name:   "malformed arguments cannot recover a later nested object",
			tokens: []string{"<|channel|>", "analysis", " to", "=", "functions", ".f", "<|message|>", `oops {"x":1}`, "<|call|>"},
		},
	}
	for _, syntax := range []struct {
		name, text, args string
	}{
		{"bare JSON", `{"name":"f","arguments":{"x":1}}`, `{"x":1}`},
		{"Qwen wrapper", `<tool_call>{"name":"f","arguments":{"x":1}}</tool_call>`, `{"x":1}`},
		{"Gemma wrapper", `<|tool_call>call:f{x:1}<tool_call|>`, `{"x":1}`},
		{"MiniCPM XML", `<function name="f"><param name="x">1</param></function>`, `{"x":"1"}`},
		{"GPTOSS recipient text", ` to=functions.f<|message|>{"x":1}<|call|>`, `{"x":1}`},
	} {
		cases = append(cases,
			consumerCase{name: "literal reasoning/" + syntax.name, tokens: []string{"<think>", syntax.text, "</think>", "answer"}, content: "answer", reasoning: syntax.text},
			consumerCase{name: "content call/" + syntax.name, tokens: []string{"<think>", syntax.text, "</think>", syntax.text}, reasoning: syntax.text, calls: []toolCallDelta{{0, "f", syntax.args}}},
		)
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			newClass := func() tokenClass {
				if tc.plain {
					return plainClass
				}
				class := reasoningClass()
				return func(token string) thinkfsm.Event {
					ev := class(token)
					if tc.toolOnly {
						switch ev.Kind {
						case thinkfsm.Consume, thinkfsm.ToolCall, thinkfsm.ToolCallStart, thinkfsm.ToolCallEnd:
						default:
							t.Errorf("tool message token %q = %+v, want tool input or Consume, never display or thinking boundaries", token, ev)
						}
					}
					return ev
				}
			}
			t.Run("stream", func(t *testing.T) {
				got := runStreamToolCall(t, newClass(), tc.tokens...)
				var content, reasoning strings.Builder
				var calls []toolCallDelta
				for _, ch := range got {
					content.WriteString(ch.Delta.Content)
					reasoning.WriteString(ch.Delta.ReasoningContent)
					for _, call := range ch.Delta.ToolCalls {
						calls = append(calls, toolCallDelta{int(call.Index), call.Function.Name, call.Function.Arguments})
					}
				}
				if content.String() != tc.content || reasoning.String() != tc.reasoning {
					t.Errorf("content = %q, reasoning = %q, want %q, %q", content.String(), reasoning.String(), tc.content, tc.reasoning)
				}
				if len(calls) != len(tc.calls) {
					t.Fatalf("calls = %+v, want %+v", calls, tc.calls)
				}
				for i, call := range calls {
					if call != tc.calls[i] {
						t.Errorf("call %d = %+v, want %+v", i, call, tc.calls[i])
					}
				}
				wantFinish := "stop"
				if len(tc.calls) > 0 {
					wantFinish = "tool_calls"
				}
				if reason := got[len(got)-1].FinishReason; reason == nil || *reason != wantFinish {
					t.Errorf("finish_reason = %v, want %s", reason, wantFinish)
				}
			})
			t.Run("blocking sink", func(t *testing.T) {
				var content, reasoning strings.Builder
				onToken, finish := sinkWithToolCalls(newClass(), &content, &reasoning, "")
				for _, token := range tc.tokens {
					onToken(token)
				}
				calls := finish()
				if content.String() != tc.content || reasoning.String() != tc.reasoning {
					t.Errorf("content = %q, reasoning = %q, want %q, %q", content.String(), reasoning.String(), tc.content, tc.reasoning)
				}
				if len(calls) != len(tc.calls) {
					t.Fatalf("calls = %+v, want %+v", calls, tc.calls)
				}
				for i, call := range calls {
					if call.Name != tc.calls[i].name || call.Arguments != tc.calls[i].arguments {
						t.Errorf("call %d = %+v, want %+v", i, call, tc.calls[i])
					}
				}
			})
		})
	}
}

// ff_log.txt is a full_text log, not a token callback dump. These sanitized,
// synthetic token vectors preserve the three observed output structures.
func TestGPTOSSLogStructureOutputs(t *testing.T) {
	cases := []struct {
		name               string
		tokens             []string
		content, reasoning string
		calls              []toolCallDelta
	}{
		{
			name:      "headerless analysis then final",
			tokens:    []string{"<|channel|>", "analysis", "<|message|>", "Consider options.", "<|end|>", "<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "Updated answer."},
			content:   "Updated answer.",
			reasoning: "Consider options.",
		},
		{
			name:      "analysis then constrained commentary call at EOF",
			tokens:    []string{"<|channel|>", "analysis", "<|message|>", "Consider options.", "<|end|>", "<|start|>", "assistant", "<|channel|>", "comment", "ary", " to", "=", "functions", ".f", " ", "<|constrain|>", "json", "<|message|>", `{"topic":""}`},
			reasoning: "Consider options.",
			calls:     []toolCallDelta{{0, "f", `{"topic":""}`}},
		},
		{
			name:      "incomplete tool header ended before second thinking body",
			tokens:    []string{"<|channel|>", "analysis", "<|message|>", "Consider options.", "<|end|>", "<|start|>", "assistant", "<|channel|>", "comment", "ary", " to", "=", "functions", ".f", "<|channel|>", "comment", "ary", "<|channel|>", "analysis", "<|message|>", "Need an indoor option.", "<|end|>", "<|start|>", "assistant", "<|channel|>", "final", "<|message|>", "Updated answer."},
			content:   "Updated answer.",
			reasoning: "Consider options.Need an indoor option.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := func(t *testing.T, content, reasoning string, calls []toolCallDelta) {
				t.Helper()
				for _, tag := range []string{"<|start|>", "<|channel|>", "<|message|>", "<|end|>", "<|call|>", "<|constrain|>"} {
					if strings.Contains(content, tag) || strings.Contains(reasoning, tag) {
						t.Errorf("protocol tag %q leaked: content = %q, reasoning = %q", tag, content, reasoning)
					}
				}
				if content != tc.content || reasoning != tc.reasoning {
					t.Errorf("content = %q, reasoning = %q, want %q, %q", content, reasoning, tc.content, tc.reasoning)
				}
				if len(calls) != len(tc.calls) {
					t.Fatalf("calls = %+v, want %+v", calls, tc.calls)
				}
				for i, call := range calls {
					if call != tc.calls[i] {
						t.Errorf("call %d = %+v, want %+v", i, call, tc.calls[i])
					}
				}
			}
			t.Run("stream", func(t *testing.T) {
				got := runStreamToolCall(t, reasoningClass(), tc.tokens...)
				var content, reasoning strings.Builder
				var calls []toolCallDelta
				for _, ch := range got {
					content.WriteString(ch.Delta.Content)
					reasoning.WriteString(ch.Delta.ReasoningContent)
					for _, call := range ch.Delta.ToolCalls {
						calls = append(calls, toolCallDelta{int(call.Index), call.Function.Name, call.Function.Arguments})
					}
				}
				check(t, content.String(), reasoning.String(), calls)
			})
			t.Run("blocking sink", func(t *testing.T) {
				var content, reasoning strings.Builder
				onToken, finish := sinkWithToolCalls(reasoningClass(), &content, &reasoning, "")
				for _, token := range tc.tokens {
					onToken(token)
				}
				var calls []toolCallDelta
				for i, call := range finish() {
					calls = append(calls, toolCallDelta{i, call.Name, call.Arguments})
				}
				check(t, content.String(), reasoning.String(), calls)
			})
		})
	}
}

// Parallel calls need a rising index, and plain text needs no tool_calls finish.
func TestStreamToolCallIndexes(t *testing.T) {
	got := runStreamToolCall(t, plainClass,
		`{"name":"a","arguments":{}}`, `{"name":"b","arguments":{}}`)
	var indexes []int64
	for _, ch := range got {
		for _, tc := range ch.Delta.ToolCalls {
			indexes = append(indexes, tc.Index)
		}
	}
	if len(indexes) != 2 || indexes[0] != 0 || indexes[1] != 1 {
		t.Errorf("indexes = %v, want [0 1]", indexes)
	}

	got = runStreamToolCall(t, plainClass, "no ", "call ", "here")
	last := got[len(got)-1]
	if last.FinishReason == nil || *last.FinishReason != "stop" {
		t.Errorf("finish_reason = %v, want stop", last.FinishReason)
	}
}
