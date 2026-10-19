// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestStatusNoModelLoaded(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/status", Status)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/status", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got StatusResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !got.Alive {
		t.Error("alive = false, want true")
	}
	if got.Models == nil || len(got.Models) != 0 {
		t.Errorf("models = %v, want empty non-nil slice", got.Models)
	}
}
