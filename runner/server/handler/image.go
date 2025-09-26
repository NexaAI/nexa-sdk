package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

// Use OpenAI's provided structs from openai-go library

// @Router			/images/generations [post]
// @Summary		Creates an image given a prompt.
// @Description	Creates an image given a prompt. This endpoint follows OpenAI DALL-E 3 API specification for compatibility.
// @Accept			json
// @Param			request	body	openai.ImageGenerateParams	true	"Image generation request"
// @Produce		json
// @Success		200	{object}	openai.ImagesResponse	"Successful image generation response"
// @Failure		400	{object}	map[string]any	"Bad request - invalid parameters"
// @Failure		404	{object}	map[string]any	"Model not found"
// @Failure		500	{object}	map[string]any	"Internal server error"
func ImageGenerations(c *gin.Context) {
	param := openai.ImageGenerateParams{}
	if err := c.ShouldBindJSON(&param); err != nil {
		slog.Error("Failed to bind JSON request", "error", err)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	slog.Info("Image generation request received",
		"model", param.Model,
		"prompt_length", len(param.Prompt),
		"n", param.N,
		"size", param.Size)

	if param.N.Value == 0 {
		param.N.Value = 1
	}
	if param.Size == "" {
		param.Size = openai.ImageGenerateParamsSize256x256
	}
	if param.ResponseFormat == "" {
		param.ResponseFormat = openai.ImageGenerateParamsResponseFormatURL
	}

	s := store.Get()
	manifest, err := s.GetManifest(param.Model)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, map[string]any{"error": "model not found"})
		} else {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
		return
	}

	if manifest.ModelType != types.ModelTypeImageGen {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model does not support image generation"})
		return
	}

	imageGen, err := service.KeepAliveGet[nexa_sdk.ImageGen](
		param.Model,
		types.ModelParam{},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	width, height, err := parseImageSize(string(param.Size))
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	var imageData []openai.Image
	n := int(param.N.Value)
	slog.Info("Starting image generation", "count", n, "size", string(param.Size))
	for i := range n {
		outputPath := fmt.Sprintf("imagegen_output_%d.png", time.Now().UnixNano())
		slog.Debug("Generating image", "index", i, "output_path", outputPath)

		config := &nexa_sdk.ImageGenerationConfig{
			Prompts:         []string{param.Prompt},
			NegativePrompts: []string{"blurry, low quality, distorted, low resolution"},
			Height:          height,
			Width:           width,
			SamplerConfig: nexa_sdk.ImageSamplerConfig{
				Method:        "ddim",
				Steps:         20,
				GuidanceScale: 7.5,
				Eta:           0.0,
				Seed:          int32(time.Now().UnixNano() % 1000000),
			},
			SchedulerConfig: nexa_sdk.SchedulerConfig{
				Type:              "ddim",
				NumTrainTimesteps: 1000,
				StepsOffset:       1,
				BetaStart:         0.00085,
				BetaEnd:           0.012,
				BetaSchedule:      "scaled_linear",
				PredictionType:    "epsilon",
				TimestepType:      "discrete",
				TimestepSpacing:   "leading",
				InterpolationType: "linear",
				ConfigPath:        "",
			},
			Strength: 1.0,
		}

		result, err := imageGen.Txt2Img(nexa_sdk.ImageGenTxt2ImgInput{
			PromptUTF8: param.Prompt,
			Config:     config,
			OutputPath: outputPath,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("image generation failed: %v", err)})
			return
		}

		data := openai.Image{
			RevisedPrompt: param.Prompt,
		}

		if param.ResponseFormat == openai.ImageGenerateParamsResponseFormatB64JSON {
			b64Data, err := encodeImageToBase64(result.OutputImagePath)
			os.Remove(result.OutputImagePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("failed to encode image: %v", err)})
				return
			}
			data.B64JSON = b64Data
		} else {
			data.URL = result.OutputImagePath
		}

		imageData = append(imageData, data)
		slog.Info("Image generated successfully", "index", i, "output_path", result.OutputImagePath)
	}

	response := openai.ImagesResponse{
		Created: time.Now().Unix(),
		Data:    imageData,
	}

	slog.Info("Image generation completed successfully", "total_images", len(imageData))
	c.JSON(http.StatusOK, response)
}

func parseImageSize(size string) (int32, int32, error) {
	parts := strings.Split(size, "x")
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid size format")
	}

	width, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, errors.New("invalid width")
	}

	height, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, errors.New("invalid height")
	}

	return int32(width), int32(height), nil
}

func encodeImageToBase64(imagePath string) (string, error) {
	imageData, err := os.ReadFile(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %v", err)
	}
	mimeType := http.DetectContentType(imageData)
	base64String := base64.StdEncoding.EncodeToString(imageData)
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64String), nil
}
