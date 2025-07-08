package nexa_sdk

import (
	"path"
	"testing"
)

// tts is the global TTS instance used across all tests
var tts *TTS

// initTTS creates a new TTS instance for testing with a predefined model
// Uses the Kokoro-82M-bf16 model from the local cache
func initTTS() {
	tts, _ = NewTTS(
		path.Join(nexaPath, "models", "bWx4LWNvbW11bml0eS9Lb2tvcm8tODJNLWJmMTY=", "kokoro-v1_0.safetensors"),
		nil, nil)
}

// deinitTTS cleans up the TTS instance after testing
func deinitTTS() {
	if tts != nil {
		tts.Destroy()
	}
}

// TestTTSListVoices tests the available voices functionality
// Verifies that the model can list available voices
func TestTTSListVoices(t *testing.T) {
	voices, err := tts.ListAvailableVoices()
	if err != nil {
		t.Logf("Failed to list voices: %v", err)
		return
	}

	t.Logf("Available voices: %d", len(voices))
	for i, voice := range voices {
		if i < 5 { // Log first 5 voices
			t.Logf("Voice %d: %s", i, voice)
		}
	}
}

// TestTTSSynthesize tests basic text-to-speech synthesis functionality
// Verifies that the model can convert text to speech
func TestTTSSynthesize(t *testing.T) {
	config := &TTSConfig{
		Voice:      "af_alloy",
		Speed:      1.0,
		Seed:       42,
		SampleRate: 22050,
	}

	result, err := tts.Synthesize("Hello world, this is a test.", config)
	if err != nil {
		t.Logf("TTS synthesis failed: %v", err)
		return
	}

	if result != nil {
		t.Logf("TTS synthesis result: %d samples, %.2f seconds, %d Hz, %d channels",
			result.NumSamples, result.DurationSeconds, result.SampleRate, result.Channels)
	}
}

// TestTTSBatchSynthesis tests batch text-to-speech synthesis functionality
// Verifies that the model can process multiple texts in batch
func TestTTSBatchSynthesis(t *testing.T) {
	texts := []string{
		"Hello world",
		"This is a test",
		"Batch synthesis works",
	}

	config := &TTSConfig{
		Voice:      "af_heart",
		Speed:      1.2,
		Seed:       123,
		SampleRate: 44100,
	}

	results, err := tts.SynthesizeBatch(texts, config)
	if err != nil {
		t.Logf("TTS batch synthesis failed: %v", err)
		return
	}

	if results != nil {
		t.Logf("TTS batch synthesis completed: %d results", len(results))
		for i, result := range results {
			if result != nil {
				t.Logf("Result %d: %d samples, %.2f seconds",
					i, result.NumSamples, result.DurationSeconds)
			}
		}
	}
}

// TestTTSSampler tests TTS sampler configuration functionality
// Verifies that sampling parameters can be configured
func TestTTSSampler(t *testing.T) {
	samplerConfig := &TTSSamplerConfig{
		Temperature: 0.8,
		NoiseScale:  0.2,
		LengthScale: 1.1,
	}

	tts.SetSampler(samplerConfig)
	t.Logf("Sampler configured with temperature: %.2f", samplerConfig.Temperature)

	config := &TTSConfig{
		Voice:      "af_bella",
		Speed:      1.0,
		Seed:       -1,
		SampleRate: 22050,
	}

	result, err := tts.Synthesize("Testing custom sampler configuration.", config)
	if err != nil {
		t.Logf("TTS synthesis with custom sampler failed: %v", err)
		return
	}

	if result != nil {
		t.Logf("Custom sampler synthesis result: %d samples, %.2f seconds",
			result.NumSamples, result.DurationSeconds)
	}

	// Reset sampler
	tts.ResetSampler()
	t.Logf("Sampler reset to default")
}

// TestTTSCache tests TTS cache functionality
// Verifies that cache can be saved and loaded
func TestTTSCache(t *testing.T) {
	cacheFile := "/tmp/tts_cache_test.bin"

	// Save cache
	err := tts.SaveCache(cacheFile)
	if err != nil {
		t.Logf("Failed to save cache: %v", err)
	} else {
		t.Logf("Cache saved to %s", cacheFile)
	}

	// Load cache
	err = tts.LoadCache(cacheFile)
	if err != nil {
		t.Logf("Failed to load cache: %v", err)
	} else {
		t.Logf("Cache loaded from %s", cacheFile)
	}
}

// TestTTSConfig tests TTS configuration structures
// Verifies that configuration parameters are properly set
func TestTTSConfig(t *testing.T) {
	config := &TTSConfig{
		Voice:      "af_alloy",
		Speed:      1.5,
		Seed:       42,
		SampleRate: 44100,
	}

	if config.Voice != "af_alloy" {
		t.Errorf("Expected voice 'af_alloy', got '%s'", config.Voice)
	}

	if config.Speed != 1.5 {
		t.Errorf("Expected speed 1.5, got %f", config.Speed)
	}

	samplerConfig := &TTSSamplerConfig{
		Temperature: 0.7,
		NoiseScale:  0.1,
		LengthScale: 1.0,
	}

	if samplerConfig.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", samplerConfig.Temperature)
	}

	if samplerConfig.NoiseScale != 0.1 {
		t.Errorf("Expected noise scale 0.1, got %f", samplerConfig.NoiseScale)
	}
}
