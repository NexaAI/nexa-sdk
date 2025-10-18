// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"
import (
	"log/slog"
	"unsafe"
)

// TTSConfig represents TTS synthesis configuration
type TTSConfig struct {
	Voice      string
	Speed      float32
	Seed       int32
	SampleRate int32
}

func (tc TTSConfig) toCPtr() *C.ml_TTSConfig {
	cPtr := (*C.ml_TTSConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_TTSConfig{}))))
	*cPtr = C.ml_TTSConfig{}

	if tc.Voice != "" {
		cPtr.voice = C.CString(tc.Voice)
	}
	cPtr.speed = C.float(tc.Speed)
	cPtr.seed = C.int32_t(tc.Seed)
	cPtr.sample_rate = C.int32_t(tc.SampleRate)

	return cPtr
}

func freeTTSConfig(cPtr *C.ml_TTSConfig) {
	if cPtr != nil {
		if cPtr.voice != nil {
			C.free(unsafe.Pointer(cPtr.voice))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// // TTSSamplerConfig represents TTS sampling parameters
// type TTSSamplerConfig struct {
// 	Temperature  float32
// 	NoiseScale   float32
// 	LengthScale  float32
// }

// func (tsc TTSSamplerConfig) toCPtr() *C.ml_TTSSamplerConfig {
// 	cPtr := (*C.ml_TTSSamplerConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_TTSSamplerConfig{}))))
// 	*cPtr = C.ml_TTSSamplerConfig{}

// 	cPtr.temperature = C.float(tsc.Temperature)
// 	cPtr.noise_scale = C.float(tsc.NoiseScale)
// 	cPtr.length_scale = C.float(tsc.LengthScale)

// 	return cPtr
// }

// func freeTTSSamplerConfig(cPtr *C.ml_TTSSamplerConfig) {
// 	if cPtr != nil {
// 		C.free(unsafe.Pointer(cPtr))
// 	}
// }

// TTSResult represents TTS synthesis result
type TTSResult struct {
	AudioPath       string
	DurationSeconds float32
	SampleRate      int32
	Channels        int32
	NumSamples      int32
}

func newTTSResultFromCPtr(c *C.ml_TTSResult) TTSResult {
	result := TTSResult{}

	if c == nil {
		return result
	}

	if c.audio_path != nil {
		result.AudioPath = C.GoString(c.audio_path)
	}
	result.DurationSeconds = float32(c.duration_seconds)
	result.SampleRate = int32(c.sample_rate)
	result.Channels = int32(c.channels)
	result.NumSamples = int32(c.num_samples)

	return result
}

func freeTTSResult(ptr *C.ml_TTSResult) {
	if ptr == nil {
		return
	}
	if ptr.audio_path != nil {
		mlFree(unsafe.Pointer(ptr.audio_path))
	}
}

// TtsCreateInput represents input parameters for TTS creation
type TtsCreateInput struct {
	ModelName   string
	ModelPath   string
	VocoderPath string
	PluginID    string
	DeviceID    string
}

func (tci TtsCreateInput) toCPtr() *C.ml_TtsCreateInput {
	cPtr := (*C.ml_TtsCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_TtsCreateInput{}))))
	*cPtr = C.ml_TtsCreateInput{}

	if tci.ModelName != "" {
		cPtr.model_name = C.CString(tci.ModelName)
	}
	if tci.ModelPath != "" {
		cPtr.model_path = C.CString(tci.ModelPath)
	}
	if tci.VocoderPath != "" {
		cPtr.vocoder_path = C.CString(tci.VocoderPath)
	}
	if tci.PluginID != "" {
		cPtr.plugin_id = C.CString(tci.PluginID)
	}
	if tci.DeviceID != "" {
		cPtr.device_id = C.CString(tci.DeviceID)
	}

	return cPtr
}

func freeTtsCreateInput(cPtr *C.ml_TtsCreateInput) {
	if cPtr != nil {
		if cPtr.model_path != nil {
			C.free(unsafe.Pointer(cPtr.model_path))
		}
		if cPtr.vocoder_path != nil {
			C.free(unsafe.Pointer(cPtr.vocoder_path))
		}
		if cPtr.plugin_id != nil {
			C.free(unsafe.Pointer(cPtr.plugin_id))
		}
		if cPtr.device_id != nil {
			C.free(unsafe.Pointer(cPtr.device_id))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// TtsSynthesizeInput represents input parameters for TTS synthesis
type TtsSynthesizeInput struct {
	TextUTF8   string
	Config     *TTSConfig
	OutputPath string
}

func (tsi TtsSynthesizeInput) toCPtr() *C.ml_TtsSynthesizeInput {
	cPtr := (*C.ml_TtsSynthesizeInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_TtsSynthesizeInput{}))))
	*cPtr = C.ml_TtsSynthesizeInput{}

	cPtr.text_utf8 = C.CString(tsi.TextUTF8)
	if tsi.Config != nil {
		cPtr.config = tsi.Config.toCPtr()
	} else {
		cPtr.config = nil
	}
	if tsi.OutputPath != "" {
		cPtr.output_path = C.CString(tsi.OutputPath)
	} else {
		cPtr.output_path = nil
	}

	return cPtr
}

