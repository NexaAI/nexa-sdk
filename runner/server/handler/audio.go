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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

func Speech(c *gin.Context) {
	param := openai.AudioSpeechNewParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}
	slog.Info("Speech request received",
		"model", param.Model,
		"input", param.Input,
		"voice", param.Voice,
		"speed", param.Speed,
	)

	audioSpeech, err := service.KeepAliveGet[nexa_sdk.TTS](
		param.Model,
		types.ModelParam{},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	// warm up
	if param.Input == "" {
		c.JSON(http.StatusOK, nil)
		return
	}

	outputPath := fmt.Sprintf("audio_speech_output_%d.wav", time.Now().UnixNano())
	defer os.Remove(outputPath)
	_, err = audioSpeech.Synthesize(
		nexa_sdk.TtsSynthesizeInput{
			TextUTF8: param.Input,
			Config: &nexa_sdk.TTSConfig{
				Voice: string(param.Voice),
				Speed: float32(param.Speed.Value),
			},
			OutputPath: outputPath,
		})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	c.File(outputPath)
}

func Transcriptions(c *gin.Context) {
	param := openai.AudioTranscriptionNewParams{}
	param.Model = c.PostForm("model")
	stream := c.PostForm("stream")

	if stream == "true" {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "streaming not supported"})
		return
	}

	slog.Info("Transcriptions request received",
		"model", param.Model,
		"stream", stream,
	)

	p, err := service.KeepAliveGet[nexa_sdk.ASR](
		string(param.Model),
		types.ModelParam{},
		false,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	// retrieve file from form data
	file, err := c.FormFile("file")
	if err != nil {
		if err == http.ErrMissingFile {
			// warm up
			c.JSON(http.StatusOK, nil)
			return
		}
		c.JSON(http.StatusBadRequest, map[string]any{"error": "failed to get file: " + err.Error()})
		return
	}
	param.File, err = file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "failed to open file: " + err.Error()})
		return
	}
	data, err := io.ReadAll(param.File)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "failed to read file: " + err.Error()})
		return
	}

	// write data to a temp file
	tmpFile, err := os.CreateTemp("", "uri-*"+path.Ext(file.Filename))
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to create temp file: " + err.Error()})
		return
	}
	_, err = tmpFile.Write(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": "failed to write temp file: " + err.Error()})
		return
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	res, err := p.Transcribe(nexa_sdk.AsrTranscribeInput{
		AudioPath: tmpFile.Name(),
	})
	result := openai.Transcription{
		Text: res.Result.Transcript,
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
	} else {
		c.JSON(http.StatusOK, result)
	}
}

type DiarizeRequest struct {
	Model string `json:"model" binding:"required"`
	Audio string `json:"audio"`
}

func Diarize(c *gin.Context) {
	param := DiarizeRequest{}
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	slog.Info("Diarize request received",
		"model", param.Model,
		"audio", param.Audio,
	)

	p, err := service.KeepAliveGet[nexa_sdk.Diarize](
		string(param.Model),
		types.ModelParam{},
		false,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
		return
	}

	// warm up
	if param.Audio == "" {
		c.JSON(http.StatusOK, nil)
		return
	}

	file, err := utils.SaveURIToTempFile(param.Audio)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "failed to save audio: " + err.Error()})
		return
	}
	defer os.Remove(file)
	res, err := p.Infer(nexa_sdk.DiarizeInferInput{
		AudioPath: file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error(), "code": nexa_sdk.SDKErrorCode(err)})
	} else {
		c.JSON(http.StatusOK, res)
	}
}

const (
	// WebSocket buffer sizes
	wsReadBufferSize  = 1024
	wsWriteBufferSize = 1024

	// ASR streaming configuration defaults
	defaultModel         = "NexaAI/parakeet-npu"
	defaultLanguage      = "en-US"
	defaultSampleRate    = 16000
	defaultChunkDuration = 0.5 // seconds
	defaultBeamSize      = 5
	defaultMaxQueueSize  = 100
	defaultBufferSize    = 4096

	// Audio conversion constants
	// For 16-bit PCM audio: negative values use 32768, positive values use 32767
	// to maintain symmetry and proper range
	int16MinValue = 32768.0 // Minimum value magnitude for signed 16-bit integers (2^15)
	int16MaxValue = 32767.0 // Maximum value for signed 16-bit integers (2^15 - 1)
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  wsReadBufferSize,
	WriteBufferSize: wsWriteBufferSize,
	CheckOrigin: func(r *http.Request) bool {
		// SECURITY WARNING: This currently allows all origins for development.
		// In production, implement proper origin validation:
		// 1. Maintain an allowlist of trusted origins
		// 2. Check r.Header.Get("Origin") against the allowlist
		// 3. Return true only for allowed origins
		// Example production code:
		//   origin := r.Header.Get("Origin")
		//   return origin == "https://yourdomain.com" || origin == "https://app.yourdomain.com"
		return true
	},
}

// StreamConfig represents the configuration for ASR streaming
type StreamConfig struct {
	Model                string  `json:"model"`
	Language             string  `json:"language"`
	SampleRate           int32   `json:"sample_rate"`
	EnablePartialResults bool    `json:"enable_partial_results"`
	EnableWordTimestamps bool    `json:"enable_word_timestamps"`
	VADEnabled           bool    `json:"vad_enabled"`
	ChunkDuration        float32 `json:"chunk_duration"`
	OverlapDuration      float32 `json:"overlap_duration"`
	BeamSize             int32   `json:"beam_size"`
}

