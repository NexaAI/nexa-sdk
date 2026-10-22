// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package utils

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/qualcomm/GenieX/cli/internal/thinkfsm"
)

// The handler receives only display text or structured calls, never FSM boundaries.
func checkToolCallParserEvents(t *testing.T, events []ToolCallEvent) {
	t.Helper()
	for _, ev := range events {
		switch ev.Kind {
		case thinkfsm.Content, thinkfsm.Thinking:
			if ev.Text == "" || ev.Call.Name != "" || ev.Call.Arguments != "" {
				t.Errorf("display event = %+v, want nonempty text and no call", ev)
			}
		case thinkfsm.ToolCall:
			if ev.Text != "" || ev.Call.Name == "" {
				t.Errorf("call event = %+v, want structured call and no text", ev)
			}
		default:
			t.Errorf("internal FSM event exposed to handler: %+v", ev)
		}
	}
}

func toolCallParserOutput(t *testing.T, events []ToolCallEvent) (string, string, []toolCallFn) {
	t.Helper()
	checkToolCallParserEvents(t, events)
	var content, thinking strings.Builder
	var calls []toolCallFn
	for _, ev := range events {
		switch ev.Kind {
		case thinkfsm.Content:
			content.WriteString(ev.Text)
		case thinkfsm.Thinking:
			thinking.WriteString(ev.Text)
		case thinkfsm.ToolCall:
			calls = append(calls, ev.Call)
		}
	}
	return content.String(), thinking.String(), calls
}

func feedToolCallParser(t *testing.T, parser *ToolCallParser, input []thinkfsm.Event) []ToolCallEvent {
	t.Helper()
	var output []ToolCallEvent
	for _, ev := range input {
		events := parser.Feed(ev)
		checkToolCallParserEvents(t, events)
		output = append(output, events...)
	}
	return output
}

func directedToolCallEvents(header, args []string, end bool) []thinkfsm.Event {
	var input []thinkfsm.Event
	for _, text := range header {
		input = append(input, thinkfsm.Event{Kind: thinkfsm.ToolCall, Text: text})
	}
	input = append(input, thinkfsm.Event{Kind: thinkfsm.ToolCallStart})
	for _, text := range args {
		input = append(input, thinkfsm.Event{Kind: thinkfsm.ToolCall, Text: text})
	}
	if end {
		input = append(input, thinkfsm.Event{Kind: thinkfsm.ToolCallEnd})
	}
	return input
}

func TestToolCallParserDirectedInput(t *testing.T) {
	cases := []struct {
		name   string
		header []string
		args   []string
		end    bool
		want   []toolCallFn
	}{
		{"complete at end", []string{" to", "=", "functions", ".f"}, []string{`{"x":1}`}, true, []toolCallFn{{Name: "f", Arguments: `{"x":1}`}}},
		{"complete at EOF", []string{" to", "=", "functions", ".get", "_weather", " "}, []string{`{"city":`, `"Paris"`, "}"}, false, []toolCallFn{{Name: "get_weather", Arguments: `{"city":"Paris"}`}}},
		{"nested object and quoted brace", []string{" to=functions.f"}, []string{`{"a":{"b":[1,2]},"s":"}"}`}, true, []toolCallFn{{Name: "f", Arguments: `{"a":{"b":[1,2]},"s":"}"}`}}},
		{"trim body whitespace", []string{" to=functions.f"}, []string{" \n", `{"x": 1}`, "\t "}, false, []toolCallFn{{Name: "f", Arguments: `{"x": 1}`}}},
		{"allowed name characters", []string{" to=functions.A1_b-c.d"}, []string{"{}"}, true, []toolCallFn{{Name: "A1_b-c.d", Arguments: `{}`}}},
		{"unclosed JSON", []string{" to=functions.f"}, []string{`{"x":`}, false, nil},
		{"invalid JSON", []string{" to=functions.f"}, []string{`{"x":,}`}, true, nil},
		{"text before JSON", []string{" to=functions.f"}, []string{`oops {"x":1}`}, true, nil},
		{"text after JSON", []string{" to=functions.f"}, []string{`{} trailing`}, true, nil},
		{"two argument objects", []string{" to=functions.f"}, []string{`{} {}`}, true, nil},
		{"array arguments", []string{" to=functions.f"}, []string{`[{"x":1}]`}, true, nil},
		{"string arguments", []string{" to=functions.f"}, []string{`"value"`}, false, nil},
		{"no arguments", []string{" to=functions.f"}, nil, true, nil},
		{"header prefix garbage", []string{"garbage to=functions.f"}, []string{"{}"}, true, nil},
		{"header missing leading space", []string{"to=functions.f"}, []string{"{}"}, true, nil},
		{"wrong functions prefix", []string{" to=functions_extra.f"}, []string{"{}"}, false, nil},
		{"invalid first name character", []string{" to=functions.$f"}, []string{"{}"}, true, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewToolCallParser(nil)
			output := feedToolCallParser(t, parser, directedToolCallEvents(tc.header, tc.args, tc.end))
			output = append(output, parser.Finish()...)
			content, thinking, calls := toolCallParserOutput(t, output)
			if content != "" || thinking != "" || !callsEqual(calls, tc.want) {
				t.Errorf("output = (%q, %q, %+v), want no display text and %+v", content, thinking, calls, tc.want)
			}
			if events := parser.Finish(); len(events) != 0 {
				t.Errorf("second Finish = %+v, want empty", events)
			}
		})
	}
}

