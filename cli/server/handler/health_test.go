// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/qualcomm/GenieX/cli/server/service"
)

func TestHealthReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", Health)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if got, want := w.Body.String(), `{"status":"ok"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestHealthLoading(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orig := modelStatus
	modelStatus = func() []service.ModelStatus {
		return []service.ModelStatus{{ID: "m", State: service.StateLoading}}
	}
	t.Cleanup(func() { modelStatus = orig })
	r := gin.New()
	r.GET("/health", Health)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", w.Code)
	}
	want := `{"error":{"code":503,"message":"Loading model","type":"unavailable_error"}}`
	if got := w.Body.String(); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestHealthActiveIsReady(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orig := modelStatus
	modelStatus = func() []service.ModelStatus {
		return []service.ModelStatus{{ID: "m", State: service.StateActive}}
	}
	t.Cleanup(func() { modelStatus = orig })
	r := gin.New()
	r.GET("/health", Health)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
}
