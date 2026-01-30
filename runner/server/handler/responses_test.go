// Copyright 2024-2026 Nexa AI, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package handler

import (
	"encoding/json"
	"fmt"
	"testing"

	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

func TestResponsesInputMessage_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    ResponsesInputMessage
		wantErr bool
	}{
		{
			name: "text content",
			json: `{"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}]}`,
			want: ResponsesInputMessage{
				Type:    "message",
				Role:    "user",
				Content: []ResponsesContent{ResponsesTextContent{Type: "input_text", Text: "hello"}},
			},
		},
		{
			name: "multiple content items",
			json: `{"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}, {"type": "input_text", "text": "world"}]}`,
			want: ResponsesInputMessage{
				Type: "message",
				Role: "user",
				Content: []ResponsesContent{
					ResponsesTextContent{Type: "input_text", Text: "hello"},
					ResponsesTextContent{Type: "input_text", Text: "world"},
				},
			},
		},
		{
			name:    "unknown content type",
			json:    `{"type": "message", "role": "user", "content": [{"type": "unknown"}]}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ResponsesInputMessage
			err := json.Unmarshal([]byte(tt.json), &got)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %q, want %q", got.Type, tt.want.Type)
			}
			if got.Role != tt.want.Role {
				t.Errorf("Role = %q, want %q", got.Role, tt.want.Role)
			}
			if len(got.Content) != len(tt.want.Content) {
				t.Fatalf("len(Content) = %d, want %d", len(got.Content), len(tt.want.Content))
			}
			for i := range tt.want.Content {
				switch wantContent := tt.want.Content[i].(type) {
				case ResponsesTextContent:
					gotContent, ok := got.Content[i].(ResponsesTextContent)
					if !ok {
						t.Fatalf("Content[%d] type = %T, want ResponsesTextContent", i, got.Content[i])
					}
					if gotContent != wantContent {
						t.Errorf("Content[%d] = %+v, want %+v", i, gotContent, wantContent)
					}
				}
			}
		})
	}
}

func TestResponsesInput_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		wantText  string
		wantItems int
		wantErr   bool
	}{
		{
			name:     "plain string",
			json:     `"hello world"`,
			wantText: "hello world",
		},
		{
			name:      "array with one message",
			json:      `[{"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}]}]`,
			wantItems: 1,
		},
		{
			name:      "array with multiple messages",
			json:      `[{"type": "message", "role": "system", "content": [{"type": "input_text", "text": "you are helpful"}]}, {"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}]}]`,
			wantItems: 2,
		},
		{
			name:     "null input",
			json:     `null`,
			wantText: "",
		},
		{
			name:      "single message object",
			json:      `{"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}]}`,
			wantItems: 1,
		},
		{
			name:    "invalid input",
			json:    `123`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got ResponsesInput
			err := json.Unmarshal([]byte(tt.json), &got)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Text != tt.wantText {
				t.Errorf("Text = %q, want %q", got.Text, tt.wantText)
			}
			if len(got.Items) != tt.wantItems {
				t.Errorf("len(Items) = %d, want %d", len(got.Items), tt.wantItems)
			}
		})
	}
}