func TestToolCallParserBuiltinBodiesAreDiscarded(t *testing.T) {
	for _, recipient := range []string{"browser", "python", "other", "container.exec"} {
		for _, body := range []string{
			`{"name":"fake","arguments":{}}`,
			` to=functions.fake {"name":"nested","arguments":{}}`,
			`{"nested":{"name":"fake","arguments":{}}}`,
			"print(1)",
		} {
			for _, end := range []bool{false, true} {
				t.Run(recipient+"/"+body+"/end="+strconv.FormatBool(end), func(t *testing.T) {
					parser := NewToolCallParser(nil)
					input := directedToolCallEvents([]string{" to", "=", recipient}, []string{body}, end)
					output := append(feedToolCallParser(t, parser, input), parser.Finish()...)
					if len(output) != 0 {
						t.Errorf("nonfunctions body = %+v, want neither generic calls nor display text", output)
					}
				})
			}
		}
	}
}

func TestToolCallParserConsumesInternalEvents(t *testing.T) {
	parser := NewToolCallParser(map[string]map[string]string{"other": {"x": "integer"}})
	for _, ev := range []thinkfsm.Event{
		{Kind: thinkfsm.Consume, Text: "header garbage"},
		{Kind: thinkfsm.ToolCallStart},
		{Kind: thinkfsm.ToolCallEnd},
		{Kind: thinkfsm.EnterThinking},
		{Kind: thinkfsm.LeaveThinking},
	} {
		if events := parser.Feed(ev); len(events) != 0 {
			t.Errorf("internal Feed(%+v) = %+v, want no output", ev, events)
		}
	}
	input := []thinkfsm.Event{
		{Kind: thinkfsm.ToolCall, Text: " to=functions.f"},
		{Kind: thinkfsm.Consume, Text: "json"},
		{Kind: thinkfsm.Consume, Text: "constraint header garbage"},
		{Kind: thinkfsm.ToolCallStart},
		{Kind: thinkfsm.ToolCall, Text: "{"},
		{Kind: thinkfsm.Consume, Text: "ignored"},
		{Kind: thinkfsm.ToolCall, Text: "}"},
		{Kind: thinkfsm.ToolCallEnd},
	}
	output := append(feedToolCallParser(t, parser, input), parser.Finish()...)
	content, thinking, calls := toolCallParserOutput(t, output)
	if content != "" || thinking != "" || !callsEqual(calls, []toolCallFn{{Name: "f", Arguments: `{}`}}) {
		t.Errorf("consumed garbage output = (%q, %q, %+v), want f call independent of schema names", content, thinking, calls)
	}
}

func TestToolCallParserHeaderOnlyIsDiscarded(t *testing.T) {
	for _, input := range [][]thinkfsm.Event{
		{{Kind: thinkfsm.ToolCall, Text: " to=functions.f"}},
		{{Kind: thinkfsm.ToolCall, Text: " to=functions.f"}, {Kind: thinkfsm.ToolCall, Text: "{}"}},
		{{Kind: thinkfsm.ToolCall, Text: " to=functions.f"}, {Kind: thinkfsm.ToolCallEnd}},
	} {
		parser := NewToolCallParser(nil)
		output := append(feedToolCallParser(t, parser, input), parser.Finish()...)
		if len(output) != 0 {
			t.Errorf("header-only input %+v emitted %+v", input, output)
		}
	}
}

