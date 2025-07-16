package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"

import (
	"unsafe"

	"golang.org/x/exp/slog"
)

// TTSConfig represents TTS synthesis configuration
type TTSConfig struct {
	Voice      string  // Voice identifier
	Speed      float32 // Speech speed (1.0 = normal)
	Seed       int32   // Random seed (-1 for random)
	SampleRate int32   // Output sample rate in Hz
}

// TTSSamplerConfig represents TTS sampling parameters
type TTSSamplerConfig struct {
	Temperature float32 // Sampling temperature
	NoiseScale  float32 // Noise scale for voice variation
	LengthScale float32 // Length scale for speech duration
}

// TTSResult represents TTS synthesis result
type TTSResult struct {
	Audio           []float32 // Audio samples: num_samples × channels
	DurationSeconds float32   // Audio duration in seconds
	SampleRate      int32     // Audio sample rate in Hz
	Channels        int32     // Number of audio channels (default: 1)
	NumSamples      int32     // Number of audio samples
}

// TTS wraps the C library TTS structure and provides Go interface
type TTS struct {
	ptr *C.ml_TTS // Pointer to the underlying C TTS structure
}

// NewTTS creates a new TTS instance with the specified model and vocoder
func NewTTS(model string, vocoder *string, devices *string) (*TTS, error) {
	slog.Debug("NewTTS called", "model", model, "vocoder", vocoder, "devices", devices)
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	var cVocoder *C.char
	if vocoder != nil {
		cVocoder = C.CString(*vocoder)
		defer C.free(unsafe.Pointer(cVocoder))
	}

	ptr := C.ml_tts_create(cModel, cVocoder, nil)
	if ptr == nil {
		slog.Debug("NewTTS failed", "error", SDKErrorUnknown)
		return nil, SDKErrorUnknown
	}
	slog.Debug("NewTTS success", "ptr", ptr)
	return &TTS{ptr: ptr}, nil
}

// Destroy frees the memory allocated for the TTS instance
func (p *TTS) Destroy() {
	slog.Debug("Destroy called", "ptr", p.ptr)
	C.ml_tts_destroy(p.ptr)
	p.ptr = nil
}

// Reset clears the TTS's internal state
func (p *TTS) Reset() {
	slog.Debug("Reset called", "ptr", p.ptr)
	// Reset TTS state if needed
}

// LoadModel loads TTS model from path with optional extra configuration data
func (p *TTS) LoadModel(modelPath string, extraData unsafe.Pointer) error {
	slog.Debug("LoadModel called", "modelPath", modelPath)
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_tts_load_model(p.ptr, cPath, extraData)
	if !res {
		slog.Debug("LoadModel failed", "error", SDKErrorUnknown)
		return SDKErrorUnknown
	}
	slog.Debug("LoadModel success")
	return nil
}

// Close cleanup TTS resources
func (p *TTS) Close() {
	slog.Debug("Close called", "ptr", p.ptr)
	C.ml_tts_close(p.ptr)
}

// SetSampler configures TTS sampling parameters
func (p *TTS) SetSampler(config *TTSSamplerConfig) {
	slog.Debug("SetSampler called", "config", config)
	if config == nil {
		return
	}

	cConfig := C.ml_TTSSamplerConfig{
		temperature:  C.float(config.Temperature),
		noise_scale:  C.float(config.NoiseScale),
		length_scale: C.float(config.LengthScale),
	}

	C.ml_tts_set_sampler(p.ptr, &cConfig)
}

// ResetSampler resets TTS sampling parameters to default values
func (p *TTS) ResetSampler() {
	slog.Debug("ResetSampler called", "ptr", p.ptr)
	C.ml_tts_reset_sampler(p.ptr)
}

// Synthesize converts text to speech audio
func (p *TTS) Synthesize(text string, config *TTSConfig) (*TTSResult, error) {
	slog.Debug("Synthesize called", "text", text, "config", config)
	cText := C.CString(text)
	defer C.free(unsafe.Pointer(cText))

	var cConfig C.ml_TTSConfig
	if config != nil {
		cVoice := C.CString(config.Voice)
		defer C.free(unsafe.Pointer(cVoice))

		cConfig = C.ml_TTSConfig{
			voice:       cVoice,
			speed:       C.float(config.Speed),
			seed:        C.int32_t(config.Seed),
			sample_rate: C.int32_t(config.SampleRate),
		}
	}

	res := C.ml_tts_synthesize(p.ptr, cText, &cConfig)
	if res.audio == nil {
		slog.Debug("Synthesize failed", "error", SDKErrorUnknown)
		return nil, SDKErrorUnknown
	}

	// Convert C result to Go result
	audioLen := int(res.num_samples)
	audio := make([]float32, audioLen)
	if audioLen > 0 {
		copy(audio, (*[1 << 30]float32)(unsafe.Pointer(res.audio))[:audioLen])
	}

	result := &TTSResult{
		Audio:           audio,
		DurationSeconds: float32(res.duration_seconds),
		SampleRate:      int32(res.sample_rate),
		Channels:        int32(res.channels),
		NumSamples:      int32(res.num_samples),
	}

	// Free C memory
	C.ml_tts_free_result(&res)

	slog.Debug("Synthesize success", "result", result)
	return result, nil
}

