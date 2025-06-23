package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server/service"
)

func createLLM(name string) func() nexa_sdk.LLM {
	return func() nexa_sdk.LLM {
		time.Sleep(2 * time.Second) // TODO: remove test code
		s := store.NewStore()
		return nexa_sdk.NewLLM(s.ModelfilePath(name), nil, 4096, nil)
	}
}

func Completions(c *gin.Context) {
	param := openai.CompletionNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
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
		Content string `json:"Content"`
	} `json:"messages"`
}

func ChatCompletions(c *gin.Context) {
	param := ChatCompletionRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

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
		content := msg.Content
		messages = append(messages, nexa_sdk.ChatMessage{
			Role:    nexa_sdk.LLMRole(msg.Role),
			Content: content,
		})
	}

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
		choice.Message.Content = data
		res := openai.ChatCompletion{
			Choices: []openai.ChatCompletionChoice{choice},
		}
		c.JSON(http.StatusOK, res)
		return
	}
}

type KVCacheRequest struct {
	Model string
	Name  string
}

func SaveKVCache(c *gin.Context) {
	param := KVCacheRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	dir, _ := path.Split(param.Name)
	if dir != "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "name invalid"})
	}

	p := service.KeepAliveGet(param.Model, createLLM(param.Model))
	s := store.NewStore()
	if err := p.SaveKVCache(s.CachefilePath(param.Name)); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err})
	} else {
		c.JSON(http.StatusOK, nil)
	}

	c.JSON(http.StatusOK, nil)

}

func LoadKVCache(c *gin.Context) {
	param := KVCacheRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	dir, _ := path.Split(param.Name)
	if dir != "" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "name invalid"})
	}

	p := service.KeepAliveGet(param.Model, createLLM(param.Model))
	s := store.NewStore()
	p.Reset()
	if err := p.LoadKVCache(s.CachefilePath(param.Name)); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err})
	} else {
		c.JSON(http.StatusOK, nil)
	}
}