func TestUnmarshalResponsesInputItem(t *testing.T) {
	t.Run("message item", func(t *testing.T) {
		got, err := unmarshalResponsesInputItem([]byte(`{"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}]}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		msg, ok := got.(ResponsesInputMessage)
		if !ok {
			t.Fatalf("got type %T, want ResponsesInputMessage", got)
		}
		if msg.Role != "user" {
			t.Errorf("Role = %q, want %q", msg.Role, "user")
		}
	})
	t.Run("function_call item", func(t *testing.T) {
		got, err := unmarshalResponsesInputItem([]byte(`{"type": "function_call", "call_id": "call_abc123", "name": "get_weather", "arguments": "{\"city\":\"Paris\"}"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		fc, ok := got.(ResponsesFunctionCall)
		if !ok {
			t.Fatalf("got type %T, want ResponsesFunctionCall", got)
		}
		if fc.CallID != "call_abc123" {
			t.Errorf("CallID = %q, want %q", fc.CallID, "call_abc123")
		}
		if fc.Name != "get_weather" {
			t.Errorf("Name = %q, want %q", fc.Name, "get_weather")
		}
	})
	t.Run("function_call_output item", func(t *testing.T) {
		got, err := unmarshalResponsesInputItem([]byte(`{"type": "function_call_output", "call_id": "call_abc123", "output": "the result"}`))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		output, ok := got.(ResponsesFunctionCallOutput)
		if !ok {
			t.Fatalf("got type %T, want ResponsesFunctionCallOutput", got)
		}
		if output.CallID != "call_abc123" {
			t.Errorf("CallID = %q, want %q", output.CallID, "call_abc123")
		}
		if output.Output != "the result" {
			t.Errorf("Output = %q, want %q", output.Output, "the result")
		}
	})
	t.Run("unknown item type", func(t *testing.T) {
		_, err := unmarshalResponsesInputItem([]byte(`{"type": "unknown_type"}`))
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestFromResponsesRequest(t *testing.T) {
	t.Run("string input", func(t *testing.T) {
		req := ResponsesRequest{
			Model: "test",
			Input: ResponsesInput{Text: "hello"},
		}
		systemPrompt, messages, tools, opts, err := FromResponsesRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if systemPrompt != "" {
			t.Errorf("systemPrompt = %q, want empty", systemPrompt)
		}
		if len(messages) != 1 {
			t.Fatalf("len(messages) = %d, want 1", len(messages))
		}
		if messages[0].Role != nexa_sdk.LLMRoleUser {
			t.Errorf("messages[0].Role = %q, want user", messages[0].Role)
		}
		if messages[0].Content != "hello" {
			t.Errorf("messages[0].Content = %q, want hello", messages[0].Content)
		}
		if tools != "" {
			t.Errorf("tools = %q, want empty", tools)
		}
		if opts.MaxTokens <= 0 {
			t.Error("opts.MaxTokens should be positive")
		}
	})
	t.Run("instructions", func(t *testing.T) {
		req := ResponsesRequest{
			Model:        "test",
			Instructions: "You are helpful.",
			Input:        ResponsesInput{Text: "hi"},
		}
		systemPrompt, messages, _, _, err := FromResponsesRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if systemPrompt != "You are helpful." {
			t.Errorf("systemPrompt = %q, want You are helpful.", systemPrompt)
		}
		if len(messages) != 1 {
			t.Fatalf("len(messages) = %d, want 1", len(messages))
		}
		if messages[0].Content != "hi" {
			t.Errorf("messages[0].Content = %q, want hi", messages[0].Content)
		}
	})
	t.Run("array input with messages", func(t *testing.T) {
		reqJSON := `{"model": "test", "input": [{"type": "message", "role": "user", "content": [{"type": "input_text", "text": "hello"}]}]}`
		var req ResponsesRequest
		if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		_, messages, _, _, err := FromResponsesRequest(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(messages) != 1 {
			t.Fatalf("len(messages) = %d, want 1", len(messages))
		}
		msg, ok := req.Input.Items[0].(ResponsesInputMessage)
		if !ok {
			t.Fatalf("Input.Items[0] type = %T", req.Input.Items[0])
		}
		role, content := convertResponsesInputMessage(msg)
		if role != "user" || content != "hello" {
			t.Errorf("role=%q content=%q, want user hello", role, content)
		}
	})
}

func TestParseToolCallsAll(t *testing.T) {
	t.Run("full JSON object inside tool_call", func(t *testing.T) {
		fullText := "<think>\nplan\n</think>\n\n<tool_call>\n{\"name\": \"get_system_health\", \"arguments\": {}}\n</tool_call>"
		calls, err := parseToolCallsAll(fullText)
		if err != nil {
			t.Fatalf("parseToolCallsAll: %v", err)
		}
		if len(calls) != 1 {
			t.Fatalf("len(calls) = %d, want 1", len(calls))
		}
		if calls[0].Name != "get_system_health" {
			t.Errorf("Name = %q, want get_system_health", calls[0].Name)
		}
		if calls[0].Arguments != "{}" {
			t.Errorf("Arguments = %q, want {}", calls[0].Arguments)
		}
	})
}

func TestToResponse(t *testing.T) {
	req := ResponsesRequest{Model: "test"}
	profile := nexa_sdk.ProfileData{
		PromptTokens:    10,
		GeneratedTokens: 5,
	}
	t.Run("text output", func(t *testing.T) {
		resp := ToResponse("test", "resp_1", "msg_1", "Hello world", profile, req)
		if resp.Status != "completed" {
			t.Errorf("Status = %q, want completed", resp.Status)
		}
		if len(resp.Output) != 1 {
			t.Fatalf("len(Output) = %d, want 1", len(resp.Output))
		}
		if resp.Output[0].Type != "message" {
			t.Errorf("Output[0].Type = %q, want message", resp.Output[0].Type)
		}
		if len(resp.Output[0].Content) != 1 {
			t.Fatalf("len(Output[0].Content) = %d, want 1", len(resp.Output[0].Content))
		}
		if resp.Output[0].Content[0].Text != "Hello world" {
			t.Errorf("Content[0].Text = %q, want Hello world", resp.Output[0].Content[0].Text)
		}
		if resp.Usage == nil {
			t.Error("Usage should be set")
		} else {
			if resp.Usage.InputTokens != 10 {
				t.Errorf("Usage.InputTokens = %d, want 10", resp.Usage.InputTokens)
			}
			if resp.Usage.OutputTokens != 5 {
				t.Errorf("Usage.OutputTokens = %d, want 5", resp.Usage.OutputTokens)
			}
		}
	})
	t.Run("tool call output", func(t *testing.T) {
		fullText := "<think>\nplan\n</think>\n\n<tool_call>\n{\"name\": \"get_system_health\", \"arguments\": {}}\n</tool_call>"
		resp := ToResponse("test", "resp_1", "msg_1", fullText, profile, req)
		if resp.Status != "completed" {
			t.Errorf("Status = %q, want completed", resp.Status)
		}
		if len(resp.Output) < 1 {
			t.Fatalf("len(Output) = %d, want at least 1", len(resp.Output))
		}
		idx := 0
		if resp.Output[0].Type == "reasoning" {
			if len(resp.Output[0].Content) != 1 || resp.Output[0].Content[0].Type != "reasoning_text" || resp.Output[0].Content[0].Text != "plan" {
				t.Errorf("Output[0] reasoning: content = %+v", resp.Output[0].Content)
			}
			idx = 1
		}
		if len(resp.Output) <= idx {
			t.Fatalf("no function_call in output")
		}
		if resp.Output[idx].Type != "function_call" {
			t.Errorf("Output[%d].Type = %q, want function_call", idx, resp.Output[idx].Type)
		}
		if resp.Output[idx].Name != "get_system_health" {
			t.Errorf("Output[%d].Name = %q, want get_system_health", idx, resp.Output[idx].Name)
		}
		if resp.Output[idx].Arguments != "{}" {
			t.Errorf("Output[%d].Arguments = %q, want {}", idx, resp.Output[idx].Arguments)
		}
	})
}

func TestResponsesStreamConverter(t *testing.T) {
	conv := NewResponsesStreamConverter("resp_123", "msg_456", "test", ResponsesRequest{})
	events := conv.ProcessToken("Hello")
	if len(events) < 4 {
		t.Fatalf("expected at least 4 events on first token, got %d", len(events))
	}
	if events[0].Event != "response.created" {
		t.Errorf("events[0].Event = %q, want response.created", events[0].Event)
	}
	if events[1].Event != "response.in_progress" {
		t.Errorf("events[1].Event = %q, want response.in_progress", events[1].Event)
	}
	events2 := conv.ProcessToken(" World")
	if len(events2) != 1 {
		t.Fatalf("expected 1 event on subsequent token, got %d", len(events2))
	}
	if events2[0].Event != "response.output_text.delta" {
		t.Errorf("events2[0].Event = %q, want response.output_text.delta", events2[0].Event)
	}
	profile := nexa_sdk.ProfileData{PromptTokens: 1, GeneratedTokens: 2}
	doneEvents := conv.ProcessDone("Hello World", profile)
	var hasCompleted bool
	for _, ev := range doneEvents {
		if ev.Event == "response.completed" {
			hasCompleted = true
			break
		}
	}
	if !hasCompleted {
		t.Error("expected response.completed event")
	}
}

// responsesCompatibilityValidResponse checks output shape per compatibility-test testOutputData (apiType === "responses").
func responsesCompatibilityValidResponse(output []ResponsesOutputItem) bool {
	for _, item := range output {
		if item.Type != "reasoning" {
			continue
		}
		if !(len(item.Content) > 0) {
			return false
		}
		for _, c := range item.Content {
			if c.Type != "reasoning_text" || len(c.Text) == 0 {
				return false
			}
		}
		return true
	}
	return false
}

func responsesCompatibilityFindToolCall(output []ResponsesOutputItem, toolName string) (arguments string, found bool) {
	for _, item := range output {
		if item.Type == "function_call" && item.Name == toolName {
			return item.Arguments, true
		}
	}
	return "", false
}

func TestResponsesCompatibilityOutputShape(t *testing.T) {
	req := ResponsesRequest{Model: "test"}
	profile := nexa_sdk.ProfileData{PromptTokens: 10, GeneratedTokens: 5}
	fullText := "<think>\nplan\n</think>\n\n<tool_call>\n{\"name\": \"get_system_health\", \"arguments\": {}}\n</tool_call>"
	resp := ToResponse("test", "resp_1", "msg_1", fullText, profile, req)
	if !responsesCompatibilityValidResponse(resp.Output) {
		t.Errorf("validResponse: output must contain reasoning item with content[].type=reasoning_text and non-empty text; output=%+v", resp.Output)
	}
	args, ok := responsesCompatibilityFindToolCall(resp.Output, "get_system_health")
	if !ok {
		t.Errorf("output must contain function_call item with name get_system_health")
	}
	if args != "{}" {
		t.Errorf("arguments = %q, want {}", args)
	}
}

func TestResponsesCompatibilityProcessDone(t *testing.T) {
	conv := NewResponsesStreamConverter("resp_1", "msg_1", "test", ResponsesRequest{})
	profile := nexa_sdk.ProfileData{PromptTokens: 10, GeneratedTokens: 5}
	fullText := "<think>\nplan\n</think>\n\n<tool_call>\n{\"name\": \"get_system_health\", \"arguments\": {}}\n</tool_call>"
	events := conv.ProcessDone(fullText, profile)
	var completedData map[string]any
	for _, ev := range events {
		if ev.Event == "response.completed" {
			if m, ok := ev.Data.(map[string]any); ok {
				completedData = m
				break
			}
		}
	}
	if completedData == nil {
		t.Fatal("expected response.completed event with data")
	}
	resp, _ := completedData["response"].(map[string]any)
	output, _ := resp["output"].([]any)
	hasReasoning := false
	var toolCallArgs string
	for _, it := range output {
		item, _ := it.(map[string]any)
		if item["type"] == "reasoning" {
			switch content := item["content"].(type) {
			case []any:
				if len(content) > 0 {
					if part, _ := content[0].(map[string]any); part != nil {
						if part["type"] == "reasoning_text" {
							if s, _ := part["text"].(string); len(s) > 0 {
								hasReasoning = true
							}
						}
					}
				}
			case []map[string]any:
				if len(content) > 0 {
					part := content[0]
					if part["type"] == "reasoning_text" {
						if s, _ := part["text"].(string); len(s) > 0 {
							hasReasoning = true
						}
					}
				}
			}
		}
		if item["type"] == "function_call" && item["name"] == "get_system_health" {
			toolCallArgs, _ = item["arguments"].(string)
		}
	}
	if !hasReasoning {
		t.Error("validResponse: response.completed output must contain reasoning item with content[].type=reasoning_text and non-empty text")
	}
	if toolCallArgs != "{}" {
		t.Errorf("function_call arguments = %q, want {}", toolCallArgs)
	}
}

var responsesCompatibilityCases = []struct {
	toolName          string
	expectedArguments string
}{
	{"get_system_health", "{}"},
	{"get_system_health", "{}"},
	{"get_system_health", "{}"},
	{"markdown_to_html", `{"markdown":"# Title\n\nSome *italic* text."}`},
	{"markdown_to_html", `{"markdown":"## Docs"}`},
	{"markdown_to_html", `{"markdown":"- item 1\n- item 2"}`},
	{"markdown_to_html", `{"markdown":"**bold**"}`},
	{"markdown_to_html", `{"markdown":"> quote"}`},
	{"detect_language", `{"text":"Buenos días, ¿cómo estás?"}`},
	{"detect_language", `{"text":"Guten Morgen"}`},
	{"detect_language", `{"text":"こんにちは、お元気ですか？"}`},
	{"detect_language", `{"text":"Привет, как дела?"}`},
	{"detect_language", `{"text":"Bonjour tout le monde"}`},
	{"generate_chart", `{"data":[[1,2],[2,4],[3,9]],"chart_type":"line"}`},
	{"generate_chart", `{"data":[[1,10],[2,20],[3,30]],"chart_type":"bar","title":"Quarterly Sales"}`},
	{"generate_chart", `{"data":[[0,1],[1,1.5],[2,2.2]],"chart_type":"scatter","title":"Experiment","x_label":"Time","y_label":"Value"}`},
	{"generate_chart", `{"data":[[1,70],[2,72],[3,68],[4,65]],"chart_type":"line","x_label":"Day"}`},
	{"generate_chart", `{"data":[[1,100],[2,150],[3,120]],"chart_type":"bar","title":"Daily Visits","y_label":"Visitors"}`},
	{"query_database", `{"table":"users","columns":["id","email"],"limit":5}`},
	{"query_database", `{"table":"orders","columns":["order_id","amount"],"filters":"status = 'shipped'"}`},
	{"query_database", `{"table":"products","columns":["name","price"],"limit":10,"order_by":"price DESC"}`},
	{"query_database", `{"table":"audit_log","columns":["id","timestamp","action"],"limit":3}`},
	{"query_database", `{"table":"customers","columns":["name","city"],"filters":"city = 'Berlin'"}`},
	{"get_weather", `{"location":"San Francisco"}`},
	{"get_weather", `{"location":"Tokyo"}`},
	{"get_weather", `{"location":"10001"}`},
	{"get_weather", `{"location":"Paris"}`},
	{"get_weather", `{"location":"Sydney"}`},
}

func TestResponsesCompatibilityToolCalls(t *testing.T) {
	req := ResponsesRequest{Model: "test"}
	profile := nexa_sdk.ProfileData{PromptTokens: 10, GeneratedTokens: 5}
	for i, tc := range responsesCompatibilityCases {
		t.Run(fmt.Sprintf("%s_%d", tc.toolName, i), func(t *testing.T) {
			fullText := "<think>\nplan\n</think>\n\n<tool_call>\n{\"name\": \"" + tc.toolName + "\", \"arguments\": " + tc.expectedArguments + "}\n</tool_call>"
			resp := ToResponse("test", "resp_1", "msg_1", fullText, profile, req)
			if !responsesCompatibilityValidResponse(resp.Output) {
				t.Errorf("validResponse: need reasoning item with content[].type=reasoning_text")
			}
			args, ok := responsesCompatibilityFindToolCall(resp.Output, tc.toolName)
			if !ok {
				t.Errorf("output must contain function_call with name %q", tc.toolName)
				return
			}
			var got, want map[string]any
			if err := json.Unmarshal([]byte(args), &got); err != nil {
				t.Errorf("arguments not valid JSON: %v", err)
				return
			}
			if err := json.Unmarshal([]byte(tc.expectedArguments), &want); err != nil {
				t.Fatalf("expected_arguments invalid: %v", err)
			}
			if !responsesDeepEqual(got, want) {
				t.Errorf("arguments = %s, want %s", args, tc.expectedArguments)
			}
		})
	}
}

func responsesDeepEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, av := range a {
		bv, ok := b[k]
		if !ok {
			return false
		}
		if !responsesValueEqual(av, bv) {
			return false
		}
	}
	return true
}

func responsesValueEqual(a, b any) bool {
	switch av := a.(type) {
	case map[string]any:
		bv, ok := b.(map[string]any)
		if !ok {
			return false
		}
		return responsesDeepEqual(av, bv)
	case []any:
		bv, ok := b.([]any)
		if !ok || len(av) != len(bv) {
			return false
		}
		for i := range av {
			if !responsesValueEqual(av[i], bv[i]) {
				return false
			}
		}
		return true
	case float64:
		if bv, ok := b.(float64); ok {
			return av == bv
		}
		if bv, ok := b.(int); ok {
			return av == float64(bv)
		}
		return false
	default:
		return a == b
	}
}
