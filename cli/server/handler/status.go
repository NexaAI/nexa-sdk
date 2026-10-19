// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qualcomm/GenieX/cli/server/service"
)

type StatusResponse struct {
	Alive  bool                  `json:"alive"`
	Models []service.ModelStatus `json:"models"`
}

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, StatusResponse{Alive: true, Models: service.Status()})
}
