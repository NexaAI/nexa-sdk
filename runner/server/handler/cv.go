package handler

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
)

type CVRequest struct {
	Model string `json:"model" binding:"required"`
	Image string `json:"image"`
}

type CVResponse struct {
	Results []nexa_sdk.CVResult `json:"results"`
}

func CV(c *gin.Context) {
	param := CVRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	p, err := service.KeepAliveGet[nexa_sdk.CV](
		string(param.Model),
		types.ModelParam{},
		false,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	// warm up
	if param.Image == "" {
		c.JSON(http.StatusOK, nil)
		return
	}

	file, err := utils.SaveURIToTempFile(param.Image)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "failed to save image: " + err.Error()})
		return
	}
	defer os.Remove(file)
	res, err := p.Infer(nexa_sdk.CVInferInput{
		InputImagePath: file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
	} else {
		c.JSON(http.StatusOK, CVResponse{Results: res.Results})
	}
}
