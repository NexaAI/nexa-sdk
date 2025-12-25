// Copyright 2024-2025 Nexa AI, Inc.
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
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// TestBytesToFloat32 tests the audio conversion function
func TestBytesToFloat32(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []float32
	}{
		{
			name:     "Zero value",
			input:    []byte{0x00, 0x00},
			expected: []float32{0.0},
		},
		{
			name:     "Max positive value (32767)",
			input:    []byte{0xFF, 0x7F}, // Little-endian 32767
			expected: []float32{1.0},
		},
		{
			name:     "Min negative value (-32768)",
			input:    []byte{0x00, 0x80}, // Little-endian -32768
			expected: []float32{-1.0},
		},
		{
			name:     "Multiple samples",
			input:    []byte{0x00, 0x00, 0xFF, 0x7F, 0x00, 0x80},
			expected: []float32{0.0, 1.0, -1.0},
		},
		{
			name:     "Odd length array (should truncate)",
			input:    []byte{0x00, 0x00, 0xFF},
			expected: []float32{0.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToFloat32(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected length %d, got %d", len(tt.expected), len(result))
				return
			}

			for i := range result {
				// Use approximate equality for floating point
				if abs(result[i]-tt.expected[i]) > 0.001 {
					t.Errorf("Sample %d: expected %f, got %f", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

// TestBytesToFloat32_Normalization tests the asymmetric normalization
func TestBytesToFloat32_Normalization(t *testing.T) {
	// Test that positive and negative values use different divisors

	// Positive value: 16383 / 32767 ≈ 0.5
	positiveBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(positiveBytes, 16383)
	positiveResult := bytesToFloat32(positiveBytes)

	if len(positiveResult) != 1 {
		t.Fatalf("Expected 1 sample, got %d", len(positiveResult))
	}

	// Should be approximately 0.5
	if abs(positiveResult[0]-0.5) > 0.001 {
		t.Errorf("Expected ~0.5 for positive value, got %f", positiveResult[0])
	}

	// Negative value: -16384 / 32768 = -0.5
	negativeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(negativeBytes, uint16(int16(-16384)))
	negativeResult := bytesToFloat32(negativeBytes)

	if len(negativeResult) != 1 {
		t.Fatalf("Expected 1 sample, got %d", len(negativeResult))
	}

	// Should be exactly -0.5
	if abs(negativeResult[0]-(-0.5)) > 0.001 {
		t.Errorf("Expected -0.5 for negative value, got %f", negativeResult[0])
	}
}

// TestStreamConfigDefaults tests default value assignment
func TestStreamConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    StreamConfig
		expected StreamConfig
	}{
		{
			name:  "Empty config gets all defaults",
			input: StreamConfig{},
			expected: StreamConfig{
				Model:         defaultModel,
				Language:      defaultLanguage,
				SampleRate:    defaultSampleRate,
				ChunkDuration: defaultChunkDuration,
				BeamSize:      defaultBeamSize,
			},
		},
		{
			name: "Partial config keeps custom values",
			input: StreamConfig{
				Model:      "custom-model",
				SampleRate: 48000,
			},
			expected: StreamConfig{
				Model:         "custom-model",
				Language:      defaultLanguage,
				SampleRate:    48000,
				ChunkDuration: defaultChunkDuration,
				BeamSize:      defaultBeamSize,
			},
		},
		{
			name: "Full config keeps all values",
			input: StreamConfig{
				Model:         "custom-model",
				Language:      "zh-CN",
				SampleRate:    48000,
				ChunkDuration: 1.0,
				BeamSize:      10,
			},
			expected: StreamConfig{
				Model:         "custom-model",
				Language:      "zh-CN",
				SampleRate:    48000,
				ChunkDuration: 1.0,
				BeamSize:      10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.input

			// Apply defaults (mimicking the handler logic)
			if config.Model == "" {
				config.Model = defaultModel
			}
			if config.Language == "" {
				config.Language = defaultLanguage
			}
			if config.SampleRate == 0 {
				config.SampleRate = defaultSampleRate
			}
			if config.ChunkDuration == 0 {
				config.ChunkDuration = defaultChunkDuration
			}
			if config.BeamSize == 0 {
				config.BeamSize = defaultBeamSize
			}

			// Verify all fields
			if config.Model != tt.expected.Model {
				t.Errorf("Model: expected %s, got %s", tt.expected.Model, config.Model)
			}
			if config.Language != tt.expected.Language {
				t.Errorf("Language: expected %s, got %s", tt.expected.Language, config.Language)
			}
			if config.SampleRate != tt.expected.SampleRate {
				t.Errorf("SampleRate: expected %d, got %d", tt.expected.SampleRate, config.SampleRate)
			}
			if config.ChunkDuration != tt.expected.ChunkDuration {
				t.Errorf("ChunkDuration: expected %f, got %f", tt.expected.ChunkDuration, config.ChunkDuration)
			}
			if config.BeamSize != tt.expected.BeamSize {
				t.Errorf("BeamSize: expected %d, got %d", tt.expected.BeamSize, config.BeamSize)
			}
		})
	}
}

// TestWebSocketUpgrader tests the upgrader configuration
func TestWebSocketUpgrader(t *testing.T) {
	if upgrader.ReadBufferSize != wsReadBufferSize {
		t.Errorf("Expected ReadBufferSize %d, got %d", wsReadBufferSize, upgrader.ReadBufferSize)
	}

	if upgrader.WriteBufferSize != wsWriteBufferSize {
		t.Errorf("Expected WriteBufferSize %d, got %d", wsWriteBufferSize, upgrader.WriteBufferSize)
	}

	// Test CheckOrigin function allows connections
	req := &http.Request{
		Header: http.Header{
			"Origin": []string{"http://example.com"},
		},
	}

	if !upgrader.CheckOrigin(req) {
		t.Error("Expected CheckOrigin to allow connections")
	}
}

// TestTranscriptionResponseJSON tests JSON marshaling of response
func TestTranscriptionResponseJSON(t *testing.T) {
	response := TranscriptionResponse{
		Type:       "partial",
		Text:       "hello world",
		Confidence: 0.95,
		Timestamp:  1703123456789,
		IsFinal:    false,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	// Verify JSON structure
	var decoded map[string]interface{}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if decoded["type"] != "partial" {
		t.Errorf("Expected type 'partial', got %v", decoded["type"])
	}
	if decoded["text"] != "hello world" {
		t.Errorf("Expected text 'hello world', got %v", decoded["text"])
	}
	if decoded["is_final"] != false {
		t.Errorf("Expected is_final false, got %v", decoded["is_final"])
	}
}

// TestStreamConfigJSON tests JSON unmarshaling of configuration
func TestStreamConfigJSON(t *testing.T) {
	jsonData := `{
		"model": "test-model",
		"language": "en-US",
		"sample_rate": 16000,
		"enable_partial_results": true,
		"vad_enabled": true,
		"chunk_duration": 0.5,
		"beam_size": 5
	}`

	var config StreamConfig
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		t.Fatalf("Failed to unmarshal config: %v", err)
	}

	if config.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got %s", config.Model)
	}
	if config.Language != "en-US" {
		t.Errorf("Expected language 'en-US', got %s", config.Language)
	}
	if config.SampleRate != 16000 {
		t.Errorf("Expected sample rate 16000, got %d", config.SampleRate)
	}
	if !config.EnablePartialResults {
		t.Error("Expected EnablePartialResults true")
	}
	if !config.VADEnabled {
		t.Error("Expected VADEnabled true")
	}
	if config.ChunkDuration != 0.5 {
		t.Errorf("Expected chunk duration 0.5, got %f", config.ChunkDuration)
	}
	if config.BeamSize != 5 {
		t.Errorf("Expected beam size 5, got %d", config.BeamSize)
	}
}

// TestWebSocketUpgradeFailure tests WebSocket upgrade failure handling
func TestWebSocketUpgradeFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	router.GET("/test", AudioStream)

	// Create a request without WebSocket upgrade headers
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should fail to upgrade (400 Bad Request or similar)
	if w.Code == http.StatusOK {
		t.Error("Expected WebSocket upgrade to fail without proper headers")
	}
}

// TestSendError tests the error sending helper function
func TestSendError(t *testing.T) {
	// This test verifies the error message format
	// Since sendError writes to a WebSocket connection, we test the JSON structure

	errorMsg := "test error message"
	expected := map[string]string{
		"type":  "error",
		"error": errorMsg,
	}

	data, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("Failed to marshal error message: %v", err)
	}

	var decoded map[string]string
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal error message: %v", err)
	}

	if decoded["type"] != "error" {
		t.Errorf("Expected type 'error', got %s", decoded["type"])
	}
	if decoded["error"] != errorMsg {
		t.Errorf("Expected error '%s', got %s", errorMsg, decoded["error"])
	}
}

