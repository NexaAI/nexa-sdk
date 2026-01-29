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
