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
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/ast"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared/constant"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
)

type ResponseNewParams responses.ResponseNewParams

var responsesToolCallRegex = regexp.MustCompile(`<tool_call>([\s\S]+?)</tool_call>` + "|" + "```json([\\s\\S]+?)```")

type ResponseCreateRequest struct {
	ResponseNewParams
	Stream bool `json:"stream"`

	EnableThink bool  `json:"enable_think"`
	NCtx        int32 `json:"nctx"`
	Ngl         int32 `json:"ngl"`

	ImageMaxLength int32 `json:"image_max_length"`

	TopK              int32   `json:"top_k"`
	MinP              float32 `json:"min_p"`
	RepetitionPenalty float32 `json:"repetition_penalty"`
	GrammarPath       string  `json:"grammar_path"`
	GrammarString     string  `json:"grammar_string"`
	EnableJson        bool    `json:"enable_json"`
}

func defaultResponseCreateRequest() ResponseCreateRequest {
	return ResponseCreateRequest{
		ResponseNewParams: ResponseNewParams{
			MaxOutputTokens: param.NewOpt[int64](2048),
		},
		Stream: false,

		EnableThink:       true,
		NCtx:              4096,
		Ngl:               999,
		ImageMaxLength:    512,
		TopK:              0,
		MinP:              0.0,
		RepetitionPenalty: 1.0,
		GrammarPath:       "",
		GrammarString:     "",
		EnableJson:        false,
	}
}

