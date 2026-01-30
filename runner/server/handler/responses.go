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
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
	"github.com/openai/openai-go/v3/shared/constant"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
)

var responsesToolCallRegex = regexp.MustCompile(`<tool_call>([\s\S]+?)</tool_call>` + "|" + "```json([\\s\\S]+?)```")

type ResponseCreateRequest struct {
	responses.ResponseNewParams
	Stream bool `json:"stream"`
}

func responsesInputFromParamsLLM(body responses.ResponseNewParams) (systemPrompt string, messages []nexa_sdk.LlmChatMessage, err error) {
	if !param.IsOmitted(body.Instructions) {
		systemPrompt = body.Instructions.Value
	}
	input := body.Input
	if param.IsOmitted(input.OfString) && len(input.OfInputItemList) == 0 {
		return systemPrompt, nil, nil
	}
	if !param.IsOmitted(input.OfString) {
		messages = []nexa_sdk.LlmChatMessage{
			{Role: nexa_sdk.LLMRoleUser, Content: input.OfString.Value},
		}
		return systemPrompt, messages, nil
	}
	messages = make([]nexa_sdk.LlmChatMessage, 0, len(input.OfInputItemList))
	for _, item := range input.OfInputItemList {
		var role string
		var text string
		if item.OfMessage != nil {
			role = string(item.OfMessage.Role)
			if !param.IsOmitted(item.OfMessage.Content.OfString) {
				text = item.OfMessage.Content.OfString.Value
			} else if len(item.OfMessage.Content.OfInputItemContentList) > 0 {
				for _, c := range item.OfMessage.Content.OfInputItemContentList {
					if c.OfInputText != nil {
						text += c.OfInputText.Text
					}
				}
			}
		} else if item.OfInputMessage != nil {
			role = item.OfInputMessage.Role
			for _, c := range item.OfInputMessage.Content {
				if c.OfInputText != nil {
					text += c.OfInputText.Text
				}
			}
		} else {
			continue
		}
		if role == "" {
			role = string(openai.MessageRoleUser)
		}
		if role == "system" || role == "developer" {
			systemPrompt += text
			continue
		}
		messages = append(messages, nexa_sdk.LlmChatMessage{
			Role:    nexa_sdk.LLMRole(role),
			Content: text,
		})
	}
	return systemPrompt, messages, nil
}

