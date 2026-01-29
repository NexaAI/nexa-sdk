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
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
)

type ResponsesContent interface {
	responsesContent()
}

type ResponsesTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (ResponsesTextContent) responsesContent() {}

type ResponsesImageContent struct {
	Type     string `json:"type"`
	Detail   string `json:"detail"`
	FileID   string `json:"file_id,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

func (ResponsesImageContent) responsesContent() {}

type ResponsesOutputTextContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (ResponsesOutputTextContent) responsesContent() {}

type ResponsesInputMessage struct {
	Type    string             `json:"type"`
	Role    string             `json:"role"`
	Content []ResponsesContent `json:"content,omitempty"`
}

func (m *ResponsesInputMessage) UnmarshalJSON(data []byte) error {
	var aux struct {
		Type    string          `json:"type"`
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	m.Type = aux.Type
	m.Role = aux.Role
	if len(aux.Content) == 0 {
		return nil
	}
	var contentStr string
	if err := json.Unmarshal(aux.Content, &contentStr); err == nil {
		m.Content = []ResponsesContent{
			ResponsesTextContent{Type: "input_text", Text: contentStr},
		}
		return nil
	}
	var rawItems []json.RawMessage
	if err := json.Unmarshal(aux.Content, &rawItems); err != nil {
		return fmt.Errorf("content must be a string or array: %w", err)
	}
	m.Content = make([]ResponsesContent, 0, len(rawItems))
	for i, raw := range rawItems {
		var typeField struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &typeField); err != nil {
			return fmt.Errorf("content[%d]: %w", i, err)
		}
		switch typeField.Type {
		case "input_text":
			var c ResponsesTextContent
			if err := json.Unmarshal(raw, &c); err != nil {
				return fmt.Errorf("content[%d]: %w", i, err)
			}
			m.Content = append(m.Content, c)
		case "input_image":
			var c ResponsesImageContent
			if err := json.Unmarshal(raw, &c); err != nil {
				return fmt.Errorf("content[%d]: %w", i, err)
			}
			m.Content = append(m.Content, c)
		case "output_text":
			var c ResponsesOutputTextContent
			if err := json.Unmarshal(raw, &c); err != nil {
				return fmt.Errorf("content[%d]: %w", i, err)
			}
			m.Content = append(m.Content, c)
		default:
			return fmt.Errorf("content[%d]: unknown content type: %s", i, typeField.Type)
		}
	}
	return nil
}

type ResponsesInputItem interface {
	responsesInputItem()
}

func (ResponsesInputMessage) responsesInputItem() {}

type ResponsesFunctionCall struct {
	ID        string `json:"id,omitempty"`
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (ResponsesFunctionCall) responsesInputItem() {}

type ResponsesFunctionCallOutput struct {
	Type   string `json:"type"`
	CallID string `json:"call_id"`
	Output string `json:"output"`
}

func (ResponsesFunctionCallOutput) responsesInputItem() {}

type ResponsesReasoningSummary struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ResponsesReasoningInput struct {
	ID               string                      `json:"id,omitempty"`
	Type             string                      `json:"type"`
	Summary          []ResponsesReasoningSummary `json:"summary,omitempty"`
	EncryptedContent string                      `json:"encrypted_content,omitempty"`
}

func (ResponsesReasoningInput) responsesInputItem() {}

func unmarshalResponsesInputItem(data []byte) (ResponsesInputItem, error) {
	var typeField struct {
		Type string `json:"type"`
		Role string `json:"role"`
	}
	if err := json.Unmarshal(data, &typeField); err != nil {
		return nil, err
	}
	itemType := typeField.Type
	if itemType == "" && typeField.Role != "" {
		itemType = "message"
	}
	switch itemType {
	case "message":
		var msg ResponsesInputMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return nil, err
		}
		return msg, nil
	case "function_call":
		var fc ResponsesFunctionCall
		if err := json.Unmarshal(data, &fc); err != nil {
			return nil, err
		}
		return fc, nil
	case "function_call_output":
		var out ResponsesFunctionCallOutput
		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}
		return out, nil
	case "reasoning":
		var r ResponsesReasoningInput
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, err
		}
		return r, nil
	default:
		return nil, fmt.Errorf("unknown input item type: %s", typeField.Type)
	}
}

type ResponsesInput struct {
	Text  string
	Items []ResponsesInputItem
}

func (r *ResponsesInput) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		r.Text = s
		return nil
	}
	var rawItems []json.RawMessage
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return fmt.Errorf("input must be a string or array: %w", err)
	}
	r.Items = make([]ResponsesInputItem, 0, len(rawItems))
	for i, raw := range rawItems {
		item, err := unmarshalResponsesInputItem(raw)
		if err != nil {
			return fmt.Errorf("input[%d]: %w", i, err)
		}
		r.Items = append(r.Items, item)
	}
	return nil
}

type ResponsesReasoning struct {
	Effort          string `json:"effort,omitempty"`
	GenerateSummary string `json:"generate_summary,omitempty"`
	Summary         string `json:"summary,omitempty"`
}

type ResponsesTextFormat struct {
	Type   string          `json:"type"`
	Name   string          `json:"name,omitempty"`
	Schema json.RawMessage `json:"schema,omitempty"`
	Strict *bool           `json:"strict,omitempty"`
}

type ResponsesText struct {
	Format *ResponsesTextFormat `json:"format,omitempty"`
}

type ResponsesTool struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description *string        `json:"description"`
	Strict      *bool           `json:"strict"`
	Parameters  map[string]any `json:"parameters"`
}

type ResponsesRequest struct {
	Model         string          `json:"model"`
	Background    bool            `json:"background"`
	Conversation  json.RawMessage `json:"conversation"`
	Include       []string        `json:"include"`
	Input         ResponsesInput   `json:"input"`
	Instructions  string          `json:"instructions,omitempty"`
	MaxOutputTokens *int          `json:"max_output_tokens,omitempty"`
	Reasoning     ResponsesReasoning `json:"reasoning"`
	Temperature   *float64       `json:"temperature"`
	Text          *ResponsesText  `json:"text,omitempty"`
	TopP          *float64       `json:"top_p"`
	Truncation    *string         `json:"truncation"`
	Tools         []ResponsesTool `json:"tools,omitempty"`
	Stream        *bool           `json:"stream,omitempty"`
}

type ResponsesHandlerOpts struct {
	Stream       bool
	SamplerConfig *nexa_sdk.SamplerConfig
	MaxTokens    int32
	EnableThink  bool
}

func responsesToolsToChatTools(tools []ResponsesTool) (string, error) {
	if len(tools) == 0 {
		return "", nil
	}
	type toolEntry struct {
		Type     string         `json:"type"`
		Function map[string]any `json:"function"`
	}
	entries := make([]toolEntry, 0, len(tools))
	for _, t := range tools {
		desc := ""
		if t.Description != nil {
			desc = *t.Description
		}
		params := t.Parameters
		if params == nil {
			params = map[string]any{"type": "object"}
		}
		entries = append(entries, toolEntry{
			Type: t.Type,
			Function: map[string]any{
				"name":        t.Name,
				"description": desc,
				"parameters":  params,
			},
		})
	}
	return sonic.MarshalString(entries)
}

func parseSamplerConfigFromResponses(r ResponsesRequest) *nexa_sdk.SamplerConfig {
	temp := 1.0
	if r.Temperature != nil {
		temp = *r.Temperature
	}
	topP := float32(1.0)
	if r.TopP != nil {
		topP = float32(*r.TopP)
	}
	return &nexa_sdk.SamplerConfig{
		Temperature: float32(temp),
		TopP:        topP,
		TopK:        0,
		MinP:        0,
		RepetitionPenalty: 1.0,
		PresencePenalty:   0,
		FrequencyPenalty:  0,
		Seed:              0,
		EnableJson:        false,
	}
}

func convertResponsesInputMessage(m ResponsesInputMessage) (role string, content string) {
	for _, c := range m.Content {
		switch v := c.(type) {
		case ResponsesTextContent:
			content += v.Text
		case ResponsesOutputTextContent:
			content += v.Text
		case ResponsesImageContent:
			// LLM path does not support image; skip
		}
	}
	return m.Role, content
}

func FromResponsesRequest(r ResponsesRequest) (systemPrompt string, messages []nexa_sdk.LlmChatMessage, tools string, opts ResponsesHandlerOpts, err error) {
	if r.Instructions != "" {
		systemPrompt = r.Instructions
	}
	if r.Input.Text != "" {
		messages = append(messages, nexa_sdk.LlmChatMessage{
			Role:    nexa_sdk.LLMRoleUser,
			Content: r.Input.Text,
		})
	}
	var pendingThinking string
	for _, item := range r.Input.Items {
		switch v := item.(type) {
		case ResponsesReasoningInput:
			pendingThinking = v.EncryptedContent
		case ResponsesInputMessage:
			role, content := convertResponsesInputMessage(v)
			msg := nexa_sdk.LlmChatMessage{
				Role:    nexa_sdk.LLMRole(role),
				Content: content,
			}
			messages = append(messages, msg)
			pendingThinking = ""
		case ResponsesFunctionCall:
			var argsAny any
			if v.Arguments != "" {
				_ = json.Unmarshal([]byte(v.Arguments), &argsAny)
			}
			inner, _ := json.Marshal(map[string]any{"name": v.Name, "arguments": argsAny})
			toolCallContent := "<tool_call>\n" + string(inner) + "\n</tool_call>"
			last := len(messages) - 1
			if last >= 0 && messages[last].Role == nexa_sdk.LLMRoleAssistant {
				messages[last].Content += "\n" + toolCallContent
			} else {
				messages = append(messages, nexa_sdk.LlmChatMessage{
					Role:    nexa_sdk.LLMRoleAssistant,
					Content: toolCallContent,
				})
			}
			pendingThinking = ""
		case ResponsesFunctionCallOutput:
			messages = append(messages, nexa_sdk.LlmChatMessage{
				Role:    nexa_sdk.LLMRoleUser,
				Content: v.Output,
			})
		}
	}
	if pendingThinking != "" {
		messages = append(messages, nexa_sdk.LlmChatMessage{
			Role:    nexa_sdk.LLMRoleAssistant,
			Content: pendingThinking,
		})
	}
	tools, err = responsesToolsToChatTools(r.Tools)
	if err != nil {
		return "", nil, "", ResponsesHandlerOpts{}, err
	}
	opts = ResponsesHandlerOpts{
		Stream:        r.Stream != nil && *r.Stream,
		SamplerConfig: parseSamplerConfigFromResponses(r),
		MaxTokens:     2048,
		EnableThink:   true,
	}
	if r.MaxOutputTokens != nil {
		opts.MaxTokens = int32(*r.MaxOutputTokens)
	}
	return systemPrompt, messages, tools, opts, nil
}

type ResponsesTextField struct {
	Format ResponsesTextFormat `json:"format"`
}

type ResponsesReasoningOutput struct {
	Effort  *string `json:"effort,omitempty"`
	Summary *string `json:"summary,omitempty"`
}

type ResponsesIncompleteDetails struct {
	Reason string `json:"reason"`
}

type ResponsesOutputContent struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Annotations []any  `json:"annotations"`
	Logprobs    []any  `json:"logprobs"`
}

type ResponsesOutputItem struct {
	Type             string                       `json:"type"`
	ID               string                       `json:"id,omitempty"`
	CallID           string                       `json:"call_id,omitempty"`
	Name             string                       `json:"name,omitempty"`
	Arguments        string                       `json:"arguments,omitempty"`
	Status           string                       `json:"status,omitempty"`
	Role             string                       `json:"role,omitempty"`
	Content          []ResponsesOutputContent     `json:"content,omitempty"`
	Summary          []ResponsesReasoningSummary  `json:"summary,omitempty"`
	EncryptedContent string                       `json:"encrypted_content,omitempty"`
}

type ResponsesInputTokensDetails struct {
	CachedTokens int `json:"cached_tokens"`
}

type ResponsesOutputTokensDetails struct {
	ReasoningTokens int `json:"reasoning_tokens"`
}

type ResponsesUsage struct {
	InputTokens         int                          `json:"input_tokens"`
	OutputTokens        int                          `json:"output_tokens"`
	TotalTokens         int                          `json:"total_tokens"`
	InputTokensDetails  ResponsesInputTokensDetails  `json:"input_tokens_details"`
	OutputTokensDetails ResponsesOutputTokensDetails `json:"output_tokens_details"`
}

type ResponsesError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ResponsesResponse struct {
	ID                 string                       `json:"id"`
	Object             string                       `json:"object"`
	CreatedAt          int64                        `json:"created_at"`
	Status             string                       `json:"status"`
	CompletedAt        *int64                       `json:"completed_at"`
	Error              *ResponsesError              `json:"error"`
	IncompleteDetails  *ResponsesIncompleteDetails  `json:"incomplete_details"`
	Instructions       *string                      `json:"instructions"`
	MaxOutputTokens    *int                         `json:"max_output_tokens"`
	Model              string                       `json:"model"`
	Output             []ResponsesOutputItem        `json:"output"`
	ParallelToolCalls  bool                         `json:"parallel_tool_calls"`
	PreviousResponseID *string                      `json:"previous_response_id"`
	Reasoning          *ResponsesReasoningOutput    `json:"reasoning"`
	Store              bool                         `json:"store"`
	Temperature        float64                      `json:"temperature"`
	Text               ResponsesTextField           `json:"text"`
	ToolChoice         any                         `json:"tool_choice"`
	Tools              []ResponsesTool              `json:"tools"`
	TopP               float64                      `json:"top_p"`
	Truncation         string                       `json:"truncation"`
	Usage              *ResponsesUsage              `json:"usage"`
	Metadata           map[string]any               `json:"metadata"`
}

func derefFloat64(p *float64, def float64) float64 {
	if p != nil {
		return *p
	}
	return def
}

const callIDChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomCallID() string {
	b := make([]byte, 24)
	for i := range b {
		b[i] = callIDChars[rand.IntN(len(callIDChars))]
	}
	return "call_" + string(b)
}

func profileToResponsesUsage(p nexa_sdk.ProfileData) *ResponsesUsage {
	return &ResponsesUsage{
		InputTokens:         int(p.PromptTokens),
		OutputTokens:        int(p.GeneratedTokens),
		TotalTokens:         int(p.TotalTokens()),
		InputTokensDetails:  ResponsesInputTokensDetails{CachedTokens: 0},
		OutputTokensDetails: ResponsesOutputTokensDetails{ReasoningTokens: 0},
	}
}

func ToResponse(model, responseID, itemID, fullText string, profile nexa_sdk.ProfileData, req ResponsesRequest) ResponsesResponse {
	var output []ResponsesOutputItem
	toolCalls, err := parseToolCallsAll(fullText)
	if err == nil {
		for i, tc := range toolCalls {
			output = append(output, ResponsesOutputItem{
				Type:      "function_call",
				ID:        fmt.Sprintf("fc_%s_%d", responseID, i),
				CallID:    randomCallID(),
				Name:      tc.Name,
				Arguments: tc.Arguments,
				Status:    "completed",
			})
		}
	} else {
		output = append(output, ResponsesOutputItem{
			Type:    "message",
			ID:      itemID,
			Status:  "completed",
			Role:    "assistant",
			Content: []ResponsesOutputContent{
				{Type: "output_text", Text: fullText, Annotations: []any{}, Logprobs: []any{}},
			},
		})
	}
	truncation := "disabled"
	if req.Truncation != nil {
		truncation = *req.Truncation
	}
	tools := req.Tools
	if tools == nil {
		tools = []ResponsesTool{}
	}
	text := ResponsesTextField{Format: ResponsesTextFormat{Type: "text"}}
	if req.Text != nil && req.Text.Format != nil {
		text.Format = *req.Text.Format
	}
	var reasoning *ResponsesReasoningOutput
	if req.Reasoning.Effort != "" || req.Reasoning.Summary != "" {
		reasoning = &ResponsesReasoningOutput{}
		if req.Reasoning.Effort != "" {
			reasoning.Effort = &req.Reasoning.Effort
		}
		if req.Reasoning.Summary != "" {
			reasoning.Summary = &req.Reasoning.Summary
		}
	}
	var instructions *string
	if req.Instructions != "" {
		instructions = &req.Instructions
	}
	createdAt := time.Now().Unix()
	return ResponsesResponse{
		ID:                 responseID,
		Object:             "response",
		CreatedAt:          createdAt,
		Status:             "completed",
		CompletedAt:        &createdAt,
		Instructions:       instructions,
		MaxOutputTokens:    req.MaxOutputTokens,
		Model:              model,
		Output:             output,
		ParallelToolCalls:  true,
		Reasoning:          reasoning,
		Store:              true,
		Temperature:        derefFloat64(req.Temperature, 1.0),
		Text:               text,
		ToolChoice:         "auto",
		Tools:              tools,
		TopP:               derefFloat64(req.TopP, 1.0),
		Truncation:         truncation,
		Usage:              profileToResponsesUsage(profile),
		Metadata:          map[string]any{},
	}
}

type ResponsesStreamEvent struct {
	Event string
	Data  any
}

type ResponsesStreamConverter struct {
	responseID      string
	itemID          string
	model           string
	request         ResponsesRequest
	firstWrite      bool
	outputIndex     int
	contentIndex    int
	contentStarted  bool
	accumulatedText string
	sequenceNumber  int
}

func NewResponsesStreamConverter(responseID, itemID, model string, request ResponsesRequest) *ResponsesStreamConverter {
	return &ResponsesStreamConverter{
		responseID:  responseID,
		itemID:      itemID,
		model:       model,
		request:     request,
		firstWrite:  true,
	}
}

func (c *ResponsesStreamConverter) newEvent(eventType string, data map[string]any) ResponsesStreamEvent {
	data["type"] = eventType
	data["sequence_number"] = c.sequenceNumber
	c.sequenceNumber++
	return ResponsesStreamEvent{Event: eventType, Data: data}
}

func (c *ResponsesStreamConverter) buildResponseObject(status string, output []any, usage map[string]any) map[string]any {
	truncation := "disabled"
	if c.request.Truncation != nil {
		truncation = *c.request.Truncation
	}
	topP := 1.0
	if c.request.TopP != nil {
		topP = *c.request.TopP
	}
	temp := 1.0
	if c.request.Temperature != nil {
		temp = *c.request.Temperature
	}
	return map[string]any{
		"id":                   c.responseID,
		"object":               "response",
		"created_at":           time.Now().Unix(),
		"completed_at":         nil,
		"status":               status,
		"model":                c.model,
		"output":               output,
		"tools":                c.request.Tools,
		"tool_choice":          "auto",
		"truncation":           truncation,
		"parallel_tool_calls":  true,
		"text":                 map[string]any{"format": map[string]any{"type": "text"}},
		"top_p":                topP,
		"temperature":          temp,
		"usage":                usage,
		"max_output_tokens":    c.request.MaxOutputTokens,
		"store":                false,
		"background":            c.request.Background,
		"service_tier":         "default",
		"metadata":             map[string]any{},
	}
}
func (c *ResponsesStreamConverter) ProcessToken(token string) []ResponsesStreamEvent {
	var events []ResponsesStreamEvent
	if c.firstWrite {
		c.firstWrite = false
		events = append(events, c.newEvent("response.created", map[string]any{
			"response": c.buildResponseObject("in_progress", []any{}, nil),
		}))
		events = append(events, c.newEvent("response.in_progress", map[string]any{
			"response": c.buildResponseObject("in_progress", []any{}, nil),
		}))
		c.contentStarted = true
		events = append(events, c.newEvent("response.output_item.added", map[string]any{
			"output_index": c.outputIndex,
			"item": map[string]any{
				"id": c.itemID, "type": "message", "status": "in_progress",
				"role": "assistant", "content": []any{},
			},
		}))
		events = append(events, c.newEvent("response.content_part.added", map[string]any{
			"item_id": c.itemID, "output_index": c.outputIndex, "content_index": c.contentIndex,
			"part": map[string]any{"type": "output_text", "text": "", "annotations": []any{}, "logprobs": []any{}},
		}))
	}
	c.accumulatedText += token
	events = append(events, c.newEvent("response.output_text.delta", map[string]any{
		"item_id": c.itemID, "output_index": c.outputIndex, "content_index": 0,
		"delta": token, "logprobs": []any{},
	}))
	return events
}

func (c *ResponsesStreamConverter) ProcessDone(fullText string, profile nexa_sdk.ProfileData) []ResponsesStreamEvent {
	var events []ResponsesStreamEvent
	usage := map[string]any{
		"input_tokens":          profile.PromptTokens,
		"output_tokens":         profile.GeneratedTokens,
		"total_tokens":          profile.TotalTokens(),
		"input_tokens_details":  map[string]any{"cached_tokens": 0},
		"output_tokens_details": map[string]any{"reasoning_tokens": 0},
	}
	toolCalls, err := parseToolCallsAll(fullText)
	var finalOutput []any
	if err == nil {
		for i, tc := range toolCalls {
			finalOutput = append(finalOutput, map[string]any{
				"type": "function_call",
				"id":   fmt.Sprintf("fc_%s_%d", c.responseID, i),
				"call_id": randomCallID(),
				"name": tc.Name,
				"arguments": tc.Arguments,
				"status": "completed",
			})
		}
	} else {
		finalOutput = []any{map[string]any{
			"id": c.itemID, "type": "message", "status": "completed", "role": "assistant",
			"content": []map[string]any{{
				"type": "output_text", "text": fullText,
				"annotations": []any{}, "logprobs": []any{},
			}},
		}}
	}
	if c.contentStarted {
		events = append(events, c.newEvent("response.output_text.done", map[string]any{
			"item_id": c.itemID, "output_index": c.outputIndex, "content_index": 0,
			"text": fullText, "logprobs": []any{},
		}))
		events = append(events, c.newEvent("response.content_part.done", map[string]any{
			"item_id": c.itemID, "output_index": c.outputIndex, "content_index": 0,
			"part": map[string]any{"type": "output_text", "text": fullText, "annotations": []any{}, "logprobs": []any{}},
		}))
		events = append(events, c.newEvent("response.output_item.done", map[string]any{
			"output_index": c.outputIndex,
			"item": map[string]any{
				"id": c.itemID, "type": "message", "status": "completed", "role": "assistant",
				"content": []map[string]any{{"type": "output_text", "text": fullText, "annotations": []any{}, "logprobs": []any{}}},
			},
		}))
	}
	resp := c.buildResponseObject("completed", finalOutput, usage)
	resp["completed_at"] = time.Now().Unix()
	events = append(events, c.newEvent("response.completed", map[string]any{"response": resp}))
	return events
}

func Responses(c *gin.Context) {
	var param ResponsesRequest
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	s := store.Get()
	name, _ := utils.NormalizeModelName(param.Model)
	manifest, err := s.GetManifest(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if manifest.ModelType != types.ModelTypeLLM {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model type not support"})
		return
	}
	systemPrompt, messages, tools, opts, err := FromResponsesRequest(param)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	p, err := service.KeepAliveGet[nexa_sdk.LLM](
		param.Model,
		types.ModelParam{NCtx: 4096, NGpuLayers: 999, SystemPrompt: systemPrompt},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]any{"error": "model not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}
	if len(messages) == 0 && systemPrompt == "" {
		c.JSON(http.StatusOK, nil)
		return
	}
	formatted, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
		Messages:            messages,
		Tools:                tools,
		EnableThink:         opts.EnableThink,
		AddGenerationPrompt: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}
	responseID := fmt.Sprintf("resp_%d", rand.Uint32())
	itemID := fmt.Sprintf("msg_%d", rand.Uint32())
	if !opts.Stream {
		genOut, err := p.Generate(nexa_sdk.LlmGenerateInput{
			PromptUTF8: formatted.FormattedText,
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens:     opts.MaxTokens,
				SamplerConfig: opts.SamplerConfig,
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
			return
		}
		resp := ToResponse(param.Model, responseID, itemID, genOut.FullText, genOut.ProfileData, param)
		c.JSON(http.StatusOK, resp)
		return
	}
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	conv := NewResponsesStreamConverter(responseID, itemID, param.Model, param)
	var (
		fullText   string
		profile    nexa_sdk.ProfileData
		genErr     error
		stopGen    bool
		dataCh     = make(chan string)
		resWg      sync.WaitGroup
	)
	resWg.Add(1)
	go func() {
		defer resWg.Done()
		var res nexa_sdk.LlmGenerateOutput
		res, genErr = p.Generate(nexa_sdk.LlmGenerateInput{
			PromptUTF8: formatted.FormattedText,
			OnToken: func(token string) bool {
				if stopGen {
					return false
				}
				dataCh <- token
				return true
			},
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens:     opts.MaxTokens,
				SamplerConfig: opts.SamplerConfig,
			},
		})
		fullText = res.FullText
		profile = res.ProfileData
		close(dataCh)
	}()
	for token := range dataCh {
		for _, ev := range conv.ProcessToken(token) {
			c.SSEvent(ev.Event, ev.Data)
			c.Writer.Flush()
		}
	}
	resWg.Wait()
	if genErr != nil {
		c.SSEvent("error", map[string]any{"error": genErr.Error(), "code": nexa_sdk.SDKErrorCode(genErr)})
		return
	}
	for _, ev := range conv.ProcessDone(fullText, profile) {
		c.SSEvent(ev.Event, ev.Data)
		c.Writer.Flush()
	}
}
