package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"unsafe"
)

// WAVHeader represents the WAV file header structure
type WAVHeader struct {
	Riff          [4]byte // "RIFF"
	FileSize      uint32  // File size - 8
	Wave          [4]byte // "WAVE"
	Fmt           [4]byte // "fmt "
	FmtSize       uint32  // Format chunk size
	AudioFormat   uint16  // Audio format (1 = PCM)
	NumChannels   uint16  // Number of channels
	SampleRate    uint32  // Sample rate
	ByteRate      uint32  // Byte rate
	BlockAlign    uint16  // Block align
	BitsPerSample uint16  // Bits per sample
	Data          [4]byte // "data"
	DataSize      uint32  // Data chunk size
}

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

// SaveWAV saves the TTS result as a WAV file
func (result *TTSResult) SaveWAV(filename string) error {
	if result == nil || len(result.Audio) == 0 {
		return fmt.Errorf("no audio data to save")
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("could not create audio file: %v", err)
	}
	defer file.Close()

	// Prepare WAV header
	header := WAVHeader{}
	copy(header.Riff[:], "RIFF")
	copy(header.Wave[:], "WAVE")
	copy(header.Fmt[:], "fmt ")
	copy(header.Data[:], "data")

	header.FmtSize = 16
	header.AudioFormat = 1 // PCM
	header.NumChannels = uint16(result.Channels)
	header.SampleRate = uint32(result.SampleRate)
	header.BitsPerSample = 16
	header.ByteRate = uint32(result.SampleRate) * uint32(result.Channels) * uint32(header.BitsPerSample) / 8
	header.BlockAlign = uint16(result.Channels) * header.BitsPerSample / 8
	header.DataSize = uint32(len(result.Audio)) * uint32(result.Channels) * uint32(header.BitsPerSample) / 8
	header.FileSize = uint32(unsafe.Sizeof(header)) - 8 + header.DataSize

	// Write header
	err = binary.Write(file, binary.LittleEndian, header)
	if err != nil {
		return fmt.Errorf("failed to write WAV header: %v", err)
	}

	// Convert float samples to 16-bit PCM and write
	pcmSamples := make([]int16, len(result.Audio)*int(result.Channels))
	for i, sample := range result.Audio {
		// Clamp audio values to [-1.0, 1.0] and convert to 16-bit PCM
		clampedSample := math.Max(-1.0, math.Min(1.0, float64(sample)))
		pcmSamples[i] = int16(clampedSample * 32767.0)
	}

	err = binary.Write(file, binary.LittleEndian, pcmSamples)
	if err != nil {
		return fmt.Errorf("failed to write audio data: %v", err)
	}

	// fmt.Printf("Audio saved to: %s\n", filename)
	// fmt.Printf("  Duration: %.2f seconds\n", float64(len(result.Audio))/float64(result.SampleRate))
	// fmt.Printf("  Sample rate: %d Hz\n", result.SampleRate)
	// fmt.Printf("  Channels: %d\n", result.Channels)
	// fmt.Printf("  Samples: %d\n", len(result.Audio))
	return nil
}

// TTS wraps the C library TTS structure and provides Go interface
type TTS struct {
	ptr *C.ml_TTS // Pointer to the underlying C TTS structure
}

// NewTTS creates a new TTS instance with the specified model and vocoder
func NewTTS(model string, vocoder *string, devices *string) (*TTS, error) {
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	var cVocoder *C.char
	if vocoder != nil {
		cVocoder = C.CString(*vocoder)
		defer C.free(unsafe.Pointer(cVocoder))
	}

	ptr := C.ml_tts_create(cModel, cVocoder, nil)
	if ptr == nil {
		return nil, ErrCreateFailed
	}

	return &TTS{ptr: ptr}, nil
}

// Destroy frees the memory allocated for the TTS instance
func (p *TTS) Destroy() {
	C.ml_tts_destroy(p.ptr)
	p.ptr = nil
}

// Reset clears the TTS's internal state
func (p *TTS) Reset() {
	// Reset TTS state if needed
}

// LoadModel loads TTS model from path with optional extra configuration data
func (p *TTS) LoadModel(modelPath string, extraData unsafe.Pointer) error {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_tts_load_model(p.ptr, cPath, extraData)
	if !res {
		return ErrCommon
	}
	return nil
}

// Close cleanup TTS resources
func (p *TTS) Close() {
	C.ml_tts_close(p.ptr)
}

// SetSampler configures TTS sampling parameters
func (p *TTS) SetSampler(config *TTSSamplerConfig) {
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
	C.ml_tts_reset_sampler(p.ptr)
}

// Synthesize converts text to speech audio
func (p *TTS) Synthesize(text string, config *TTSConfig) (*TTSResult, error) {
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
		return nil, ErrCommon
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

	return result, nil
}

// SynthesizeBatch processes multiple texts in batch mode
func (p *TTS) SynthesizeBatch(texts []string, config *TTSConfig) ([]*TTSResult, error) {
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
		return nil, ErrCommon
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

	return results, nil
}

// SaveCache saves TTS cache state to file
func (p *TTS) SaveCache(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	C.ml_tts_save_cache(p.ptr, cPath)
	return nil
}

// LoadCache loads TTS cache state from file
func (p *TTS) LoadCache(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	C.ml_tts_load_cache(p.ptr, cPath)
	return nil
}

// ListAvailableVoices returns list of available voices
func (p *TTS) ListAvailableVoices() ([]string, error) {
	var count C.int32_t
	cVoices := C.ml_tts_list_available_voices(p.ptr, &count)
	if cVoices == nil {
		return nil, ErrCommon
	}

	voices := make([]string, int(count))
	for i := 0; i < int(count); i++ {
		voice := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cVoices)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))
		voices[i] = C.GoString(voice)
	}

	return voices, nil
}

// GetProfilingData retrieves performance metrics from the TTS instance
func (p *TTS) GetProfilingData() (*ProfilingData, error) {
	// Note: TTS doesn't have profiling data in the C API
	// Return empty profiling data for consistency
	return &ProfilingData{}, nil
}

// PrintResult prints TTS result information to standard output for debugging
func (result *TTSResult) PrintResult() {
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