func responsesToolsString(params responses.ResponseNewParams) (string, error) {
	if len(params.Tools) == 0 {
		return "", nil
	}
	chatTools := make([]map[string]any, 0, len(params.Tools))
	for _, t := range params.Tools {
		if t.OfFunction == nil {
			continue
		}
		fn := map[string]any{
			"name":       t.OfFunction.Name,
			"parameters": t.OfFunction.Parameters,
		}
		if !param.IsOmitted(t.OfFunction.Description) {
			fn["description"] = t.OfFunction.Description.Value
		}
		chatTools = append(chatTools, map[string]any{
			"type":     "function",
			"function": fn,
		})
	}
	if len(chatTools) == 0 {
		return "", nil
	}
	b, err := json.Marshal(chatTools)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func responsesInputFromParamsVLM(body responses.ResponseNewParams) (systemPrompt string, messages []nexa_sdk.VlmChatMessage, err error) {
	if !param.IsOmitted(body.Instructions) {
		systemPrompt = body.Instructions.Value
	}
	input := body.Input
	if param.IsOmitted(input.OfString) && len(input.OfInputItemList) == 0 {
		return systemPrompt, nil, nil
	}
	if !param.IsOmitted(input.OfString) {
		messages = []nexa_sdk.VlmChatMessage{
			{Role: nexa_sdk.VlmRole(string(openai.MessageRoleUser)), Contents: []nexa_sdk.VlmContent{{Type: nexa_sdk.VlmContentTypeText, Text: input.OfString.Value}}},
		}
		return systemPrompt, messages, nil
	}
	messages = make([]nexa_sdk.VlmChatMessage, 0, len(input.OfInputItemList))
	for _, item := range input.OfInputItemList {
		var role string
		var contents []nexa_sdk.VlmContent
		if item.OfMessage != nil {
			role = string(item.OfMessage.Role)
			if !param.IsOmitted(item.OfMessage.Content.OfString) {
				contents = []nexa_sdk.VlmContent{{Type: nexa_sdk.VlmContentTypeText, Text: item.OfMessage.Content.OfString.Value}}
			} else if len(item.OfMessage.Content.OfInputItemContentList) > 0 {
				contents = make([]nexa_sdk.VlmContent, 0, len(item.OfMessage.Content.OfInputItemContentList))
				for _, c := range item.OfMessage.Content.OfInputItemContentList {
					if c.OfInputText != nil {
						contents = append(contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeText, Text: c.OfInputText.Text})
					}
					if c.OfInputImage != nil && !param.IsOmitted(c.OfInputImage.ImageURL) {
						file, err := utils.SaveURIToTempFile(c.OfInputImage.ImageURL.Value)
						if err == nil {
							contents = append(contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeImage, Text: file})
						}
					}
				}
			}
		} else if item.OfInputMessage != nil {
			role = item.OfInputMessage.Role
			contents = make([]nexa_sdk.VlmContent, 0, len(item.OfInputMessage.Content))
			for _, c := range item.OfInputMessage.Content {
				if c.OfInputText != nil {
					contents = append(contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeText, Text: c.OfInputText.Text})
				}
				if c.OfInputImage != nil && !param.IsOmitted(c.OfInputImage.ImageURL) {
					file, err := utils.SaveURIToTempFile(c.OfInputImage.ImageURL.Value)
					if err == nil {
						contents = append(contents, nexa_sdk.VlmContent{Type: nexa_sdk.VlmContentTypeImage, Text: file})
					}
				}
			}
		} else {
			continue
		}
		if role == "" {
			role = string(openai.MessageRoleUser)
		}
		if role == "system" || role == "developer" {
			for _, ct := range contents {
				if ct.Type == nexa_sdk.VlmContentTypeText {
					systemPrompt += ct.Text
				}
			}
			continue
		}
		if len(contents) == 0 {
			continue
		}
		messages = append(messages, nexa_sdk.VlmChatMessage{
			Role:     nexa_sdk.VlmRole(role),
			Contents: contents,
		})
	}
	return systemPrompt, messages, nil
}

func responseUsageFromProfile(p nexa_sdk.ProfileData) responses.ResponseUsage {
	return responses.ResponseUsage{
		InputTokens:         p.PromptTokens,
		InputTokensDetails:  responses.ResponseUsageInputTokensDetails{CachedTokens: 0},
		OutputTokens:        p.GeneratedTokens,
		OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{ReasoningTokens: 0},
		TotalTokens:         p.TotalTokens(),
	}
}

type parsedToolCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func parseToolCallFromResponse(fullText string) (name, arguments string, ok bool) {
	match := responsesToolCallRegex.FindStringSubmatch(fullText)
	if len(match) < 2 {
		return "", "", false
	}
	content := match[1]
	if content == "" && len(match) > 2 {
		content = match[2]
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return "", "", false
	}
	jsonStr := content
	if content[0] != '{' {
		jsonStr = "{" + content + "}"
	}
	var parsed parsedToolCall
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return "", "", false
	}
	if parsed.Name == "" {
		return "", "", false
	}
	if len(parsed.Arguments) == 0 {
		arguments = "{}"
	} else if parsed.Arguments[0] == '"' {
		_ = json.Unmarshal(parsed.Arguments, &arguments)
	} else {
		arguments = string(parsed.Arguments)
	}
	return parsed.Name, arguments, true
}

