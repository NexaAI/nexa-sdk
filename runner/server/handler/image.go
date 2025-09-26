package handler

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/NexaAI/nexa-sdk/runner/internal/store"
	"github.com/NexaAI/nexa-sdk/runner/internal/types"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
	"github.com/NexaAI/nexa-sdk/runner/server/service"
)

// ImageGenerationRequest represents the request body for image generation
// Following OpenAI DALL-E 3 API specification
type ImageGenerationRequest struct {
	Model          string `json:"model" binding:"required" example:"dall-e-3"`
	Prompt         string `json:"prompt" binding:"required" example:"A beautiful sunset over mountains"`
	N              int    `json:"n,omitempty" example:"1"`                 // Number of images to generate (1-10)
	Size           string `json:"size,omitempty" example:"1024x1024"`      // Image size: 1024x1024, 1792x1024, 1024x1792
	Quality        string `json:"quality,omitempty" example:"standard"`    // Quality: standard, hd
	Style          string `json:"style,omitempty" example:"vivid"`         // Style: vivid, natural
	ResponseFormat string `json:"response_format,omitempty" example:"url"` // Response format: url, b64_json
	User           string `json:"user,omitempty"`                          // User identifier for tracking
}

// ImageGenerationResponse represents the response for image generation
// Following OpenAI DALL-E 3 API specification
type ImageGenerationResponse struct {
	Created int64                 `json:"created"`
	Data    []ImageGenerationData `json:"data"`
}

// ImageGenerationData represents individual image data in the response
type ImageGenerationData struct {
	URL           string `json:"url,omitempty"`            // Image URL (when response_format is "url")
	B64JSON       string `json:"b64_json,omitempty"`       // Base64 encoded image (when response_format is "b64_json")
	RevisedPrompt string `json:"revised_prompt,omitempty"` // Revised prompt used for generation
}

// @Router			/images/generations [post]
// @Summary		Creates an image given a prompt.
// @Description	Creates an image given a prompt. This endpoint follows OpenAI DALL-E 3 API specification for compatibility.
// @Accept			json
// @Param			request	body	ImageGenerationRequest	true	"Image generation request"
// @Produce		json
// @Success		200	{object}	ImageGenerationResponse	"Successful image generation response"
// @Failure		400	{object}	map[string]any	"Bad request - invalid parameters"
// @Failure		404	{object}	map[string]any	"Model not found"
// @Failure		500	{object}	map[string]any	"Internal server error"
func ImageGenerations(c *gin.Context) {
	param := ImageGenerationRequest{}
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

	// Set default values for optional parameters
	setDefaultImageGenerationParams(&param)

	// Validate parameters
	if err := validateImageGenerationParams(param); err != nil {
		slog.Error("Invalid image generation parameters", "error", err, "params", param)
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Get model manifest
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

	// Check if model supports image generation
	if manifest.ModelType != types.ModelTypeImageGen {
		c.JSON(http.StatusBadRequest, map[string]any{"error": "model does not support image generation"})
		return
	}

	// Get ImageGen instance
	imageGen, err := service.KeepAliveGet[nexa_sdk.ImageGen](
		param.Model,
		types.ModelParam{},
		c.GetHeader("Nexa-KeepCache") != "true",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}

	// Parse image size
	width, height, err := parseImageSize(param.Size)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		return
	}

	// Generate images
	var imageData []ImageGenerationData
	slog.Info("Starting image generation", "count", param.N, "size", param.Size)

	for i := 0; i < param.N; i++ {
		// Generate unique output filename
		outputPath := fmt.Sprintf("imagegen_output_%d.png", time.Now().Unix())
		slog.Debug("Generating image", "index", i, "output_path", outputPath)

		// Create image generation configuration
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
				Seed:          int32(time.Now().UnixNano() % 1000000), // Random seed
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

		// Generate image
		result, err := imageGen.Txt2Img(nexa_sdk.ImageGenTxt2ImgInput{
			PromptUTF8: param.Prompt,
			Config:     config,
			OutputPath: outputPath,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("image generation failed: %v", err)})
			return
		}

		// Create response data based on format
		data := ImageGenerationData{
			RevisedPrompt: param.Prompt, // In a real implementation, this might be different
		}

		if param.ResponseFormat == "b64_json" {
			// Read image file and encode as base64
			b64Data, err := encodeImageToBase64(result.OutputImagePath)
			if err != nil {
				c.JSON(http.StatusInternalServerError, map[string]any{"error": fmt.Sprintf("failed to encode image: %v", err)})
				return
			}
			data.B64JSON = b64Data
		} else {
			// Return file path as URL (in production, this should be a proper URL)
			data.URL = result.OutputImagePath
		}

		imageData = append(imageData, data)
		slog.Info("Image generated successfully", "index", i, "output_path", result.OutputImagePath)
	}

	// Create response
	response := ImageGenerationResponse{
		Created: time.Now().Unix(),
		Data:    imageData,
	}

	slog.Info("Image generation completed successfully", "total_images", len(imageData))
	c.JSON(http.StatusOK, response)
}

