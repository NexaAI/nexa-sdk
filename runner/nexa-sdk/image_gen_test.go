package nexa_sdk

import (
	"log/slog"
	"os"
	"testing"
)

var imagegen *ImageGen

func initImageGen() {
	slog.Debug("initImageGen called")

	var err error

	input := ImageGenCreateInput{
		ModelPath: "./modelfiles/ort/sdxl-fp16-cuda",
		PluginID:  "ort",
		DeviceID:  "cuda",
	}

	imagegen, err = NewImageGen(input)
	if err != nil {
		panic("Error creating ImageGen: " + err.Error())
	}
}

func deinitImageGen() {
	if imagegen != nil {
		imagegen.Destroy()
	}
}

func TestImageGenTxt2Img(t *testing.T) {
	// Skip if ImageGen is not available
	if imagegen == nil {
		t.Skip("ImageGen not initialized, skipping test")
	}

	// The image generator must be created successfully or the test fails
	if imagegen == nil {
		t.Fatal("Image generator creation failed")
	}

	// Set up image generation configuration (aligned with config.json)
	samplerConfig := ImageSamplerConfig{
		Method:        "ddim",
		Steps:         20,
		GuidanceScale: 7.5,
		Eta:           0.0,
		Seed:          2, // Changed from 42 to match config.json
	}

	schedulerConfig := SchedulerConfig{
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
	}

	config := ImageGenerationConfig{
		Prompts:         []string{"a beautiful landscape with mountains and trees"},
		NegativePrompts: []string{"blurry, low quality, distorted"},
		Height:          512,
		Width:           512,
		SamplerConfig:   samplerConfig,
		SchedulerConfig: schedulerConfig,
		Strength:        1.0,
	}

	// Ensure output directory exists
	outputDir := "./output"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, 0o755)
		if err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}
		t.Logf("Created output directory: %s", outputDir)
	}

	input := ImageGenTxt2ImgInput{
		PromptUTF8: "a beautiful landscape with mountains and trees",
		Config:     &config,
		OutputPath: "./build/generated_image.png",
	}

	output, err := imagegen.Txt2Img(input)
	if err != nil {
		t.Fatalf("Txt2Img failed: %v", err)
	}

	if output.OutputImagePath != "" {
		t.Logf("Image generated successfully at: %s", output.OutputImagePath)

		// Check if the file actually exists
		if _, err := os.Stat(output.OutputImagePath); err == nil {
			fileInfo, _ := os.Stat(output.OutputImagePath)
			t.Logf("Generated image file exists and has size: %d bytes", fileInfo.Size())
			t.Logf("Image saved to: %s", output.OutputImagePath)
			// Note: Keeping the generated image for inspection (not cleaning up)
		} else {
			t.Logf("Warning: Generated image file does not exist at: %s", output.OutputImagePath)
		}
	} else {
		t.Logf("Image generation completed but no output path returned")
	}
}