func buildResponseOutput(fullText string, hasTools bool, respID string) []responses.ResponseOutputItemUnion {
	if hasTools {
		name, arguments, ok := parseToolCallFromResponse(fullText)
		if ok {
			suffix := respID
			if len(respID) > 5 {
				suffix = respID[5:]
			} else {
				suffix = fmt.Sprintf("%d", time.Now().UnixNano())
			}
			fcID := "fc_" + suffix
			callID := "call_" + fmt.Sprintf("%x", rand.Int63())
			return []responses.ResponseOutputItemUnion{
				{
					Type:      "function_call",
					ID:        fcID,
					CallID:    callID,
					Name:      name,
					Arguments: arguments,
					Status:    "completed",
				},
			}
		}
	}
	msgID := "msg_" + respID[5:]
	if len(respID) <= 5 {
		msgID = "msg_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return []responses.ResponseOutputItemUnion{
		{
			ID:      msgID,
			Type:    "message",
			Role:    constant.Assistant(openai.MessageRoleAssistant),
			Status:  "completed",
			Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: fullText, Annotations: nil}},
		},
	}
}

func buildResponse(id, model string, output []responses.ResponseOutputItemUnion, createdAt, completedAt float64, usage responses.ResponseUsage) responses.Response {
	return responses.Response{
		ID:                id,
		Object:            (constant.Response)("response"),
		CreatedAt:         createdAt,
		Status:            responses.ResponseStatusCompleted,
		CompletedAt:       completedAt,
		Error:             responses.ResponseError{},
		IncompleteDetails: responses.ResponseIncompleteDetails{},
		Model:             shared.ResponsesModel(model),
		Output:            output,
		ParallelToolCalls: true,
		Temperature:       1,
		ToolChoice:        responses.ResponseToolChoiceUnion{OfToolChoiceMode: responses.ToolChoiceOptionsAuto},
		Tools:             nil,
		TopP:              1,
		Usage:             usage,
		Metadata:          shared.Metadata{},
	}
}

func ResponsesCreate(c *gin.Context) {
	req := ResponseCreateRequest{}
	req.ResponseNewParams = responses.ResponseNewParams{
		MaxOutputTokens: param.NewOpt[int64](2048),
		Temperature:     param.NewOpt[float64](1),
		TopP:            param.NewOpt[float64](1),
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model is required"})
		return
	}

	maxTokens := int64(2048)
	if !param.IsOmitted(req.MaxOutputTokens) {
		maxTokens = req.MaxOutputTokens.Value
	}
	temp := float64(1)
	if !param.IsOmitted(req.Temperature) {
		temp = req.Temperature.Value
	}
	topP := float64(1)
	if !param.IsOmitted(req.TopP) {
		topP = req.TopP.Value
	}

	slog.Info("ResponsesCreate", "model", req.Model, "stream", req.Stream)

	s := store.Get()
	name, _ := utils.NormalizeModelName(string(req.Model))
	manifest, err := s.GetManifest(name)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	switch manifest.ModelType {
	case types.ModelTypeLLM:
		systemPrompt, messages, err := responsesInputFromParamsLLM(req.ResponseNewParams)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		responsesCreateLLM(c, req, systemPrompt, messages, maxTokens, temp, topP)
	case types.ModelTypeVLM:
		systemPrompt, vlmMessages, err := responsesInputFromParamsVLM(req.ResponseNewParams)
		if err != nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		responsesCreateVLM(c, req, systemPrompt, vlmMessages, maxTokens, temp, topP)
	default:
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model type not support"})
	}
}