func freeTtsSynthesizeInput(cPtr *C.ml_TtsSynthesizeInput) {
	if cPtr != nil {
		if cPtr.text_utf8 != nil {
			C.free(unsafe.Pointer(cPtr.text_utf8))
		}
		if cPtr.config != nil {
			freeTTSConfig(cPtr.config)
		}
		if cPtr.output_path != nil {
			C.free(unsafe.Pointer(cPtr.output_path))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// TtsSynthesizeOutput represents output from TTS synthesis
type TtsSynthesizeOutput struct {
	Result      TTSResult
	ProfileData ProfileData
}

func newTtsSynthesizeOutputFromCPtr(c *C.ml_TtsSynthesizeOutput) TtsSynthesizeOutput {
	output := TtsSynthesizeOutput{}

	if c == nil {
		return output
	}

	output.Result = newTTSResultFromCPtr(&c.result)
	output.ProfileData = newProfileDataFromCPtr(c.profile_data)
	return output
}

func freeTtsSynthesizeOutput(ptr *C.ml_TtsSynthesizeOutput) {
	if ptr == nil {
		return
	}
	freeTTSResult(&ptr.result)
}

// TtsListAvailableVoicesInput represents input for listing available voices
type TtsListAvailableVoicesInput struct {
	Reserved any
}

func (tlavi TtsListAvailableVoicesInput) toCPtr() *C.ml_TtsListAvailableVoicesInput {
	cPtr := (*C.ml_TtsListAvailableVoicesInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_TtsListAvailableVoicesInput{}))))
	*cPtr = C.ml_TtsListAvailableVoicesInput{}
	return cPtr
}

func freeTtsListAvailableVoicesInput(cPtr *C.ml_TtsListAvailableVoicesInput) {
	if cPtr != nil {
		C.free(unsafe.Pointer(cPtr))
	}
}

// TtsListAvailableVoicesOutput represents output for listing available voices
type TtsListAvailableVoicesOutput struct {
	VoiceIDs []string
}

func newTtsListAvailableVoicesOutputFromCPtr(c *C.ml_TtsListAvailableVoicesOutput) TtsListAvailableVoicesOutput {
	output := TtsListAvailableVoicesOutput{}

	if c == nil {
		return output
	}

	output.VoiceIDs = cCharArrayToSlice(c.voice_ids, c.voice_count)

	return output
}

func freeTtsListAvailableVoicesOutput(ptr *C.ml_TtsListAvailableVoicesOutput) {
	if ptr == nil {
		return
	}
	if ptr.voice_ids != nil {
		mlFreeCCharArray(ptr.voice_ids, ptr.voice_count)
	}
}

// TTS represents a TTS instance
type TTS struct {
	ptr *C.ml_TTS
}

// NewTTS creates a new TTS instance
func NewTTS(input TtsCreateInput) (*TTS, error) {
	slog.Debug("NewTTS called", "input", input)

	cInput := input.toCPtr()
	defer freeTtsCreateInput(cInput)

	var cHandle *C.ml_TTS
	res := C.ml_tts_create(cInput, &cHandle)
	if res < 0 {
		return nil, SDKError(res)
	}

	return &TTS{ptr: cHandle}, nil
}

// Destroy destroys the TTS instance and frees associated resources
func (t *TTS) Destroy() error {
	slog.Debug("Destroy called", "ptr", t.ptr)

	if t.ptr == nil {
		return nil
	}

	res := C.ml_tts_destroy(t.ptr)
	if res < 0 {
		return SDKError(res)
	}
	t.ptr = nil
	return nil
}

// Synthesize synthesizes speech from text and saves to filesystem
func (t *TTS) Synthesize(input TtsSynthesizeInput) (TtsSynthesizeOutput, error) {
	slog.Debug("Synthesize called", "input", input)

	cInput := input.toCPtr()
	defer freeTtsSynthesizeInput(cInput)

	var cOutput C.ml_TtsSynthesizeOutput
	defer freeTtsSynthesizeOutput(&cOutput)

	res := C.ml_tts_synthesize(t.ptr, cInput, &cOutput)
	if res < 0 {
		return TtsSynthesizeOutput{}, SDKError(res)
	}

	output := newTtsSynthesizeOutputFromCPtr(&cOutput)
	return output, nil
}

// ListAvailableVoices gets list of available voice identifiers
func (t *TTS) ListAvailableVoices() (TtsListAvailableVoicesOutput, error) {
	slog.Debug("ListAvailableVoices called")

	input := TtsListAvailableVoicesInput{}
	cInput := input.toCPtr()
	defer freeTtsListAvailableVoicesInput(cInput)

	var cOutput C.ml_TtsListAvailableVoicesOutput
	defer freeTtsListAvailableVoicesOutput(&cOutput)

	res := C.ml_tts_list_available_voices(t.ptr, cInput, &cOutput)
	if res < 0 {
		return TtsListAvailableVoicesOutput{}, SDKError(res)
	}

	output := newTtsListAvailableVoicesOutputFromCPtr(&cOutput)
	return output, nil
}
