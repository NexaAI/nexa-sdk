package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/shared/constant"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server/service"
)

func createLLM(name string) func() nexa_sdk.LLM {
	return func() nexa_sdk.LLM {
		time.Sleep(2 * time.Second) // TODO: remove test code
		s := store.NewStore()
		file, err := s.ModelfilePath(name)
		if err != nil {
			panic(err) // TODO: fix signature
		}
		return nexa_sdk.NewLLM(file, nil, 4096, nil)
	}
}

// @Router /completions [post]
// @Summary completion
// @Description legancy completion
// @Accept    json
// @Param     model    body        ChatCompletionRequest   true   "example"
// @Produce   json
// @Success   200      {string}    Helloworld
func Completions(c *gin.Context) {
	param := openai.CompletionNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	p := service.KeepAliveGet(string(param.Model), createLLM(string(param.Model)))

	data, err := p.Generate(param.Prompt.OfString.String())
	choice := openai.CompletionChoice{}
	choice.Text = data
	res := openai.Completion{
		Choices: []openai.CompletionChoice{choice},
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
	} else {
		c.JSON(http.StatusOK, res)
	}
}

type ChatCompletionRequest struct {
	Stream   bool   `json:"stream"`
	Model    string `json:"model"`
	Messages []struct {
		Role    string `json:"role"`
		Content any    `json:"content"`
	} `json:"messages"`
	Tools []openai.ChatCompletionToolParam `json:"tools"`
}

var toolCallRegex = regexp.MustCompile("<tool_call>([\\s\\S]+)<\\/tool_call>")

// Function Call
// curl -v http://localhost:18181/v1/chat/completions -d '{ "model": "Qwen/Qwen2.5-1.5B-Instruct-GGUF", "messages": [ { "role": "user", "content": "What is the weather like in Boston today?" } ], "tools": [ { "type": "function", "function": { "name": "get_current_weather", "description": "Get the current weather in a given location", "parameters": { "type": "object", "properties": { "location": { "type": "string", "description": "The city and state, e.g. San Francisco, CA" }, "unit": { "type": "string", "enum": ["celsius", "fahrenheit"] } }, "required": ["location"] } } } ] }'
//
// VLM
// curl -v http://localhost:18181/v1/chat/completions -d '{ "model": "mradermacher/VLM-R1-Qwen2.5VL-3B-OVD-0321-i1-GGUF", "messages": [ { "role": "user", "content": [ { "type": "text", "text": "what is main color of the picture" }, { "type": "image_url", "image_url": "/home/remilia/Pictures/ScreenShot/20200201_182517.png" } ] } ] }'
func ChatCompletions(c *gin.Context) {
	param := ChatCompletionRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	for _, msg := range param.Messages {
		if _, ok := msg.Content.(string); !ok {
			chatVLMCompletions(c, param)
			return
		}
	}
	chatLLMCompletions(c, param)
}

func chatLLMCompletions(c *gin.Context, param ChatCompletionRequest) {
	// get llm
	p := service.KeepAliveGet(string(param.Model), createLLM(string(param.Model)))

	// emtry request for warm up
	if len(param.Messages) == 0 {
		p.Reset()
		c.JSON(http.StatusOK, nil)
		return
	}

	messages := make([]nexa_sdk.ChatMessage, 0, len(param.Messages))
	for _, msg := range param.Messages {
		content := msg.Content.(string)
		messages = append(messages, nexa_sdk.ChatMessage{
			Role:    nexa_sdk.LLMRole(msg.Role),
			Content: content,
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
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
			return
		}

		data, err := p.Generate(formatted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
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
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
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
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
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
	// get llm

	s := store.NewStore()
	file, err := s.ModelfilePath(param.Model)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
		return
	}
	p := nexa_sdk.NewVLM(file, nil, 4096, nil)
	defer p.Destroy()

	// emtry request for warm up
	if len(param.Messages) == 0 {
		p.Reset()
		c.JSON(http.StatusOK, nil)
		return
	}

	var imageUrl string

	messages := make([]nexa_sdk.ChatMessage, 0, len(param.Messages))
	for _, msg := range param.Messages {
		for _, ct := range msg.Content.([]any) {
			ct := ct.(map[string]any)
			if ct["type"] == "text" {
				messages = append(messages, nexa_sdk.ChatMessage{
					Role:    nexa_sdk.LLMRole(msg.Role),
					Content: ct["text"].(string),
				})
			} else {
				imageUrl = ct["image_url"].(string)
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
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
			return
		}

		data, err := p.Generate(formatted, &imageUrl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
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
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
			return
		}

		if param.Stream {
			ctx, cancel := context.WithCancel(context.Background())
			dataCh, errCh := p.GenerateStream(ctx, formatted, &imageUrl)

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
			data, err := p.Generate(formatted, &imageUrl)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": err})
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
//	s := store.NewStore()
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
//	s := store.NewStore()
//	p.Reset()
//	if err := p.LoadKVCache(s.CachefilePath(param.Name)); err != nil {
//		c.JSON(http.StatusBadRequest, map[string]any{"error": err})
//	} else {
//		c.JSON(http.StatusOK, nil)
//	}
//}