func responsesCreateLLM(c *gin.Context, req ResponseCreateRequest, systemPrompt string, messages []nexa_sdk.LlmChatMessage, maxTokens int64, temp, topP float64) {
	p, err := service.KeepAliveGet[nexa_sdk.LLM](
		string(req.Model),
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

	if len(messages) == 0 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "input is required"})
		return
	}

	samplerConfig := &nexa_sdk.SamplerConfig{
		Temperature: float32(temp),
		TopP:        float32(topP),
	}

	toolsStr, err := responsesToolsString(req.ResponseNewParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	formatted, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
		Messages:            messages,
		Tools:               toolsStr,
		EnableThink:         true,
		AddGenerationPrompt: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	now := float64(time.Now().Unix())
	respID := fmt.Sprintf("resp_%x%x", time.Now().UnixNano(), rand.Int63())

	if req.Stream {
		responsesCreateLLMStream(c, respID, string(req.Model), now, p, formatted.FormattedText, samplerConfig, maxTokens, len(req.Tools) > 0)
		return
	}

	genOut, err := p.Generate(nexa_sdk.LlmGenerateInput{
		PromptUTF8: formatted.FormattedText,
		Config: &nexa_sdk.GenerationConfig{
			MaxTokens:     int32(maxTokens),
			SamplerConfig: samplerConfig,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	usage := responseUsageFromProfile(genOut.ProfileData)
	output := buildResponseOutput(genOut.FullText, len(req.Tools) > 0, respID)
	res := buildResponse(respID, string(req.Model), output, now, float64(time.Now().Unix()), usage)
	c.JSON(http.StatusOK, res)
}

func responsesCreateLLMStream(c *gin.Context, respID, model string, createdAt float64, p *nexa_sdk.LLM, prompt string, samplerConfig *nexa_sdk.SamplerConfig, maxTokens int64, hasTools bool) {
	usage := responses.ResponseUsage{
		InputTokens:         0,
		InputTokensDetails:  responses.ResponseUsageInputTokensDetails{CachedTokens: 0},
		OutputTokens:        0,
		OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{ReasoningTokens: 0},
		TotalTokens:         0,
	}
	initialResponse := buildResponse(respID, model, nil, createdAt, 0, usage)
	initialResponse.Status = responses.ResponseStatusInProgress
	initialResponse.CompletedAt = 0
	initialResponse.Output = nil
	initialResponse.Usage = responses.ResponseUsage{}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	seq := int64(1)
	emit := func(ev any) {
		c.SSEvent("", ev)
		seq++
	}
	sendEvent := func(typ string, data map[string]any) {
		event := map[string]any{"type": typ, "sequence_number": seq}
		for k, v := range data {
			event[k] = v
		}
		seq++
		c.SSEvent("", event)
	}

	emit(responses.ResponseCreatedEvent{
		Response:       initialResponse,
		SequenceNumber: seq,
		Type:           (constant.ResponseCreated)("response.created"),
	})

	msgID := "msg_" + respID[5:]
	if len(respID) <= 5 {
		msgID = "msg_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	emit(responses.ResponseOutputItemAddedEvent{
		Item:           responses.ResponseOutputItemUnion{ID: msgID, Status: "in_progress", Type: "message", Role: constant.Assistant(openai.MessageRoleAssistant), Content: nil},
		OutputIndex:    0,
		SequenceNumber: seq,
		Type:           (constant.ResponseOutputItemAdded)("response.output_item.added"),
	})
	emit(responses.ResponseContentPartAddedEvent{
		ItemID:         msgID,
		OutputIndex:    0,
		ContentIndex:   0,
		Part:           responses.ResponseContentPartAddedEventPartUnion{Type: "output_text", Text: "", Annotations: nil},
		SequenceNumber: seq,
		Type:           (constant.ResponseContentPartAdded)("response.content_part.added"),
	})

	stopGen := false
	var fullText string
	var profile nexa_sdk.ProfileData
	var genErr error
	var resWg sync.WaitGroup
	dataCh := make(chan string)

	resWg.Add(1)
	go func() {
		defer resWg.Done()
		out, err := p.Generate(nexa_sdk.LlmGenerateInput{
			PromptUTF8: prompt,
			OnToken: func(token string) bool {
				if stopGen {
					return false
				}
				dataCh <- token
				return true
			},
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens:     int32(maxTokens),
				SamplerConfig: samplerConfig,
			},
		})
		if err != nil {
			genErr = err
			close(dataCh)
			return
		}
		fullText = out.FullText
		profile = out.ProfileData
		close(dataCh)
	}()

	c.Stream(func(w io.Writer) bool {
		token, ok := <-dataCh
		if ok {
			sendEvent("response.output_text.delta", map[string]any{
				"item_id": msgID, "output_index": 0, "content_index": 0, "delta": token,
			})
			return true
		}
		resWg.Wait()
		if genErr != nil {
			sendEvent("error", map[string]any{"code": "server_error", "message": genErr.Error()})
			return false
		}

		sendEvent("response.output_text.done", map[string]any{
			"item_id": msgID, "output_index": 0, "content_index": 0, "text": fullText,
		})
		sendEvent("response.content_part.done", map[string]any{
			"item_id": msgID, "output_index": 0, "content_index": 0,
			"part": map[string]any{"type": "output_text", "text": fullText, "annotations": []any{}},
		})
		sendEvent("response.output_item.done", map[string]any{
			"output_index": 0,
			"item": responses.ResponseOutputItemUnion{
				ID: msgID, Status: "completed", Type: "message", Role: constant.Assistant(openai.MessageRoleAssistant),
				Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: fullText, Annotations: nil}},
			},
		})

		usage = responseUsageFromProfile(profile)
		output := buildResponseOutput(fullText, hasTools, respID)
		completedResp := buildResponse(respID, model, output, createdAt, float64(time.Now().Unix()), usage)
		sendEvent("response.completed", map[string]any{"response": completedResp})
		return false
	})

	stopGen = true
	for range dataCh {
	}
}

