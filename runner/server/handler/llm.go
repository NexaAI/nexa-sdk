package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared/constant"

	"github.com/NexaAI/nexa-sdk/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server/service"
)

// @Router			/completions [post]
// @Summary		completion
// @Description	Legacy completion endpoint for text generation. It is recommended to use the Chat Completions endpoint for new applications.
// @Accept			json
// @Param			request	body	openai.CompletionNewParams	true	"Completion request"
// @Produce		json
// @Success		200	{object}	openai.Completion
func Completions(c *gin.Context) {
	param := openai.CompletionNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	p, err := service.KeepAliveGet[nexa_sdk.LLM](
		string(param.Model),
		types.ModelParam{Device: nil, CtxLen: 8192},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	data, err := p.Generate(param.Prompt.OfString.String())
	choice := openai.CompletionChoice{}
	choice.Text = data
	res := openai.Completion{
		Choices: []openai.CompletionChoice{choice},
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, res)
	}
}

type ChatCompletionNewParams openai.ChatCompletionNewParams

// ChatCompletionRequest defines the request body for the chat completions API.
// example: { "model": "nexaml/nexaml-models", "messages": [ { "role": "user", "content": "why is the sky blue?" } ] }
type ChatCompletionRequest struct {
	Stream bool `json:"stream" default:"false"`

	ChatCompletionNewParams
}

var toolCallRegex = regexp.MustCompile(`<tool_call>([\s\S]+)<\/tool_call>`)

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

	for _, msg := range param.Messages {
		if _, ok := msg.GetContent().AsAny().(*string); !ok {
			chatVLMCompletions(c, param)
			return
		}
	}
	chatLLMCompletions(c, param)
}

func chatLLMCompletions(c *gin.Context, param ChatCompletionRequest) {
	// get llm
	p, err := service.KeepAliveGet[nexa_sdk.LLM](
		string(param.Model),
		types.ModelParam{CtxLen: 8192},
	)
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, map[string]any{"error": "model not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// emtry request for warm up
	if len(param.Messages) == 0 {
		p.Reset()
		c.JSON(http.StatusOK, nil)
		return
	}

	messages := make([]nexa_sdk.ChatMessage, 0, len(param.Messages))
	for _, msg := range param.Messages {
		messages = append(messages, nexa_sdk.ChatMessage{
			Role:    nexa_sdk.LLMRole(*msg.GetRole()),
			Content: *msg.GetContent().AsAny().(*string),
		})
	}

	if len(param.Tools) > 0 {
		tools := make([]nexa_sdk.ChatTool, 0, len(param.Tools))
		for _, tool := range param.Tools {
			tools = append(tools, nexa_sdk.ChatTool{
				Type: string(tool.Type),
				Function: nexa_sdk.ChatToolFunction{
					Name: tool.Function.Name,
				},
			})
		}
		param := nexa_sdk.ChatTemplateParam{
			Messages: messages,
			Tools:    tools,
		}
		formatted, err := p.ApplyJinjaTemplate(param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		data, err := p.Generate(formatted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		// parse result
		match := toolCallRegex.FindStringSubmatch(data)
		if len(match) <= 1 {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "not match"})
			return
		}
		toolCall := openai.ChatCompletionMessageToolCall{Type: constant.Function("")}
		err = sonic.UnmarshalString(match[1], &toolCall.Function)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "data": match[1]})
			return
		}

		choice := openai.ChatCompletionChoice{}
		choice.Message.Role = constant.Assistant("")
		choice.Message.ToolCalls = []openai.ChatCompletionMessageToolCall{toolCall}
		res := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{choice},
		}
		c.JSON(http.StatusOK, res)
		return

	} else {

		formatted, err := p.ApplyChatTemplate(messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		if param.Stream {
			ctx, cancel := context.WithCancel(context.Background())
			dataCh, errCh := p.GenerateStream(ctx, formatted)

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
			cancel()

			e, ok := <-errCh
			if ok {
				fmt.Printf("GenerateStream Error: %s\n", e)
			}
		} else {
			data, err := p.Generate(formatted)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}

			choice := openai.ChatCompletionChoice{}
			choice.Message.Role = constant.Assistant(openai.MessageRoleAssistant)
			choice.Message.Content = data
			res := openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{choice},
			}
			c.JSON(http.StatusOK, res)
			return
		}
	}
}

