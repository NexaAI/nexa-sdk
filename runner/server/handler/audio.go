// Copyright 2024-2026 Nexa AI, Inc.
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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/NexaAI/nexa-sdk/runner/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/v3"

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