func TestToolCallParserEndCompletionIsIdempotent(t *testing.T) {
	parser := NewToolCallParser(nil)
	input := directedToolCallEvents([]string{" to", "=", "functions", ".f"}, []string{"{}"}, false)
	if output := feedToolCallParser(t, parser, input); len(output) != 0 {
		t.Fatalf("unfinished tool emitted %+v", output)
	}
	output := parser.Feed(thinkfsm.Event{Kind: thinkfsm.ToolCallEnd})
	content, thinking, calls := toolCallParserOutput(t, output)
	if content != "" || thinking != "" || !callsEqual(calls, []toolCallFn{{Name: "f", Arguments: `{}`}}) {
		t.Errorf("end output = (%q, %q, %+v)", content, thinking, calls)
	}
	for _, ev := range []thinkfsm.Event{{Kind: thinkfsm.ToolCallEnd}, {Kind: thinkfsm.ToolCallEnd}} {
		if output := parser.Feed(ev); len(output) != 0 {
			t.Errorf("after completion Feed(%+v) = %+v, want empty", ev, output)
		}
	}
	if output := parser.Finish(); len(output) != 0 {
		t.Errorf("after completion Finish = %+v, want empty", output)
	}
}

func TestToolCallParserEndResetsCandidatePrefixes(t *testing.T) {
	old := directedToolCallEvents([]string{" to", "=", "functions", ".old"}, []string{`{"x":`, "1}"}, false)
	for n := 0; n <= len(old); n++ {
		t.Run("prefix "+strconv.Itoa(n), func(t *testing.T) {
			parser := NewToolCallParser(nil)
			output := feedToolCallParser(t, parser, old[:n])
			output = append(output, parser.Feed(thinkfsm.Event{Kind: thinkfsm.ToolCallEnd})...)
			var want []toolCallFn
			if n == len(old) {
				want = append(want, toolCallFn{Name: "old", Arguments: `{"x":1}`})
			}
			content, thinking, calls := toolCallParserOutput(t, output)
			if content != "" || thinking != "" || !callsEqual(calls, want) {
				t.Fatalf("ended prefix = (%q, %q, %+v), want calls %+v", content, thinking, calls, want)
			}
			for i := 0; i < 2; i++ {
				if events := parser.Feed(thinkfsm.Event{Kind: thinkfsm.ToolCallEnd}); len(events) != 0 {
					t.Errorf("repeated End = %+v, want empty", events)
				}
				if events := parser.Finish(); len(events) != 0 {
					t.Errorf("Finish after End = %+v, want empty", events)
				}
			}
			fresh := directedToolCallEvents([]string{" to", "=", "functions", ".fresh"}, []string{`{"x":1}`}, true)
			output = append(output, feedToolCallParser(t, parser, fresh)...)
			output = append(output, parser.Finish()...)
			want = append(want, toolCallFn{Name: "fresh", Arguments: `{"x":1}`})
			content, thinking, calls = toolCallParserOutput(t, output)
			if content != "" || thinking != "" || !callsEqual(calls, want) {
				t.Errorf("old then fresh output = (%q, %q, %+v), want %+v", content, thinking, calls, want)
			}
		})
	}
}

func TestToolCallParserEndDiscardsInvalidCandidate(t *testing.T) {
	for _, args := range []string{`{"x":,}`, `{"x":`, `{} trailing`} {
		t.Run(args, func(t *testing.T) {
			parser := NewToolCallParser(nil)
			old := directedToolCallEvents([]string{" to=functions.old"}, []string{args}, true)
			output := append(feedToolCallParser(t, parser, old), parser.Finish()...)
			if len(output) != 0 {
				t.Fatalf("invalid candidate emitted %+v", output)
			}
			fresh := directedToolCallEvents([]string{" to=functions.fresh"}, []string{"{}"}, true)
			output = append(feedToolCallParser(t, parser, fresh), parser.Finish()...)
			content, thinking, calls := toolCallParserOutput(t, output)
			if content != "" || thinking != "" || !callsEqual(calls, []toolCallFn{{Name: "fresh", Arguments: `{}`}}) {
				t.Errorf("after invalid candidate = (%q, %q, %+v), want fresh only", content, thinking, calls)
			}
		})
	}
}