func chatVLMCompletions(c *gin.Context, param ChatCompletionRequest) {
	// get vlm

	p, err := service.KeepAliveGet[nexa_sdk.VLM](
		string(param.Model),
		types.ModelParam{CtxLen: 8192},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// emtry request for warm up
	if len(param.Messages) == 0 {
		p.Reset()
		c.JSON(http.StatusOK, nil)
		return
	}

	var images, audios []string

	messages := make([]nexa_sdk.ChatMessage, 0, len(param.Messages))
	for _, msg := range param.Messages {
		content := msg.GetContent().AsAny().(*[]openai.ChatCompletionContentPartUnionParam)
		for _, ct := range *content {
			switch *ct.GetType() {
			case "text":
				messages = append(messages, nexa_sdk.ChatMessage{
					Role:    nexa_sdk.LLMRole(*msg.GetRole()),
					Content: *ct.GetText(),
				})
			case "input_audio":
				audios = append(audios, ct.GetInputAudio().Data)
			case "image_url":
				images = append(images, ct.GetImageURL().URL)
			}
		}
	}

	if len(param.Tools) > 0 {
		tools := make([]nexa_sdk.ChatTool, 0, len(param.Tools))
		for _, tool := range param.Tools {
			tools = append(tools, nexa_sdk.ChatTool{
				Type: string(tool.Type),
				Function: nexa_sdk.ChatToolFunction{
					Name: tool.Function.Name,
				},
			})
		}
		param := nexa_sdk.ChatTemplateParam{
			Messages: messages,
			Tools:    tools,
		}
		formatted, err := p.ApplyJinjaTemplate(param)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		data, err := p.Generate(formatted, images, audios)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		// parse result
		match := toolCallRegex.FindStringSubmatch(data)
		if len(match) <= 1 {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": "not match"})
			return
		}
		toolCall := openai.ChatCompletionMessageToolCall{Type: constant.Function("")}
		err = sonic.UnmarshalString(match[1], &toolCall.Function)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err, "data": match[1]})
			return
		}

		choice := openai.ChatCompletionChoice{}
		choice.Message.Role = constant.Assistant("")
		choice.Message.ToolCalls = []openai.ChatCompletionMessageToolCall{toolCall}
		res := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{choice},
		}
		c.JSON(http.StatusOK, res)
		return

	} else {

		formatted, err := p.ApplyChatTemplate(messages)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}

		if param.Stream {
			ctx, cancel := context.WithCancel(context.Background())
			dataCh, errCh := p.GenerateStream(ctx, formatted, images, audios)

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
			cancel()

			e, ok := <-errCh
			if ok {
				fmt.Printf("GenerateStream Error: %s\n", e)
			}
		} else {
			data, err := p.Generate(formatted, images, audios)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}

			choice := openai.ChatCompletionChoice{}
			choice.Message.Role = constant.Assistant(openai.MessageRoleAssistant)
			choice.Message.Content = data
			res := openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{choice},
			}
			c.JSON(http.StatusOK, res)
			return
		}
	}
}

//type KVCacheRequest struct {
//	Model string
//	Name  string
//}
//
//func SaveKVCache(c *gin.Context) {
//	param := KVCacheRequest{}
//	if err := c.ShouldBindJSON(&param); err != nil {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
//	}
//
//	dir, _ := path.Split(param.Name)
//	if dir != "" {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": "name invalid"})
//	}
//
//	p := service.KeepAliveGet(param.Model, createLLM(param.Model))
//	s := store.Get()
//	if err := p.SaveKVCache(s.CachefilePath(param.Name)); err != nil {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": err})
//	} else {
//		c.JSON(http.StatusOK, nil)
//	}
//
//	c.JSON(http.StatusOK, nil)
//
//}
//
//func LoadKVCache(c *gin.Context) {
//	param := KVCacheRequest{}
//	if err := c.ShouldBindJSON(&param); err != nil {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
//	}
//
//	dir, _ := path.Split(param.Name)
//	if dir != "" {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": "name invalid"})
//	}
//
//	p := service.KeepAliveGet(param.Model, createLLM(param.Model))
//	s := store.Get()
//	p.Reset()
//	if err := p.LoadKVCache(s.CachefilePath(param.Name)); err != nil {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": err})
//	} else {
//		c.JSON(http.StatusOK, nil)
//	}
//}
