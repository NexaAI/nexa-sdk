package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"
)

func Completions(c *gin.Context) {
	param := openai.CompletionNewParams{}
	c.ShouldBindJSON(param)
	c.JSON(http.StatusOK, nil)

}

func ChatCompletions(c *gin.Context) {
	c.JSON(http.StatusOK, nil)
}