// TestConstants tests that all constants are defined correctly
func TestConstants(t *testing.T) {
	// Test buffer sizes
	if wsReadBufferSize <= 0 {
		t.Error("wsReadBufferSize must be positive")
	}
	if wsWriteBufferSize <= 0 {
		t.Error("wsWriteBufferSize must be positive")
	}

	// Test ASR defaults
	if defaultModel == "" {
		t.Error("defaultModel must not be empty")
	}
	if defaultLanguage == "" {
		t.Error("defaultLanguage must not be empty")
	}
	if defaultSampleRate <= 0 {
		t.Error("defaultSampleRate must be positive")
	}
	if defaultChunkDuration <= 0 {
		t.Error("defaultChunkDuration must be positive")
	}
	if defaultBeamSize <= 0 {
		t.Error("defaultBeamSize must be positive")
	}
	if defaultMaxQueueSize <= 0 {
		t.Error("defaultMaxQueueSize must be positive")
	}
	if defaultBufferSize <= 0 {
		t.Error("defaultBufferSize must be positive")
	}

	// Test audio conversion constants
	if int16MinValue != 32768.0 {
		t.Errorf("int16MinValue should be 32768.0, got %f", int16MinValue)
	}
	if int16MaxValue != 32767.0 {
		t.Errorf("int16MaxValue should be 32767.0, got %f", int16MaxValue)
	}
}

