package nexa_sdk

import (
	"log/slog"
	"os"
	"testing"
)

var tts *TTS

func initTTS() {
	slog.Debug("initTTS called")

	var err error

	input := TtsCreateInput{
		ModelPath:   "modelfiles/mlx/Kokoro-82M-4bit/kokoro-v1_0.safetensors",
		PluginID:    "mlx",
	}

	tts, err = NewTTS(input)
	if err != nil {
		panic("Error creating TTS: " + err.Error())
	}
}

func deinitTTS() {
	tts.Destroy()
}

func TestTTSListAvailableVoices(t *testing.T) {
	output, err := tts.ListAvailableVoices()
	if err != nil {
		t.Errorf("ListAvailableVoices failed: %v", err)
		return
	}

	t.Logf("Available voices: %v", output.VoiceIDs)
	t.Logf("Voice count: %d", len(output.VoiceIDs))
}

func TestTTSSynthesize(t *testing.T) {
	input := TtsSynthesizeInput{
		TextUTF8: "Hello, this is a test of text-to-speech synthesis.",
		Config: &TTSConfig{
			Voice:      "en_speaker_0",
			Speed:      1.0,
			Seed:       -1,
			SampleRate: 24000,
		},
		OutputPath: "./test_output.wav",
	}

	output, err := tts.Synthesize(input)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}

	t.Logf("Synthesize completed successfully")
	t.Logf("Audio path: %s", output.Result.AudioPath)
	t.Logf("Duration: %.2f seconds", output.Result.DurationSeconds)
	t.Logf("Sample rate: %d Hz", output.Result.SampleRate)
	t.Logf("Channels: %d", output.Result.Channels)
	t.Logf("Num samples: %d", output.Result.NumSamples)

	// Clean up test file
	if output.Result.AudioPath != "" {
		_ = os.Remove(output.Result.AudioPath)
	}
}

func TestTTSSynthesizeWithoutConfig(t *testing.T) {
	input := TtsSynthesizeInput{
		TextUTF8: "This is a test without specific configuration.",
		Config:   nil, // Use default configuration
	}

	output, err := tts.Synthesize(input)
	if err != nil {
		t.Fatalf("Synthesize without config failed: %v", err)
	}

	t.Logf("Synthesize without config completed successfully")
	t.Logf("Audio path: %s", output.Result.AudioPath)
	t.Logf("Duration: %.2f seconds", output.Result.DurationSeconds)

	// Clean up test file
	if output.Result.AudioPath != "" {
		_ = os.Remove(output.Result.AudioPath)
	}
}

func TestTTSSynthesizeWithCustomOutputPath(t *testing.T) {
	input := TtsSynthesizeInput{
		TextUTF8: "Testing with custom output path.",
		Config: &TTSConfig{
			Voice:      "en_speaker_1",
			Speed:      0.8,
			Seed:       42,
			SampleRate: 22050,
		},
		OutputPath: "./custom_test_output.wav",
	}

	output, err := tts.Synthesize(input)
	if err != nil {
		t.Fatalf("Synthesize with custom output path failed: %v", err)
	}

	t.Logf("Synthesize with custom output path completed successfully")
	t.Logf("Audio path: %s", output.Result.AudioPath)
	t.Logf("Duration: %.2f seconds", output.Result.DurationSeconds)

	// Clean up test file
	if output.Result.AudioPath != "" {
		_ = os.Remove(output.Result.AudioPath)
	}
}

func TestTTSSynthesizeLongText(t *testing.T) {
	longText := `This is a longer text for testing text-to-speech synthesis capabilities. 
	It contains multiple sentences and should test the TTS system's ability to handle 
	longer inputs and maintain natural speech patterns throughout the synthesis process.`

	input := TtsSynthesizeInput{
		TextUTF8: longText,
		Config: &TTSConfig{
			Voice:      "en_speaker_0",
			Speed:      1.2,
			Seed:       123,
			SampleRate: 24000,
		},
	}

	output, err := tts.Synthesize(input)
	if err != nil {
		t.Fatalf("Synthesize long text failed: %v", err)
	}

	t.Logf("Synthesize long text completed successfully")
	t.Logf("Audio path: %s", output.Result.AudioPath)
	t.Logf("Duration: %.2f seconds", output.Result.DurationSeconds)

	// Clean up test file
	if output.Result.AudioPath != "" {
		_ = os.Remove(output.Result.AudioPath)
	}
}

func TestTTSSynthesizeSpecialCharacters(t *testing.T) {
	textWithSpecialChars := "Hello! This text contains special characters: @#$%^&*() and numbers: 12345. It also has punctuation marks: ,.;:!?"

	input := TtsSynthesizeInput{
		TextUTF8: textWithSpecialChars,
		Config: &TTSConfig{
			Voice:      "en_speaker_0",
			Speed:      1.0,
			Seed:       999,
			SampleRate: 24000,
		},
	}

	output, err := tts.Synthesize(input)
	if err != nil {
		t.Fatalf("Synthesize with special characters failed: %v", err)
	}

	t.Logf("Synthesize with special characters completed successfully")
	t.Logf("Audio path: %s", output.Result.AudioPath)
	t.Logf("Duration: %.2f seconds", output.Result.DurationSeconds)

	// Clean up test file
	if output.Result.AudioPath != "" {
		_ = os.Remove(output.Result.AudioPath)
	}
} 