// TranscriptionResponse represents the response sent back to the client
type TranscriptionResponse struct {
	Type       string  `json:"type"` // "partial" or "final"
	Text       string  `json:"text"`
	Confidence float64 `json:"confidence,omitempty"`
	Timestamp  int64   `json:"timestamp"`
	IsFinal    bool    `json:"is_final"`
}

// AudioStream handles WebSocket connections for real-time ASR
func AudioStream(c *gin.Context) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error("Failed to upgrade to WebSocket", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("WebSocket connection established")

	// Read configuration message (expected as first message)
	_, configMsg, err := conn.ReadMessage()
	if err != nil {
		slog.Error("Failed to read config message", "error", err)
		sendError(conn, "Failed to read configuration")
		return
	}

	var config StreamConfig
	if err := json.Unmarshal(configMsg, &config); err != nil {
		slog.Error("Failed to parse config", "error", err)
		sendError(conn, "Invalid configuration format")
		return
	}

	// Validate and set defaults
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

	slog.Info("Stream config received",
		"model", config.Model,
		"language", config.Language,
		"sample_rate", config.SampleRate,
		"enable_partial_results", config.EnablePartialResults,
	)

	// Get ASR instance from service
	asr, err := service.KeepAliveGet[nexa_sdk.ASR](
		config.Model,
		types.ModelParam{},
		false,
	)
	if err != nil {
		slog.Error("Failed to get ASR instance", "error", err)
		sendError(conn, fmt.Sprintf("Failed to initialize ASR: %v", err))
		return
	}

	// Set up callback for transcription results
	var mu sync.Mutex
	transcriptionCallback := func(text string, userData any) {
		mu.Lock()
		defer mu.Unlock()

		response := TranscriptionResponse{
			Type:      "partial",
			Text:      text,
			Timestamp: time.Now().UnixMilli(),
			IsFinal:   false,
		}

		// If partial results are not enabled, mark as final
		if !config.EnablePartialResults {
			response.Type = "final"
			response.IsFinal = true
		}

		// WriteJSON is protected by mutex to prevent concurrent writes to the WebSocket connection
		if err := conn.WriteJSON(response); err != nil {
			slog.Error("Failed to send transcription", "error", err)
		}
	}

	// Initialize ASR streaming
	streamConfig := &nexa_sdk.ASRStreamConfig{
		ChunkDuration:   config.ChunkDuration,
		OverlapDuration: config.OverlapDuration,
		SampleRate:      config.SampleRate,
		MaxQueueSize:    defaultMaxQueueSize,
		BufferSize:      defaultBufferSize,
		BeamSize:        config.BeamSize,
	}

	_, err = asr.StreamBegin(nexa_sdk.AsrStreamBeginInput{
		StreamConfig:    streamConfig,
		Language:        config.Language,
		OnTranscription: transcriptionCallback,
	})
	if err != nil {
		slog.Error("Failed to begin ASR streaming", "error", err)
		sendError(conn, fmt.Sprintf("Failed to start streaming: %v", err))
		return
	}

	// Ensure we stop streaming when done
	defer func() {
		if err := asr.StreamStop(nexa_sdk.AsrStreamStopInput{Graceful: true}); err != nil {
			slog.Error("Failed to stop ASR streaming", "error", err)
		}
	}()

	// Process incoming audio data
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("WebSocket error", "error", err)
			} else {
				slog.Info("WebSocket closed", "error", err)
			}
			break
		}

		// Handle binary audio data
		if messageType == websocket.BinaryMessage {
			// Convert bytes to float32 array (assuming 16-bit PCM audio)
			audioData := bytesToFloat32(message)

			if len(audioData) > 0 {
				err = asr.StreamPushAudio(nexa_sdk.AsrStreamPushAudioInput{
					AudioData: audioData,
				})
				if err != nil {
					slog.Error("Failed to push audio data", "error", err)
					sendError(conn, fmt.Sprintf("Failed to process audio: %v", err))
					break
				}
			}
		} else if messageType == websocket.TextMessage {
			// Handle control messages (e.g., "stop", "pause")
			var controlMsg map[string]string
			if err := json.Unmarshal(message, &controlMsg); err == nil {
				if controlMsg["action"] == "stop" {
					slog.Info("Stop signal received from client")
					break
				}
			}
		}
	}

	slog.Info("WebSocket connection closed")
}

// bytesToFloat32 converts byte array (16-bit PCM) to float32 array
func bytesToFloat32(data []byte) []float32 {
	// Validate byte array length (must be even for 16-bit samples)
	if len(data)%2 != 0 {
		slog.Warn("Invalid audio data length (not even), truncating last byte", "length", len(data))
		data = data[:len(data)-1]
	}

	// Assuming 16-bit PCM audio (little-endian)
	numSamples := len(data) / 2
	result := make([]float32, numSamples)

	for i := 0; i < numSamples; i++ {
		// Read 16-bit sample
		sample := int16(binary.LittleEndian.Uint16(data[i*2 : i*2+2]))
		// Normalize to [-1.0, 1.0]
		// Use different divisors for negative and positive values for precise conversion
		if sample < 0 {
			result[i] = float32(sample) / int16MinValue
		} else {
			result[i] = float32(sample) / int16MaxValue
		}
	}

	return result
}

// sendError sends an error message to the WebSocket client
func sendError(conn *websocket.Conn, message string) {
	err := conn.WriteJSON(map[string]string{
		"type":  "error",
		"error": message,
	})
	if err != nil {
		slog.Error("Failed to send error message", "error", err)
	}
}