// SynthesizeBatch processes multiple texts in batch mode
func (p *TTS) SynthesizeBatch(texts []string, config *TTSConfig) ([]*TTSResult, error) {
	slog.Debug("SynthesizeBatch called", "texts", texts, "config", config)
	cTexts := make([]*C.char, len(texts))
	for i, text := range texts {
		cText := &cTexts[i]
		*cText = C.CString(text)
		defer C.free(unsafe.Pointer(*cText))
	}

	var cConfig C.ml_TTSConfig
	if config != nil {
		cVoice := C.CString(config.Voice)
		defer C.free(unsafe.Pointer(cVoice))

		cConfig = C.ml_TTSConfig{
			voice:       cVoice,
			speed:       C.float(config.Speed),
			seed:        C.int32_t(config.Seed),
			sample_rate: C.int32_t(config.SampleRate),
		}
	}

	cRes := C.ml_tts_synthesize_batch(p.ptr, &cTexts[0], C.int32_t(len(texts)), &cConfig)
	if cRes == nil {
		slog.Debug("SynthesizeBatch failed", "error", SDKErrorUnknown)
		return nil, SDKErrorUnknown
	}

	// Convert C results to Go results
	results := make([]*TTSResult, len(texts))
	for i := 0; i < len(texts); i++ {
		res := (*C.ml_TTSResult)(unsafe.Pointer(uintptr(unsafe.Pointer(cRes)) + uintptr(i)*unsafe.Sizeof(C.ml_TTSResult{})))

		audioLen := int(res.num_samples)
		audio := make([]float32, audioLen)
		if audioLen > 0 {
			copy(audio, (*[1 << 30]float32)(unsafe.Pointer(res.audio))[:audioLen])
		}

		results[i] = &TTSResult{
			Audio:           audio,
			DurationSeconds: float32(res.duration_seconds),
			SampleRate:      int32(res.sample_rate),
			Channels:        int32(res.channels),
			NumSamples:      int32(res.num_samples),
		}
	}

	// Free C memory
	C.free(unsafe.Pointer(cRes))

	slog.Debug("SynthesizeBatch success", "results", results)
	return results, nil
}

// SaveCache saves TTS cache state to file
func (p *TTS) SaveCache(path string) error {
	slog.Debug("SaveCache called", "path", path)
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	C.ml_tts_save_cache(p.ptr, cPath)
	slog.Debug("SaveCache success")
	return nil
}

// LoadCache loads TTS cache state from file
func (p *TTS) LoadCache(path string) error {
	slog.Debug("LoadCache called", "path", path)
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	C.ml_tts_load_cache(p.ptr, cPath)
	slog.Debug("LoadCache success")
	return nil
}

// ListAvailableVoices returns list of available voices
func (p *TTS) ListAvailableVoices() ([]string, error) {
	slog.Debug("ListAvailableVoices called")
	var count C.int32_t
	cVoices := C.ml_tts_list_available_voices(p.ptr, &count)
	if cVoices == nil {
		slog.Debug("ListAvailableVoices failed", "error", SDKErrorUnknown)
		return nil, SDKErrorUnknown
	}

	voices := make([]string, int(count))
	for i := 0; i < int(count); i++ {
		voice := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cVoices)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))
		voices[i] = C.GoString(voice)
	}

	slog.Debug("ListAvailableVoices success", "voices", voices)
	return voices, nil
}

// GetProfilingData retrieves performance metrics from the TTS instance
func (p *TTS) GetProfilingData() (*ProfilingData, error) {
	slog.Debug("GetProfilingData called")
	// Note: TTS doesn't have profiling data in the C API
	// Return empty profiling data for consistency
	return &ProfilingData{}, nil
}

// PrintResult prints TTS result information to standard output for debugging
func (result *TTSResult) PrintResult() {
	slog.Debug("PrintResult called", "result", result)
	if result == nil {
		return
	}

	cResult := C.ml_TTSResult{
		audio:            (*C.float)(unsafe.Pointer(&result.Audio[0])),
		duration_seconds: C.float(result.DurationSeconds),
		sample_rate:      C.int32_t(result.SampleRate),
		channels:         C.int32_t(result.Channels),
		num_samples:      C.int32_t(result.NumSamples),
	}

	C.ml_tts_print_result(&cResult)
}