func Responses(c *gin.Context) {
	req := defaultResponseCreateRequest()
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Error("Failed to bind JSON", "error", err, "request", c.Request.Body)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	slog.Debug("DefaultResponseCreateRequest called", "req", req)

	// Automatically adjust NCtx if MaxCompletionTokens is larger
	if req.NCtx < int32(req.MaxOutputTokens.Value) {
		slog.Debug("Adjust NCtx to MaxOutputTokens", "from", req.NCtx, "to", req.MaxOutputTokens.Value)
		req.NCtx = int32(req.MaxOutputTokens.Value)
	}

	s := store.Get()
	name, _ := utils.NormalizeModelName(string(req.Model))
	manifest, err := s.GetManifest(name)
	if err != nil {
		slog.Error("Failed to get model manifest", "model", req.Model, "error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	switch manifest.ModelType {
	case types.ModelTypeLLM:
		responsesCreateLLM(c, req)
	case types.ModelTypeVLM:
		panic("not support yet")
	default:
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model type not support"})
	}
}

func responsesToolsString(params ResponseNewParams) (string, error) {
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

func profile2ResponseUsage(p nexa_sdk.ProfileData) responses.ResponseUsage {
	return responses.ResponseUsage{
		InputTokens:         p.PromptTokens,
		InputTokensDetails:  responses.ResponseUsageInputTokensDetails{CachedTokens: 0},
		OutputTokens:        p.GeneratedTokens,
		OutputTokensDetails: responses.ResponseUsageOutputTokensDetails{ReasoningTokens: 0},
		TotalTokens:         p.TotalTokens(),
	}
}

func parseResponseToolCalls(fullText string) (responses.ResponseOutputItemUnion, error) {
	match := responsesToolCallRegex.FindStringSubmatch(fullText)
	if len(match) <= 1 {
		return responses.ResponseOutputItemUnion{}, errors.New("tool call not match")
	}
	matched := match[1]
	if matched == "" && len(match) > 2 {
		matched = match[2]
	}
	toolCall := responses.ResponseOutputItemUnion{
		Type:   "function_call",
		ID:     "fc_" + fmt.Sprintf("%x", rand.Int63()),
		CallID: "call_" + fmt.Sprintf("%x", rand.Int63()),
	}
	name, err := sonic.GetFromString(matched, "name")
	if err != nil {
		return toolCall, err
	}
	toolCall.Name, err = name.String()
	if err != nil {
		return toolCall, err
	}
	arguments, err := sonic.GetFromString(matched, "arguments")
	if err != nil {
		return toolCall, err
	}
	switch arguments.TypeSafe() {
	case ast.V_OBJECT:
		toolCall.Arguments, _ = arguments.Raw()
	case ast.V_STRING:
		toolCall.Arguments, _ = arguments.String()
	default:
		return toolCall, errors.New("unknown arguments type")
	}
	toolCall.Status = "completed"
	return toolCall, nil
}

func responsesCreateLLM(c *gin.Context, req ResponseCreateRequest) {
	createdAt := float64(time.Now().Unix())
	p, err := service.KeepAliveGet[nexa_sdk.LLM](
		string(req.Model),
		types.ModelParam{NCtx: req.NCtx, NGpuLayers: req.Ngl},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]any{"error": "model not found"})
		return
	}
	slog.Debug("ParseNewParamsLLM called", "req", req.ResponseNewParams)
	input, err := parseNewParamsLLM(req.ResponseNewParams)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	slog.Debug("ApplyChatTemplate called", "input", input)
	promptText, err := p.ApplyChatTemplate(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	slog.Debug("ApplyChatTemplate output", "output", promptText.FormattedText)
	if req.Stream {
		// responsesCreateLLMStream(c)
	}
	responsesCreateLLMNonStream(c, req, p, promptText.FormattedText, createdAt)
}

func parseNewParamsLLM(params ResponseNewParams) (nexa_sdk.LlmApplyChatTemplateInput, error) {
	tmpl := nexa_sdk.LlmApplyChatTemplateInput{
		EnableThink:         true,
		AddGenerationPrompt: true,
	}
	if !param.IsOmitted(params.Instructions) {
		tmpl.Messages = []nexa_sdk.LlmChatMessage{
			{Role: nexa_sdk.LLMRoleSystem, Content: params.Instructions.Value},
		}
	}
	input := params.Input
	if param.IsOmitted(input.OfString) && len(input.OfInputItemList) == 0 {
		return tmpl, nil
	}
	if !param.IsOmitted(input.OfString) {
		userMsg := nexa_sdk.LlmChatMessage{
			Role:    nexa_sdk.LLMRole(string(openai.MessageRoleUser)),
			Content: input.OfString.Value,
		}
		if tmpl.Messages == nil {
			tmpl.Messages = []nexa_sdk.LlmChatMessage{userMsg}
		} else {
			tmpl.Messages = append(tmpl.Messages, userMsg)
		}
		toolsStr, err := responsesToolsString(params)
		if err != nil {
			slog.Error("Failed to convert tools to string", "error", err)
			return tmpl, fmt.Errorf("failed to convert tools to string: %w", err)
		}
		tmpl.Tools = toolsStr
		return tmpl, nil
	}
	tmpl.Messages = make([]nexa_sdk.LlmChatMessage, 0, len(input.OfInputItemList))
	for _, item := range input.OfInputItemList {
		if item.OfMessage != nil {
			role := string(item.OfMessage.Role)
			if role == "" {
				role = string(openai.MessageRoleUser)
			}
			content := ""
			if !param.IsOmitted(item.OfMessage.Content.OfString) {
				content = item.OfMessage.Content.OfString.Value
			}
			tmpl.Messages = append(tmpl.Messages, nexa_sdk.LlmChatMessage{
				Role:    nexa_sdk.LLMRole(role),
				Content: content,
			})
			continue
		}
		if item.OfInputMessage != nil {
			role := item.OfInputMessage.Role
			if role == "" {
				role = string(openai.MessageRoleUser)
			}
			var text string
			for _, c := range item.OfInputMessage.Content {
				if c.OfInputText != nil {
					text += c.OfInputText.Text
				}
			}
			tmpl.Messages = append(tmpl.Messages, nexa_sdk.LlmChatMessage{
				Role:    nexa_sdk.LLMRole(role),
				Content: text,
			})
			continue
		}
	}
	toolsStr, err := responsesToolsString(params)
	if err != nil {
		slog.Error("Failed to convert tools to string", "error", err)
		return tmpl, fmt.Errorf("failed to convert tools to string: %w", err)
	}
	tmpl.Tools = toolsStr
	return tmpl, nil
}

func responsesCreateLLMNonStream(c *gin.Context, req ResponseCreateRequest, p *nexa_sdk.LLM, prompt string, createdAt float64) {
	out, err := p.Generate(nexa_sdk.LlmGenerateInput{
		PromptUTF8: prompt,
		OnToken: func(token string) bool {
			return true
		},
		Config: &nexa_sdk.GenerationConfig{
			MaxTokens: int32(req.MaxOutputTokens.Value),
			SamplerConfig: &nexa_sdk.SamplerConfig{
				Temperature:       float32(req.Temperature.Value),
				TopP:              float32(req.TopP.Value),
				TopK:              req.TopK,
				MinP:              req.MinP,
				RepetitionPenalty: req.RepetitionPenalty,
			},
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	usage := profile2ResponseUsage(out.ProfileData)
	respID := fmt.Sprintf("resp_%x%x", time.Now().UnixNano(), rand.Int63())

	res := responses.Response{
		ID:          respID,
		Object:      (constant.Response)("response"),
		CreatedAt:   createdAt,
		Status:      responses.ResponseStatusCompleted,
		CompletedAt: float64(time.Now().Unix()),
		Usage:       usage,
		Output:      []responses.ResponseOutputItemUnion{},
	}
	if len(req.Tools) > 0 {
		toolCall, err := parseResponseToolCalls(out.FullText)
		if err != nil {
			slog.Error("Failed to parse tool call", "error", err, "fullText", out.FullText)
			c.JSON(http.StatusInternalServerError, map[string]any{"parse tool call error": err.Error()})
			return
		}
		res.Output = append(res.Output, toolCall)
		c.JSON(http.StatusOK, res)
		return
	}
	res.Output = append(res.Output, responses.ResponseOutputItemUnion{
		Type:    "message",
		Role:    constant.Assistant(openai.MessageRoleAssistant),
		Status:  "completed",
		Content: []responses.ResponseOutputMessageContentUnion{{Type: "output_text", Text: out.FullText, Annotations: nil}},
	})
	c.JSON(http.StatusOK, res)
}