func TestToolCallParserDisplayTransitionsEndCandidate(t *testing.T) {
	for _, kind := range []thinkfsm.EventKind{thinkfsm.Content, thinkfsm.Thinking, thinkfsm.EnterThinking, thinkfsm.LeaveThinking} {
		for _, args := range []string{"{}", `{"x":`} {
			t.Run("kind "+strconv.Itoa(int(kind))+"/"+args, func(t *testing.T) {
				parser := NewToolCallParser(nil)
				input := directedToolCallEvents([]string{" to=functions.old"}, []string{args}, false)
				output := feedToolCallParser(t, parser, input)
				ev := thinkfsm.Event{Kind: kind}
				var want []ToolCallEvent
				if args == "{}" {
					want = append(want, ToolCallEvent{Kind: thinkfsm.ToolCall, Call: toolCallFn{Name: "old", Arguments: `{}`}})
				}
				if kind == thinkfsm.Content || kind == thinkfsm.Thinking {
					ev.Text = "next"
					want = append(want, ToolCallEvent{Kind: kind, Text: "next"})
				}
				output = append(output, parser.Feed(ev)...)
				checkToolCallParserEvents(t, output)
				if !reflect.DeepEqual(output, want) {
					t.Errorf("display transition = %+v, want ordered events %+v", output, want)
				}
				if events := parser.Feed(thinkfsm.Event{Kind: thinkfsm.ToolCallEnd}); len(events) != 0 {
					t.Errorf("End after display transition = %+v, want empty", events)
				}
				if events := parser.Finish(); len(events) != 0 {
					t.Errorf("Finish after display transition = %+v, want empty", events)
				}
				output = parser.Feed(thinkfsm.Event{Kind: thinkfsm.Content, Text: "answer"})
				output = append(output, parser.Feed(thinkfsm.Event{Kind: thinkfsm.Thinking, Text: "reason"})...)
				output = append(output, parser.Finish()...)
				content, thinking, calls := toolCallParserOutput(t, output)
				if content != "answer" || thinking != "reason" || len(calls) != 0 {
					t.Errorf("next display = (%q, %q, %+v), want answer/reason without stale call", content, thinking, calls)
				}
			})
		}
	}
}

func TestToolCallParserInternalBoundariesAreNotOutput(t *testing.T) {
	parser := NewToolCallParser(nil)
	input := []thinkfsm.Event{
		{Kind: thinkfsm.Consume, Text: "ignored header"},
		{Kind: thinkfsm.EnterThinking},
		{Kind: thinkfsm.Thinking, Text: "first reason"},
		{Kind: thinkfsm.LeaveThinking},
		{Kind: thinkfsm.ToolCall, Text: " to=functions.old"},
		{Kind: thinkfsm.Consume, Text: "constraint type and header garbage"},
		{Kind: thinkfsm.ToolCallEnd},
		{Kind: thinkfsm.EnterThinking},
		{Kind: thinkfsm.Thinking, Text: "second reason"},
		{Kind: thinkfsm.LeaveThinking},
		{Kind: thinkfsm.Content, Text: "answer"},
	}
	input = append(input, directedToolCallEvents([]string{" to", "=", "functions", ".f"}, []string{"{}"}, true)...)
	output := append(feedToolCallParser(t, parser, input), parser.Finish()...)
	content, thinking, calls := toolCallParserOutput(t, output)
	if content != "answer" || thinking != "first reasonsecond reason" || !callsEqual(calls, []toolCallFn{{Name: "f", Arguments: `{}`}}) {
		t.Errorf("output = (%q, %q, %+v), want only display bodies and f call", content, thinking, calls)
	}
}

func TestToolCallParserFlushTextAttribution(t *testing.T) {
	cases := []struct {
		name     string
		kind     thinkfsm.EventKind
		boundary thinkfsm.EventKind
		next     thinkfsm.EventKind
	}{
		{"content to thinking", thinkfsm.Content, thinkfsm.Thinking, thinkfsm.Thinking},
		{"thinking to content", thinkfsm.Thinking, thinkfsm.Content, thinkfsm.Content},
		{"enter thinking", thinkfsm.Content, thinkfsm.EnterThinking, thinkfsm.Thinking},
		{"leave thinking", thinkfsm.Thinking, thinkfsm.LeaveThinking, thinkfsm.Content},
		{"content to tool header", thinkfsm.Content, thinkfsm.ToolCall, thinkfsm.Content},
		{"thinking to tool header", thinkfsm.Thinking, thinkfsm.ToolCall, thinkfsm.Thinking},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			parser := NewToolCallParser(nil)
			first := parser.Feed(thinkfsm.Event{Kind: tc.kind, Text: "{"})
			checkToolCallParserEvents(t, first)
			if tc.kind == thinkfsm.Content {
				if len(first) != 0 {
					t.Fatalf("held content brace emitted %+v", first)
				}
			} else if len(first) != 1 || first[0].Kind != thinkfsm.Thinking || first[0].Text != "{" {
				t.Fatalf("thinking brace = %+v, want immediate untouched Thinking", first)
			}
			ev := thinkfsm.Event{Kind: tc.boundary}
			if tc.boundary == thinkfsm.ToolCall {
				ev.Text = " to=functions.old"
			}
			output := parser.Feed(ev)
			checkToolCallParserEvents(t, output)
			if tc.kind == thinkfsm.Content {
				if len(output) != 1 || output[0].Kind != thinkfsm.Content || output[0].Text != "{" {
					t.Fatalf("boundary flush = %+v, want brace attributed to Content", output)
				}
			} else if len(output) != 0 {
				t.Fatalf("thinking bypass left a held prefix: %+v", output)
			}
			output = append(parser.Feed(thinkfsm.Event{Kind: tc.next, Text: "next"}), parser.Finish()...)
			checkToolCallParserEvents(t, output)
			if len(output) != 1 || output[0].Kind != tc.next || output[0].Text != "next" {
				t.Errorf("after boundary = %+v, want next text in kind %v", output, tc.next)
			}
		})
	}
}

