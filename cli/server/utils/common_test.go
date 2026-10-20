// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package utils

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSaveURIToTempFile_WebP tests WebP to PNG conversion
func TestSaveURIToTempFile_WebP(t *testing.T) {
	// Create a simple test WebP image (1x1 red pixel)
	// Note: This is a minimal valid WebP file in base64
	webpData := "UklGRiQAAABXRUJQVlA4IBgAAAAwAQCdASoBAAEAAQAcJaQAA3AA/v3AgAA="

	webpBytes, err := base64.StdEncoding.DecodeString(webpData)
	if err != nil {
		t.Fatalf("Failed to decode test WebP data: %v", err)
	}

	// Create temp WebP file
	tmpWebP, err := os.CreateTemp("", "test-*.webp")
	if err != nil {
		t.Fatalf("Failed to create temp WebP file: %v", err)
	}
	defer os.Remove(tmpWebP.Name())

	if _, err := tmpWebP.Write(webpBytes); err != nil {
		t.Fatalf("Failed to write WebP data: %v", err)
	}
	tmpWebP.Close()

	// Test conversion via file path
	resultPath, err := SaveURIToTempFile(tmpWebP.Name())
	if err != nil {
		t.Fatalf("SaveURIToTempFile failed: %v", err)
	}
	defer os.Remove(resultPath)

	// Verify output is PNG
	if !strings.HasSuffix(resultPath, ".png") {
		t.Errorf("Expected PNG extension, got: %s", resultPath)
	}

	// Verify the file can be decoded as PNG
	resultFile, err := os.Open(resultPath)
	if err != nil {
		t.Fatalf("Failed to open result file: %v", err)
	}
	defer resultFile.Close()

	_, err = png.Decode(resultFile)
	if err != nil {
		t.Errorf("Failed to decode result as PNG: %v", err)
	}
}

// TestSaveURIToTempFile_WebP_DataURI tests WebP conversion from data URI
func TestSaveURIToTempFile_WebP_DataURI(t *testing.T) {
	// Minimal valid WebP as data URI
	webpData := "UklGRiQAAABXRUJQVlA4IBgAAAAwAQCdASoBAAEAAQAcJaQAA3AA/v3AgAA="
	dataURI := "data:image/webp;base64," + webpData

	resultPath, err := SaveURIToTempFile(dataURI)
	if err != nil {
		t.Fatalf("SaveURIToTempFile failed for data URI: %v", err)
	}
	defer os.Remove(resultPath)

	if !strings.HasSuffix(resultPath, ".png") {
		t.Errorf("Expected PNG extension for converted WebP, got: %s", resultPath)
	}
	resultFile, err := os.Open(resultPath)
	if err != nil {
		t.Fatal(err)
	}
	defer resultFile.Close()
	if _, err := png.Decode(resultFile); err != nil {
		t.Errorf("data URI output is not PNG: %v", err)
	}
}

func TestSaveURIToTempFile_PNG(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "source.png")
	if err := os.WriteFile(source, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	resultPath, err := SaveURIToTempFile(source)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(resultPath)
	got, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, buf.Bytes()) {
		t.Error("PNG input was modified")
	}
}