func responsesCreateVLM(c *gin.Context, req ResponseCreateRequest, systemPrompt string, vlmMessages []nexa_sdk.VlmChatMessage, maxTokens int64, temp, topP float64) {
	p, err := service.KeepAliveGet[nexa_sdk.VLM](
		string(req.Model),
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

	if len(vlmMessages) == 0 {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "input is required"})
		return
	}

	samplerConfig := &nexa_sdk.SamplerConfig{
		Temperature: float32(temp),
		TopP:        float32(topP),
	}

	toolsStr, err := responsesToolsString(req.ResponseNewParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	formatted, err := p.ApplyChatTemplate(nexa_sdk.VlmApplyChatTemplateInput{
		Messages:    vlmMessages,
		Tools:       toolsStr,
		EnableThink: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}
	images := make([]string, 0)
	audios := make([]string, 0)
	for _, content := range vlmMessages[len(vlmMessages)-1].Contents {
		switch content.Type {
		case nexa_sdk.VlmContentTypeImage:
			images = append(images, content.Text)
		case nexa_sdk.VlmContentTypeAudio:
			audios = append(audios, content.Text)
		}
	}

	now := float64(time.Now().Unix())
	respID := fmt.Sprintf("resp_%x%x", time.Now().UnixNano(), rand.Int63())

	if req.Stream {
		responsesCreateVLMStream(c, respID, string(req.Model), now, p, formatted.FormattedText, samplerConfig, maxTokens, images, audios, len(req.Tools) > 0)
		return
	}

	genOut, err := p.Generate(nexa_sdk.VlmGenerateInput{
		PromptUTF8: formatted.FormattedText,
		Config: &nexa_sdk.GenerationConfig{
			MaxTokens:     int32(maxTokens),
			SamplerConfig: samplerConfig,
			ImagePaths:    images,
			AudioPaths:    audios,
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	usage := responseUsageFromProfile(genOut.ProfileData)
	output := buildResponseOutput(genOut.FullText, len(req.Tools) > 0, respID)
	res := buildResponse(respID, string(req.Model), output, now, float64(time.Now().Unix()), usage)
	c.JSON(http.StatusOK, res)
}

func responsesCreateVLMStream(c *gin.Context, respID, model string, createdAt float64, p *nexa_sdk.VLM, prompt string, samplerConfig *nexa_sdk.SamplerConfig, maxTokens int64, imagePaths, audioPaths []string, hasTools bool) {
	usage := responses.ResponseUsage{
		InputTokens:         0,
		InputTokensDetails:  responses.ResponseUsageInputTokensDetails{CachedTokens: 0},
		OutputTokens:        0,
		OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{ReasoningTokens: 0},
		TotalTokens:         0,
	}
	initialResponse := buildResponse(respID, model, nil, createdAt, 0, usage)
	initialResponse.Status = responses.ResponseStatusInProgress
	initialResponse.CompletedAt = 0
	initialResponse.Output = nil
	initialResponse.Usage = responses.ResponseUsage{}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	seq := 1
	sendEvent := func(typ string, data map[string]any) {
		event := map[string]any{"type": typ, "sequence_number": seq}
		for k, val := range data {
			event[k] = val
		}
		seq++
		c.SSEvent("", event)
	}

	sendEvent("response.created", map[string]any{"response": initialResponse})
	msgID := "msg_" + respID[5:]
	if len(respID) <= 5 {
		msgID = "msg_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	sendEvent("response.output_item.added", map[string]any{
		"output_index": 0,
		"item":         map[string]any{"id": msgID, "status": "in_progress", "type": "message", "role": string(openai.MessageRoleAssistant), "content": []any{}},
	})
	sendEvent("response.content_part.added", map[string]any{
		"item_id": msgID, "output_index": 0, "content_index": 0,
		"part": map[string]any{"type": "output_text", "text": "", "annotations": []any{}},
	})

	stopGen := false
	var fullText string
	var profile nexa_sdk.ProfileData
	var genErr error
	var resWg sync.WaitGroup
	dataCh := make(chan string)

	resWg.Add(1)
	go func() {
		defer resWg.Done()
		out, err := p.Generate(nexa_sdk.VlmGenerateInput{
			PromptUTF8: prompt,
			OnToken: func(token string) bool {
				if stopGen {
					return false
				}
				dataCh <- token
				return true
			},
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens:     int32(maxTokens),
				SamplerConfig: samplerConfig,
				ImagePaths:    imagePaths,
				AudioPaths:    audioPaths,
			},
		})
		if err != nil {
			genErr = err
			close(dataCh)
			return
		}
		fullText = out.FullText
		profile = out.ProfileData
		close(dataCh)
	}()

	c.Stream(func(w io.Writer) bool {
		token, ok := <-dataCh
		if ok {
			sendEvent("response.output_text.delta", map[string]any{"item_id": msgID, "output_index": 0, "content_index": 0, "delta": token})
			return true
		}
		resWg.Wait()
		if genErr != nil {
			c.SSEvent("", map[string]any{"type": "error", "code": "server_error", "message": genErr.Error(), "sequence_number": seq})
			return false
		}
		sendEvent("response.output_text.done", map[string]any{"item_id": msgID, "output_index": 0, "content_index": 0, "text": fullText})
		sendEvent("response.content_part.done", map[string]any{
			"item_id": msgID, "output_index": 0, "content_index": 0,
			"part": map[string]any{"type": "output_text", "text": fullText, "annotations": []any{}},
		})
		sendEvent("response.output_item.done", map[string]any{
			"output_index": 0,
			"item": responses.ResponseOutputItemUnion{
				ID: msgID, Status: "completed", Type: "message", Role: constant.Assistant(openai.MessageRoleAssistant),
				Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: fullText, Annotations: nil}},
			},
		})
		usage = responseUsageFromProfile(profile)
		output := buildResponseOutput(fullText, hasTools, respID)
		completedResp := buildResponse(respID, model, output, createdAt, float64(time.Now().Unix()), usage)
		sendEvent("response.completed", map[string]any{"response": completedResp})
		return false
	})

	stopGen = true
	for range dataCh {
	}
}