func TestToolCallParserThinkingFlushesContentWithoutScanning(t *testing.T) {
	parser := NewToolCallParser(nil)
	if events := parser.Feed(thinkfsm.Event{Kind: thinkfsm.Content, Text: "{"}); len(events) != 0 {
		t.Fatalf("held Content brace emitted %+v", events)
	}
	const reasoning = `"name":"f","arguments":{}}`
	output := parser.Feed(thinkfsm.Event{Kind: thinkfsm.Thinking, Text: reasoning})
	want := []ToolCallEvent{
		{Kind: thinkfsm.Content, Text: "{"},
		{Kind: thinkfsm.Thinking, Text: reasoning},
	}
	checkToolCallParserEvents(t, output)
	if !reflect.DeepEqual(output, want) {
		t.Errorf("Thinking boundary = %+v, want ordered Content then literal Thinking %+v", output, want)
	}
	if events := parser.Finish(); len(events) != 0 {
		t.Errorf("Finish = %+v, want no cross-kind call or held text", events)
	}
}

func TestToolCallParserThinkingCallSyntaxIsImmediate(t *testing.T) {
	for _, text := range []string{
		`{"name":"f","arguments":{"x":1}}`,
		`<tool_call>{"name":"f","arguments":{"x":1}}</tool_call>`,
		`<|tool_call>call:f{x:1}<tool_call|>`,
		`<function name="f"><param name="x">1</param></function>`,
		` to=functions.f<|message|>{"x":1}<|call|>`,
	} {
		t.Run(text, func(t *testing.T) {
			parser := NewToolCallParser(nil)
			// Split inside each opener: even partial call-looking reasoning is immediate.
			for _, part := range []string{text[:1], text[1:]} {
				output := parser.Feed(thinkfsm.Event{Kind: thinkfsm.Thinking, Text: part})
				content, thinking, calls := toolCallParserOutput(t, output)
				if content != "" || thinking != part || len(calls) != 0 {
					t.Errorf("Thinking Feed(%q) = (%q, %q, %+v), want immediate literal reasoning", part, content, thinking, calls)
				}
			}
			if events := parser.Finish(); len(events) != 0 {
				t.Errorf("Finish after literal reasoning = %+v, want empty", events)
			}
		})
	}
}

func TestToolCallParserGenericTextAndFinish(t *testing.T) {
	for _, kind := range []thinkfsm.EventKind{thinkfsm.Content, thinkfsm.Thinking} {
		parser := NewToolCallParser(nil)
		input := []thinkfsm.Event{
			{Kind: kind, Text: "before "},
			{Kind: kind, Text: `{"name":"f","arg`},
			{Kind: kind, Text: `uments":{"x":1}}`},
			{Kind: kind, Text: " after {unfinished"},
		}
		output := append(feedToolCallParser(t, parser, input), parser.Finish()...)
		content, thinking, calls := toolCallParserOutput(t, output)
		wantContent, wantThinking := "before  after {unfinished", ""
		wantCalls := []toolCallFn{{Name: "f", Arguments: `{"x":1}`}}
		if kind == thinkfsm.Thinking {
			wantContent, wantThinking = "", `before {"name":"f","arguments":{"x":1}} after {unfinished`
			wantCalls = nil
		}
		if content != wantContent || thinking != wantThinking || !callsEqual(calls, wantCalls) {
			t.Errorf("kind %v output = (%q, %q, %+v)", kind, content, thinking, calls)
		}
		if output := parser.Finish(); len(output) != 0 {
			t.Errorf("second Finish = %+v, want empty", output)
		}
	}
}

