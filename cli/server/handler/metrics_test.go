// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
)

func getMetrics(t *testing.T) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/metrics", Metrics)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	return w
}

func TestMetricsDisabled(t *testing.T) {
	viper.Set("metrics", false)
	w := getMetrics(t)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want 501", w.Code)
	}
	if !strings.Contains(w.Body.String(), "not_supported_error") {
		t.Errorf("body = %s", w.Body.String())
	}
}

func TestMetricsCountsProfiles(t *testing.T) {
	viper.Set("metrics", true)
	t.Cleanup(func() { viper.Set("metrics", false) })
	promptTokens.Store(0)
	promptMicros.Store(0)
	predictedTokens.Store(0)
	predictedMicros.Store(0)

	recordProfile(geniex_sdk.ProfileData{PromptTokens: 10, PromptTime: 500000, GeneratedTokens: 20, DecodeTime: 2000000})
	recordProfile(geniex_sdk.ProfileData{PromptTokens: 5, PromptTime: 500000, GeneratedTokens: 7, DecodeTime: 1000000})

	w := getMetrics(t)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/plain") {
		t.Errorf("content-type = %s", ct)
	}
	body := w.Body.String()
	for _, want := range []string{
		"# TYPE geniex:prompt_tokens_total counter\ngeniex:prompt_tokens_total 15\n",
		"geniex:prompt_seconds_total 1\n",
		"geniex:tokens_predicted_total 27\n",
		"geniex:tokens_predicted_seconds_total 3\n",
		"# TYPE geniex:requests_processing gauge\ngeniex:requests_processing 0\n",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestTrackRequestGauge(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var during int64
	r := gin.New()
	r.Use(TrackRequest)
	r.GET("/x", func(c *gin.Context) { during = requestsProcessing.Load() })

	before := requestsProcessing.Load()
	r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))

	if during != before+1 {
		t.Errorf("during = %d, want %d", during, before+1)
	}
	if got := requestsProcessing.Load(); got != before {
		t.Errorf("after = %d, want %d", got, before)
	}
}
