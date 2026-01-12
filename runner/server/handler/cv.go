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
	"log/slog"
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

	slog.Info("CV request received",
		"model", param.Model,
		"image", param.Image,
	)

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
