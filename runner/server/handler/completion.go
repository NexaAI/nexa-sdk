package handler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared/constant"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
)

func Completions(c *gin.Context) {
	c.JSON(http.StatusGone, map[string]any{"error": "this endpoint is deprecated, please use /chat/completions instead"})
}

type ChatCompletionNewParams openai.ChatCompletionNewParams

// ChatCompletionRequest defines the request body for the chat completions API.
// example: { "model": "nexaml/nexaml-models", "messages": [ { "role": "user", "content": "why is the sky blue?" } ] }
type ChatCompletionRequest struct {
	Stream bool `json:"stream" default:"false"`

	EnableThink bool `json:"enable_think" default:"true"`

	TopK              int32   `json:"top_k" default:"0"`
	MinP              float32 `json:"min_p" default:"0.0"`
	ReqetitionPenalty float32 `json:"repetition_penalty" default:"1.0"`
	GrammarPath       string  `json:"grammar_path" default:""`
	GrammarString     string  `json:"grammar_string" default:""`
	EnableJson        bool    `json:"enable_json" default:"false"`

	ChatCompletionNewParams
}

var toolCallRegex = regexp.MustCompile(`<tool_call>([\s\S]+)<\/tool_call>` + "|" + "```json([\\s\\S]+)```")

// @Router			/chat/completions [post]
// @Summary		Creates a model response for the given chat conversation.
// @Description	This endpoint generates a model response for a given conversation, which can include text and images. It supports both single-turn and multi-turn conversations and can be used for various tasks like question answering, code generation, and function calling.
// @Accept			json
// @Param			request	body	ChatCompletionRequest	true	"Chat completion request"
// @Produce		json
// @Success		200	{object}	openai.ChatCompletion	"Successful response for non-streaming requests."
func ChatCompletions(c *gin.Context) {
	param := ChatCompletionRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	s := store.Get()
	manifest, err := s.GetManifest(param.Model)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	switch manifest.ModelType {
	case types.ModelTypeLLM:
		chatCompletionsLLM(c, param)
	case types.ModelTypeVLM:
		chatCompletionsVLM(c, param)
	default:
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model type not support"})
		return
	}
}

func chatCompletionsLLM(c *gin.Context, param ChatCompletionRequest) {
	// Build message list for LLM template
	var systemPrompt string
	messages := make([]nexa_sdk.LlmChatMessage, 0, len(param.Messages))
	for _, msg := range param.Messages {
		if msg.GetRole() == nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "role is nil"})
			return
		}
		switch msg.GetContent().AsAny().(type) {
		case *string: // ok
		default:
			c.JSON(http.StatusBadRequest, map[string]any{"error": "content type not support"})
			return
		}
		// patch for npu
		if *msg.GetRole() == "system" {
			systemPrompt += *msg.GetContent().AsAny().(*string)
		}
		messages = append(messages, nexa_sdk.LlmChatMessage{
			Role:    nexa_sdk.LLMRole(*msg.GetRole()),
			Content: *msg.GetContent().AsAny().(*string),
		})
	}

	// Prepare tools if provided
	parseTool, tools, err := parseTools(param)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	samplerConfig := parseSamplerConfig(param)

	// Get LLM instance
	p, err := service.KeepAliveGet[nexa_sdk.LLM](
		string(param.Model),
		types.ModelParam{NCtx: 4096, NGpuLayers: 999, SystemPrompt: systemPrompt},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]any{"error": "model not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	// Empty request for warm up
	if len(param.Messages) == 0 {
		c.JSON(http.StatusOK, nil)
		return
	}

	// Format prompt using chat template
	formatted, err := p.ApplyChatTemplate(nexa_sdk.LlmApplyChatTemplateInput{
		Messages:    messages,
		Tools:       tools,
		EnableThink: param.EnableThink,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	if param.Stream {
		// Streaming response mode
		stopGen := false
		dataCh := make(chan string)
		resCh := make(chan nexa_sdk.LlmGenerateOutput)

		go func() {
			res, err := p.Generate(nexa_sdk.LlmGenerateInput{
				PromptUTF8: formatted.FormattedText,
				OnToken: func(token string) bool {
					if stopGen {
						return false
					}
					dataCh <- token
					return true
				},
				Config: &nexa_sdk.GenerationConfig{
					MaxTokens:     2048,
					SamplerConfig: samplerConfig,
				},
			},
			)
			if err != nil {
				slog.Warn("Generate Error", "error", err)
			}

			close(dataCh)
			resCh <- res
			close(resCh)
		}()

		c.Stream(func(w io.Writer) bool {
			r, ok := <-dataCh
			if ok {
				chunk := openai.ChatCompletionChunk{}
				chunk.Choices = append(chunk.Choices, openai.ChatCompletionChunkChoice{
					Delta: openai.ChatCompletionChunkChoiceDelta{
						Content: r,
					},
				})

				c.SSEvent("", chunk)
				return true
			}

			if param.StreamOptions.IncludeUsage.Value {
				res := <-resCh
				c.SSEvent("", openai.ChatCompletionChunk{
					Usage: profile2Usage(res.ProfileData),
				})
			}

			c.SSEvent("", "[DONE]")

			return false
		})

		stopGen = true
		for range dataCh {
		}
		for range resCh {
		}

	} else {
		// Blocking response mode
		genOut, err := p.Generate(nexa_sdk.LlmGenerateInput{
			PromptUTF8: formatted.FormattedText,
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens:     2048,
				SamplerConfig: samplerConfig,
			},
		},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		// Standard output (no tool parse)
		if !parseTool {
			choice := openai.ChatCompletionChoice{}
			choice.Message.Role = constant.Assistant(openai.MessageRoleAssistant)
			choice.Message.Content = genOut.FullText
			res := openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{choice},
				Usage:   profile2Usage(genOut.ProfileData),
			}
			c.JSON(http.StatusOK, res)
			return
		} else {
			// Tool call output parsing
			match := toolCallRegex.FindStringSubmatch(genOut.FullText)
			if len(match) <= 1 {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": "not match", "data": genOut.FullText})
				return
			}
			toolCall := openai.ChatCompletionMessageToolCall{Type: constant.Function("")}
			err = sonic.UnmarshalString("{"+match[1]+"}", &toolCall.Function)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "data": match[1]})
				return
			}

			choice := openai.ChatCompletionChoice{}
			choice.Message.Role = constant.Assistant("")
			choice.Message.ToolCalls = []openai.ChatCompletionMessageToolCall{toolCall}
			res := openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{choice},
				Usage:   profile2Usage(genOut.ProfileData),
			}
			c.JSON(http.StatusOK, res)
			return
		}
	}
}

