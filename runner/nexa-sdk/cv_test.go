package nexa_sdk

import (
	"log/slog"
	"os"
	"testing"
)

var cv *CV

func initCV() {
	slog.Debug("initCV called")

	var err error

	input := CVCreateInput{
		Config: CVModelConfig{
			Capabilities: CVCapabilityOCR,
			RecModelPath: "modelfiles/mlx/paddle-ocr-ml/ch_ptocr_v4_rec_infer_f16.safetensors",
			DetModelPath: "modelfiles/mlx/paddle-ocr-ml/ch_ptocr_v4_det_infer_f16.safetensors",
			CharDictPath: "",
		},
		PluginID: "mlx",
		DeviceID: "",
	}

	cv, err = NewCV(input)
	if err != nil {
		panic("Error creating CV: " + err.Error())
	}
}

func deinitCV() {
	cv.Destroy()
}

func TestCVInferBasic(t *testing.T) {
	testImagePath := "modelfiles/assets/test_image.png"

	// Check if test image file exists
	if _, err := os.Stat(testImagePath); os.IsNotExist(err) {
		t.Skipf("Test image file not found: %s", testImagePath)
		return
	}

	t.Logf("Testing CV inference with image: %s", testImagePath)

	input := CVInferInput{
		InputImagePath: testImagePath,
	}

	output, err := cv.Infer(input)
	if err != nil {
		t.Fatalf("CV inference failed: %v", err)
	}

	// Verify results structure
	resultCount := len(output.Results)
	if resultCount < 0 {
		t.Errorf("Invalid result count: %d", resultCount)
		return
	}

	t.Logf("Got %d CV results", resultCount)

	if resultCount > 0 {

		// Process each result
		for i, result := range output.Results {
			// Verify confidence score is in valid range
			if result.Confidence < 0.0 || result.Confidence > 1.0 {
				t.Errorf("Result %d: confidence score out of range [0,1]: %f", i, result.Confidence)
			}

			t.Logf("Result %d: confidence=%.3f", i, result.Confidence)

			// Check text if present (for OCR)
			if result.Text != "" {
				t.Logf("  Text: %s", result.Text)
			}

			// Check class_id if present (for classification)
			if result.ClassID >= 0 {
				t.Logf("  Class ID: %d", result.ClassID)
			}

			// Check bounding box if present (for detection)
			if result.BBox.Width > 0 && result.BBox.Height > 0 {
				t.Logf("  Bounding Box: [%.2f, %.2f, %.2f, %.2f]",
					result.BBox.X, result.BBox.Y, result.BBox.Width, result.BBox.Height)

				// Verify bounding box coordinates are reasonable
				if result.BBox.X < 0.0 {
					t.Errorf("Result %d: invalid bbox x coordinate: %f", i, result.BBox.X)
				}
				if result.BBox.Y < 0.0 {
					t.Errorf("Result %d: invalid bbox y coordinate: %f", i, result.BBox.Y)
				}
				if result.BBox.Width <= 0.0 {
					t.Errorf("Result %d: invalid bbox width: %f", i, result.BBox.Width)
				}
				if result.BBox.Height <= 0.0 {
					t.Errorf("Result %d: invalid bbox height: %f", i, result.BBox.Height)
				}
			}

			// Check embedding if present (for feature extraction)
			if len(result.Embedding) > 0 && result.EmbeddingDim > 0 {
				t.Logf("  Embedding dimension: %d", result.EmbeddingDim)

				// Verify embedding values are reasonable
				for j := 0; j < min(5, len(result.Embedding)); j++ {
					if isNaN(result.Embedding[j]) || isInf(result.Embedding[j]) {
						t.Errorf("Result %d: invalid embedding value at index %d: %f", i, j, result.Embedding[j])
					}
				}
			}
		}
	} else {
		t.Logf("No CV results found - this may be normal for some models")
	}
}

func isNaN(f float32) bool {
	return f != f
}

func isInf(f float32) bool {
	return f > 1e38 || f < -1e38
}
