// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/qualcomm/GenieX/cli/server/service"
)

var modelStatus = service.Status

func Health(c *gin.Context) {
	for _, m := range modelStatus() {
		if m.State == service.StateLoading {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": gin.H{
				"code":    http.StatusServiceUnavailable,
				"message": "Loading model",
				"type":    "unavailable_error",
			}})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
