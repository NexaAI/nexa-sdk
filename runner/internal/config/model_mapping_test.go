// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package config

import (
	"testing"
)

func TestGetModelMapping(t *testing.T) {
	tests := []struct {
		shortcut    string
		expected    string
		shouldExist bool
	}{
		{"qwen3", "Qwen/Qwen3-4B-GGUF", true},
		{"qwen2vl", "ggml-org/Qwen2-VL-2B-Instruct-GGUF", true},
		{"qwen2.5vl", "Qwen/Qwen2.5-VL-3B-Instruct", true},
		{"gemma3", "ggml-org/gemma-3-4b-it-GGUF", true},
		{"smolvlm", "ggml-org/SmolVLM-500M-Instruct-GGUF", true},
		{"unknown", "", false},
		{"", "", false},
	}

	for _, test := range tests {
		actual, exists := GetModelMapping(test.shortcut)
		if exists != test.shouldExist {
			t.Errorf("GetModelMapping(%q) exists = %v, want %v", test.shortcut, exists, test.shouldExist)
		}
		if exists && actual != test.expected {
			t.Errorf("GetModelMapping(%q) = %q, want %q", test.shortcut, actual, test.expected)
		}
	}
}
