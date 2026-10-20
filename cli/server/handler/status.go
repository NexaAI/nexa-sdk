// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/qualcomm/GenieX/cli/internal/config"
	"github.com/qualcomm/GenieX/cli/server/service"
)

var (
	Version   string
	startTime = time.Now()
)

type StatusResponse struct {
	Alive            bool                  `json:"alive"`
	Version          string                `json:"version,omitempty"`
	UptimeSeconds    int64                 `json:"uptime_seconds"`
	KeepAliveSeconds int64                 `json:"keepalive_seconds"`
	Models           []service.ModelStatus `json:"models"`
}

func Status(c *gin.Context) {
	c.JSON(http.StatusOK, StatusResponse{
		Alive:            true,
		Version:          Version,
		UptimeSeconds:    int64(time.Since(startTime).Seconds()),
		KeepAliveSeconds: config.Get().KeepAlive,
		Models:           service.Status(),
	})
}