func TestToolCallParserGenericCallRecoveredAtFinish(t *testing.T) {
	const text = `<tool_call>{"name":"f","arguments":{}}`
	for _, kind := range []thinkfsm.EventKind{thinkfsm.Content, thinkfsm.Thinking} {
		t.Run("kind "+strconv.Itoa(int(kind)), func(t *testing.T) {
			parser := NewToolCallParser(nil)
			output := parser.Feed(thinkfsm.Event{Kind: kind, Text: text})
			if kind == thinkfsm.Content {
				if len(output) != 0 {
					t.Fatalf("unclosed Content wrapper emitted %+v before EOF", output)
				}
				content, thinking, calls := toolCallParserOutput(t, parser.Finish())
				if content != "" || thinking != "" || !callsEqual(calls, []toolCallFn{{Name: "f", Arguments: `{}`}}) {
					t.Errorf("Content EOF recovery = (%q, %q, %+v), want f call only", content, thinking, calls)
				}
			} else {
				content, thinking, calls := toolCallParserOutput(t, output)
				if content != "" || thinking != text || len(calls) != 0 {
					t.Errorf("Thinking wrapper = (%q, %q, %+v), want immediate literal reasoning", content, thinking, calls)
				}
			}
			if output := parser.Finish(); len(output) != 0 {
				t.Errorf("repeated Finish = %+v, want empty", output)
			}
		})
	}
}

func TestToolCallParserInstancesAreIndependent(t *testing.T) {
	a, b := NewToolCallParser(nil), NewToolCallParser(nil)
	input := directedToolCallEvents([]string{" to=functions.a"}, []string{`{"x":1}`}, false)
	if output := feedToolCallParser(t, a, input); len(output) != 0 {
		t.Fatalf("a before EOF = %+v", output)
	}
	b.Feed(thinkfsm.Event{Kind: thinkfsm.EnterThinking})
	bOutput := b.Feed(thinkfsm.Event{Kind: thinkfsm.Thinking, Text: "reason {"})
	b.Feed(thinkfsm.Event{Kind: thinkfsm.ToolCallEnd})
	_, _, calls := toolCallParserOutput(t, a.Finish())
	if !callsEqual(calls, []toolCallFn{{Name: "a", Arguments: `{"x":1}`}}) {
		t.Errorf("a calls = %+v", calls)
	}
	content, thinking, calls := toolCallParserOutput(t, append(bOutput, b.Finish()...))
	if content != "" || thinking != "reason {" || len(calls) != 0 {
		t.Errorf("b output = (%q, %q, %+v), want independent thinking", content, thinking, calls)
	}
}

func TestToolCallParserGPTOSSAbandonedRecipientDoesNotSuppressFreshCall(t *testing.T) {
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
			{"unmatched then empty body", []string{"x", "<|message|>", "<|end|>"}},
			{"empty ordinary body", []string{"<|message|>", "<|end|>"}},
		} {
			t.Run(prefix.name+"/"+exit.name, func(t *testing.T) {
				parser, splitter := NewToolCallParser(nil), thinkfsm.New()
				tokens := append(append([]string(nil), prefix.tokens...), exit.tokens...)
				tokens = append(tokens, "<|start|>", "assistant", " to", "=", "functions", ".f", "<|channel|>", "analysis", "<|message|>", "{}", "<|call|>")
				var output []ToolCallEvent
				for _, token := range tokens {
					ev := splitter.Feed(token)
					switch ev.Kind {
					case thinkfsm.EnterThinking, thinkfsm.LeaveThinking, thinkfsm.ToolCallStart, thinkfsm.ToolCallEnd:
						if ev.Text != "" {
							t.Errorf("token %q: boundary = %+v, want empty text", token, ev)
						}
					}
					output = append(output, parser.Feed(ev)...)
				}
				output = append(output, parser.Finish()...)
				content, thinking, calls := toolCallParserOutput(t, output)
				if content != "" || thinking != "" || !callsEqual(calls, []toolCallFn{{Name: "f", Arguments: "{}"}}) {
					t.Errorf("output = (%q, %q, %+v), want one fresh f call with {} and no display text", content, thinking, calls)
				}
				if events := parser.Finish(); len(events) != 0 {
					t.Errorf("second Finish = %+v, want empty", events)
				}
			})
		}
	}
}