// setDefaultImageGenerationParams sets default values for optional parameters
func setDefaultImageGenerationParams(param *ImageGenerationRequest) {
	// Set default number of images
	if param.N == 0 {
		param.N = 1
	}

	// Set default size
	if param.Size == "" {
		param.Size = "512x512"
	}

	// Set default quality
	if param.Quality == "" {
		param.Quality = "standard"
	}

	// Set default style
	if param.Style == "" {
		param.Style = "vivid"
	}

	// Set default response format
	if param.ResponseFormat == "" {
		param.ResponseFormat = "url"
	}
}

// validateImageGenerationParams validates the request parameters
func validateImageGenerationParams(param ImageGenerationRequest) error {
	// Validate N (number of images)
	if param.N < 1 || param.N > 10 {
		return errors.New("n must be between 1 and 10")
	}

	// Validate size
	validSizes := []string{"512x512", "1024x1024", "1792x1024", "1024x1792"}
	if param.Size != "" {
		valid := slices.Contains(validSizes, param.Size)
		if !valid {
			return errors.New("size must be one of: 512x512, 1024x1024, 1792x1024, 1024x1792")
		}
	}

	// Validate quality
	if param.Quality != "" && param.Quality != "standard" && param.Quality != "hd" {
		return errors.New("quality must be 'standard' or 'hd'")
	}

	// Validate style
	if param.Style != "" && param.Style != "vivid" && param.Style != "natural" {
		return errors.New("style must be 'vivid' or 'natural'")
	}

	// Validate response format
	if param.ResponseFormat != "" && param.ResponseFormat != "url" && param.ResponseFormat != "b64_json" {
		return errors.New("response_format must be 'url' or 'b64_json'")
	}

	return nil
}

// parseImageSize parses the size string and returns width and height
func parseImageSize(size string) (int32, int32, error) {
	if size == "" {
		size = "512x512" // Default size
	}

	// Parse size string (e.g., "1024x1024")
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

// encodeImageToBase64 reads an image file and encodes it as base64
func encodeImageToBase64(imagePath string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return "", fmt.Errorf("image file not found: %s", imagePath)
	}

	// Read the image file
	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to open image file: %v", err)
	}
	defer file.Close()

	// Read file content
	imageData, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %v", err)
	}

	// Encode to base64
	base64String := base64.StdEncoding.EncodeToString(imageData)

	// Determine MIME type based on file extension
	var mimeType string
	ext := strings.ToLower(filepath.Ext(imagePath))
	switch ext {
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".gif":
		mimeType = "image/gif"
	case ".webp":
		mimeType = "image/webp"
	default:
		mimeType = "image/png" // Default to PNG
	}

	// Return data URL format
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64String), nil
}
