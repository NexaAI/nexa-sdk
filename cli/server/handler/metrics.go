// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/gin-gonic/gin"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
	"github.com/qualcomm/GenieX/cli/internal/config"
)

var (
	promptTokens       atomic.Int64
	promptMicros       atomic.Int64
	predictedTokens    atomic.Int64
	predictedMicros    atomic.Int64
	requestsProcessing atomic.Int64
)

func recordProfile(p geniex_sdk.ProfileData) {
	promptTokens.Add(p.PromptTokens)
	promptMicros.Add(p.PromptTime)
	predictedTokens.Add(p.GeneratedTokens)
	predictedMicros.Add(p.DecodeTime)
}

func TrackRequest(c *gin.Context) {
	requestsProcessing.Add(1)
	defer requestsProcessing.Add(-1)
	c.Next()
}

func Metrics(c *gin.Context) {
	if !config.Get().Metrics {
		c.JSON(http.StatusNotImplemented, gin.H{"error": gin.H{
			"code":    http.StatusNotImplemented,
			"message": "This server does not support metrics endpoint. Start it with `--metrics`",
			"type":    "not_supported_error",
		}})
		return
	}
	var b strings.Builder
	metric := func(name, kind, help string, value any) {
		fmt.Fprintf(&b, "# HELP geniex:%s %s\n# TYPE geniex:%s %s\ngeniex:%s %v\n", name, help, name, kind, name, value)
	}
	metric("prompt_tokens_total", "counter", "Number of prompt tokens processed.", promptTokens.Load())
	metric("prompt_seconds_total", "counter", "Prompt process time in seconds.", float64(promptMicros.Load())/1e6)
	metric("tokens_predicted_total", "counter", "Number of generation tokens processed.", predictedTokens.Load())
	metric("tokens_predicted_seconds_total", "counter", "Predict process time in seconds.", float64(predictedMicros.Load())/1e6)
	metric("requests_processing", "gauge", "Number of requests processing.", requestsProcessing.Load())
	c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(b.String()))
}