func chatCompletionsVLM(c *gin.Context, param ChatCompletionRequest) {
	// Build message list for VLM template
	var systemPrompt string
	messages := make([]nexa_sdk.VlmChatMessage, 0, len(param.Messages))
	images := make([]string, 0)
	audios := make([]string, 0)
	for _, msg := range param.Messages {
		if msg.GetRole() == nil {
			c.JSON(http.StatusBadRequest, map[string]any{"error": "role is nil"})
			return
		}

		switch content := msg.GetContent().AsAny().(type) {
		case *string:
			if *msg.GetRole() == "system" {
				systemPrompt += *content
			}
			messages = append(messages, nexa_sdk.VlmChatMessage{
				Role: nexa_sdk.VlmRole(*msg.GetRole()),
				Contents: []nexa_sdk.VlmContent{
					{Type: nexa_sdk.VlmContentTypeText, Text: *msg.GetContent().AsAny().(*string)},
				},
			})

		case *[]openai.ChatCompletionContentPartUnionParam:
			contents := make([]nexa_sdk.VlmContent, 0, len(*content))

			for _, ct := range *content {
				switch *ct.GetType() {
				case "text":
					contents = append(contents, nexa_sdk.VlmContent{
						Type: nexa_sdk.VlmContentTypeText,
						Text: *ct.GetText(),
					})
				case "image_url":
					file, err := utils.SaveURIToTempFile(ct.GetImageURL().URL)
					slog.Debug("Saved image file", "file", file)
					if err != nil {
						c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
						return
					}
					defer os.Remove(file)
					contents = append(contents, nexa_sdk.VlmContent{
						Type: nexa_sdk.VlmContentTypeImage,
						Text: file,
					})
					images = append(images, file)
				case "input_audio":
					file, err := utils.SaveURIToTempFile(ct.GetInputAudio().Data)
					slog.Debug("Saved audio file", "file", file)
					if err != nil {
						c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
						return
					}
					defer os.Remove(file)
					contents = append(contents, nexa_sdk.VlmContent{
						Type: nexa_sdk.VlmContentTypeAudio,
						Text: file,
					})
					audios = append(audios, file)
				}
			}

			messages = append(messages, nexa_sdk.VlmChatMessage{
				Role:     nexa_sdk.VlmRole(*msg.GetRole()),
				Contents: contents,
			})

		default:
			c.JSON(http.StatusBadRequest, map[string]any{"error": "unknown content type"})
			return
		}
	}

	// Prepare tools if provided
	parseTool, tools, err := parseTools(param)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	samplerConfig := parseSamplerConfig(param)

	// Get VLM instance
	p, err := service.KeepAliveGet[nexa_sdk.VLM](
		string(param.Model),
		types.ModelParam{NCtx: 4096, NGpuLayers: 999, SystemPrompt: systemPrompt},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]any{"error": "model not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Empty request for warm up, just reset model state
	if len(param.Messages) == 0 {
		c.JSON(http.StatusOK, nil)
		return
	}

	// Format prompt using VLM chat template
	formatted, err := p.ApplyChatTemplate(nexa_sdk.VlmApplyChatTemplateInput{
		Messages:    messages,
		Tools:       tools,
		EnableThink: param.EnableThink,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	if param.Stream {
		// Streaming response mode
		stopGen := false
		dataCh := make(chan string)

		go func() {
			_, err := p.Generate(nexa_sdk.VlmGenerateInput{
				PromptUTF8: formatted.FormattedText,
				OnToken: func(token string) bool {
					if stopGen {
						return false
					}
					dataCh <- token
					return true
				},
				Config: &nexa_sdk.GenerationConfig{
					MaxTokens:     2048,
					SamplerConfig: samplerConfig,
					ImagePaths:    images,
					AudioPaths:    audios,
				},
			},
			)

			close(dataCh)

			if err != nil {
				slog.Warn("Generate Error", "error", err)
			}
		}()

		c.Stream(func(w io.Writer) bool {
			r, ok := <-dataCh
			if ok {
				chunk := openai.ChatCompletionChunk{}
				chunk.Choices = append(chunk.Choices, openai.ChatCompletionChunkChoice{
					Delta: openai.ChatCompletionChunkChoiceDelta{
						Content: r,
					},
				})

				c.SSEvent("", chunk)
				return true
			}
			c.SSEvent("", "[DONE]")

			return false
		})

		stopGen = true
		for range dataCh {
		}
	} else {
		// Blocking response mode
		genOut, err := p.Generate(nexa_sdk.VlmGenerateInput{
			PromptUTF8: formatted.FormattedText,
			Config: &nexa_sdk.GenerationConfig{
				MaxTokens:     2048,
				SamplerConfig: samplerConfig,
				ImagePaths:    images,
				AudioPaths:    audios,
			},
		},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		// Standard output (no tool parse)
		if !parseTool {
			choice := openai.ChatCompletionChoice{}
			choice.Message.Role = constant.Assistant(openai.MessageRoleAssistant)
			choice.Message.Content = genOut.FullText
			res := openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{choice},
				Usage:   profile2Usage(genOut.ProfileData),
			}
			c.JSON(http.StatusOK, res)
			return
		} else {
			// Tool call output parsing
			match := toolCallRegex.FindStringSubmatch(genOut.FullText)
			if len(match) <= 1 {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": "not match", "data": genOut.FullText})
				return
			}
			toolCall := openai.ChatCompletionMessageToolCall{Type: constant.Function("")}
			err = sonic.UnmarshalString("{"+match[1]+"}", &toolCall.Function)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "data": match[1]})
				return
			}

			choice := openai.ChatCompletionChoice{}
			choice.Message.Role = constant.Assistant("")
			choice.Message.ToolCalls = []openai.ChatCompletionMessageToolCall{toolCall}
			res := openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{choice},
				Usage:   profile2Usage(genOut.ProfileData),
			}
			c.JSON(http.StatusOK, res)
			return
		}
	}
}

func profile2Usage(p nexa_sdk.ProfileData) openai.CompletionUsage {
	return openai.CompletionUsage{
		CompletionTokens: p.GeneratedTokens,
		PromptTokens:     p.PromptTokens,
		TotalTokens:      p.TotalTokens(),
	}
}

func parseSamplerConfig(param ChatCompletionRequest) *nexa_sdk.SamplerConfig {
	// parse sampling parameters
	var samplerConfig *nexa_sdk.SamplerConfig
	samplerConfig = &nexa_sdk.SamplerConfig{
		Temperature:       float32(param.Temperature.Value),
		TopP:              float32(param.TopP.Value),
		TopK:              param.TopK,
		MinP:              param.MinP,
		RepetitionPenalty: param.ReqetitionPenalty,
		PresencePenalty:   float32(param.PresencePenalty.Value),
		FrequencyPenalty:  float32(param.FrequencyPenalty.Value),
		Seed:              int32(param.Seed.Value),
		EnableJson:        param.EnableJson,
	}
	return samplerConfig
}

func parseTools(param ChatCompletionRequest) (bool, string, error) {
	if len(param.Tools) == 0 {
		return false, "", nil
	}

	tools, err := sonic.MarshalString(param.Tools)
	return true, tools, err
}
