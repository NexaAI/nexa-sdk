package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/internal/store"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/server/service"
)

func Completions(c *gin.Context) {
	param := openai.CompletionNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
	}

	s := store.NewStore()
	p := service.KeepAliveGet(string(param.Model), func() nexa_sdk.LLM {
		return nexa_sdk.NewLLM(s.ModelfilePath(string(param.Model)), nil, 4096, nil)
	})

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

func ChatCompletions(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