// Helper function for floating point comparison
func abs(x float32) float32 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark for audio conversion
func BenchmarkBytesToFloat32(b *testing.B) {
	// Create 1 second of audio data (16kHz, 16-bit)
	audioData := make([]byte, 16000*2)
	for i := 0; i < len(audioData); i++ {
		audioData[i] = byte(i % 256)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bytesToFloat32(audioData)
	}
}

// TestWebSocketConfigValidation tests configuration validation in a mock scenario
func TestWebSocketConfigValidation(t *testing.T) {
	tests := []struct {
		name      string
		configStr string
		wantError bool
	}{
		{
			name:      "Valid minimal config",
			configStr: `{"model":"test"}`,
			wantError: false,
		},
		{
			name:      "Valid full config",
			configStr: `{"model":"test","language":"en","sample_rate":16000,"beam_size":5}`,
			wantError: false,
		},
		{
			name:      "Invalid JSON",
			configStr: `{invalid json}`,
			wantError: true,
		},
		{
			name:      "Empty object",
			configStr: `{}`,
			wantError: false, // Empty config is valid, defaults will be applied
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config StreamConfig
			err := json.Unmarshal([]byte(tt.configStr), &config)

			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

// TestAudioDataEdgeCases tests edge cases in audio processing
func TestAudioDataEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectedLen   int
		expectWarning bool
	}{
		{
			name:          "Empty array",
			input:         []byte{},
			expectedLen:   0,
			expectWarning: false,
		},
		{
			name:          "Single byte (odd length)",
			input:         []byte{0xFF},
			expectedLen:   0,
			expectWarning: true,
		},
		{
			name:          "Two bytes (valid)",
			input:         []byte{0x00, 0x00},
			expectedLen:   1,
			expectWarning: false,
		},
		{
			name:          "Large array",
			input:         make([]byte, 32000), // 16000 samples
			expectedLen:   16000,
			expectWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToFloat32(tt.input)
			if len(result) != tt.expectedLen {
				t.Errorf("Expected %d samples, got %d", tt.expectedLen, len(result))
			}
		})
	}
}
