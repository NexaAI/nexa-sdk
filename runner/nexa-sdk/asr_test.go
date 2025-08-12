package nexa_sdk

import (
	"log/slog"
	"testing"
)

var asr *ASR

func initASR() {
	slog.Debug("initASR called")

	var err error

	input := AsrCreateInput{
		ModelPath: "modelfiles/mlx/whisper-tiny-mlx-q4/weights.npz",
		PluginID:  "mlx",
	}

	asr, err = NewASR(input)
	if err != nil {
		panic("Error creating ASR: " + err.Error())
	}
}

func deinitASR() {
	asr.Destroy()
}

func TestASRListSupportedLanguages(t *testing.T) {
	output, err := asr.ListSupportedLanguages()
	if err != nil {
		t.Errorf("ListSupportedLanguages failed: %v", err)
		return
	}

	t.Logf("Supported languages: %v", output.LanguageCodes)
	t.Logf("Language count: %d", len(output.LanguageCodes))
}

func TestASRTranscribe(t *testing.T) {
	input := AsrTranscribeInput{
		AudioPath: "modelfiles/assets/OSR_us_000_0010_8k.wav",
		Language:  "en",
		Config: &ASRConfig{
			Timestamps: "word",
			BeamSize:   5,
			Stream:     false,
		},
	}

	output, err := asr.Transcribe(input)
	if err != nil {
		t.Fatalf("Transcribe failed: %v", err)
	}

	t.Logf("Transcribe completed successfully")
	t.Logf("Transcript: %s", output.Result.Transcript)

	if len(output.Result.ConfidenceScores) > 0 {
		t.Logf("Average confidence: %.2f", averageFloat32(output.Result.ConfidenceScores))
	}
}

func TestASRTranscribeWithoutConfig(t *testing.T) {
	input := AsrTranscribeInput{
		AudioPath: "modelfiles/assets/OSR_us_000_0010_8k.wav",
		Language:  "en",
		Config:    nil, // Use default configuration
	}

	output, err := asr.Transcribe(input)
	if err != nil {
		t.Fatalf("Transcribe without config failed: %v", err)
	}

	t.Logf("Transcribe without config completed successfully")
	t.Logf("Transcript: %s", output.Result.Transcript)
}

func TestASRTranscribeWithAutoLanguage(t *testing.T) {
	input := AsrTranscribeInput{
		AudioPath: "modelfiles/assets/OSR_us_000_0010_8k.wav",
		Language:  "", // Auto-detect language
		Config: &ASRConfig{
			Timestamps: "segment",
			BeamSize:   3,
			Stream:     false,
		},
	}

	output, err := asr.Transcribe(input)
	if err != nil {
		t.Fatalf("Transcribe with auto language failed: %v", err)
	}

	t.Logf("Transcribe with auto language completed successfully")
	t.Logf("Transcript: %s", output.Result.Transcript)

	if len(output.Result.Timestamps) > 0 {
		t.Logf("First timestamp: %.2f", output.Result.Timestamps[0])
		t.Logf("Last timestamp: %.2f", output.Result.Timestamps[len(output.Result.Timestamps)-1])
	}
}

func TestASRTranscribeWithStreaming(t *testing.T) {
	input := AsrTranscribeInput{
		AudioPath: "modelfiles/assets/OSR_us_000_0010_8k.wav",
		Language:  "en",
		Config: &ASRConfig{
			Timestamps: "word",
			BeamSize:   1,
			Stream:     true, // Enable streaming mode
		},
	}

	output, err := asr.Transcribe(input)
	if err != nil {
		t.Fatalf("Transcribe with streaming failed: %v", err)
	}

	t.Logf("Transcribe with streaming completed successfully")
	t.Logf("Transcript: %s", output.Result.Transcript)
}

func TestASRTranscribeLongAudio(t *testing.T) {
	input := AsrTranscribeInput{
		AudioPath: "./long_test_audio.wav",
		Language:  "en",
		Config: &ASRConfig{
			Timestamps: "segment",
			BeamSize:   5,
			Stream:     false,
		},
	}

	output, err := asr.Transcribe(input)
	if err != nil {
		t.Fatalf("Transcribe long audio failed: %v", err)
	}

	t.Logf("Transcribe long audio completed successfully")
	t.Logf("Transcript: %s", output.Result.Transcript)

	if len(output.Result.ConfidenceScores) > 0 {
		t.Logf("Min confidence: %.2f", minFloat32(output.Result.ConfidenceScores))
		t.Logf("Max confidence: %.2f", maxFloat32(output.Result.ConfidenceScores))
	}
}

func TestASRTranscribeDifferentLanguage(t *testing.T) {
	input := AsrTranscribeInput{
		AudioPath: "./chinese_test_audio.wav",
		Language:  "zh",
		Config: &ASRConfig{
			Timestamps: "word",
			BeamSize:   5,
			Stream:     false,
		},
	}

	output, err := asr.Transcribe(input)
	if err != nil {
		t.Fatalf("Transcribe different language failed: %v", err)
	}

	t.Logf("Transcribe different language completed successfully")
	t.Logf("Transcript: %s", output.Result.Transcript)
}

// Helper functions for float32 slice operations
func averageFloat32(slice []float32) float32 {
	if len(slice) == 0 {
		return 0
	}
	sum := float32(0)
	for _, v := range slice {
		sum += v
	}
	return sum / float32(len(slice))
}

func minFloat32(slice []float32) float32 {
	if len(slice) == 0 {
		return 0
	}
	min := slice[0]
	for _, v := range slice {
		if v < min {
			min = v
		}
	}
	return min
}

func maxFloat32(slice []float32) float32 {
	if len(slice) == 0 {
		return 0
	}
	max := slice[0]
	for _, v := range slice {
		if v > max {
			max = v
		}
	}
	return max
}
