package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/NexaAI/nexa-sdk/runner/server/service"
	"github.com/gin-gonic/gin"
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

func Transcription() {

}
