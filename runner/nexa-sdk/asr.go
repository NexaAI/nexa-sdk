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

// ASRConfig represents ASR processing configuration
type ASRConfig struct {
	Timestamps string
	BeamSize   int32
	Stream     bool
}

func (ac ASRConfig) toCPtr() *C.ml_ASRConfig {
	cPtr := (*C.ml_ASRConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_ASRConfig{}))))
	*cPtr = C.ml_ASRConfig{}

	if ac.Timestamps != "" {
		cPtr.timestamps = C.CString(ac.Timestamps)
	}
	cPtr.beam_size = C.int32_t(ac.BeamSize)
	cPtr.stream = C.bool(ac.Stream)

	return cPtr
}

func freeASRConfig(cPtr *C.ml_ASRConfig) {
	if cPtr != nil {
		if cPtr.timestamps != nil {
			C.free(unsafe.Pointer(cPtr.timestamps))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// ASRResult represents ASR transcription result
type ASRResult struct {
	Transcript       string
	ConfidenceScores []float32
	Timestamps       []float32
}

func newASRResultFromCPtr(c *C.ml_ASRResult) ASRResult {
	result := ASRResult{}

	if c == nil {
		return result
	}

	if c.transcript != nil {
		result.Transcript = C.GoString(c.transcript)
	}

	// Convert confidence scores array
	if c.confidence_scores != nil && c.confidence_count > 0 {
		scores := unsafe.Slice((*C.float)(unsafe.Pointer(c.confidence_scores)), int(c.confidence_count))
		result.ConfidenceScores = make([]float32, c.confidence_count)
		for i := range result.ConfidenceScores {
			result.ConfidenceScores[i] = float32(scores[i])
		}
	}

	// Convert timestamps array
	if c.timestamps != nil && c.timestamp_count > 0 {
		timestamps := unsafe.Slice((*C.float)(unsafe.Pointer(c.timestamps)), int(c.timestamp_count))
		result.Timestamps = make([]float32, c.timestamp_count)
		for i := range result.Timestamps {
			result.Timestamps[i] = float32(timestamps[i])
		}
	}

	return result
}

func freeASRResult(ptr *C.ml_ASRResult) {
	if ptr == nil {
		return
	}
	if ptr.transcript != nil {
		mlFree(unsafe.Pointer(ptr.transcript))
	}
	if ptr.confidence_scores != nil {
		mlFree(unsafe.Pointer(ptr.confidence_scores))
	}
	if ptr.timestamps != nil {
		mlFree(unsafe.Pointer(ptr.timestamps))
	}
}

// AsrCreateInput represents input parameters for ASR creation
type AsrCreateInput struct {
	ModelPath     string
	TokenizerPath string
	Language      string
	PluginID      string
	DeviceID      string
}

func (aci AsrCreateInput) toCPtr() *C.ml_AsrCreateInput {
	cPtr := (*C.ml_AsrCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrCreateInput{}))))
	*cPtr = C.ml_AsrCreateInput{}

	if aci.ModelPath != "" {
		cPtr.model_path = C.CString(aci.ModelPath)
	}
	if aci.TokenizerPath != "" {
		cPtr.tokenizer_path = C.CString(aci.TokenizerPath)
	}
	if aci.Language != "" {
		cPtr.language = C.CString(aci.Language)
	}
	if aci.PluginID != "" {
		cPtr.plugin_id = C.CString(aci.PluginID)
	}
	if aci.DeviceID != "" {
		cPtr.device_id = C.CString(aci.DeviceID)
	}

	return cPtr
}

func freeAsrCreateInput(cPtr *C.ml_AsrCreateInput) {
	if cPtr != nil {
		if cPtr.model_path != nil {
			C.free(unsafe.Pointer(cPtr.model_path))
		}
		if cPtr.tokenizer_path != nil {
			C.free(unsafe.Pointer(cPtr.tokenizer_path))
		}
		if cPtr.language != nil {
			C.free(unsafe.Pointer(cPtr.language))
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

// AsrTranscribeInput represents input parameters for ASR transcription
type AsrTranscribeInput struct {
	AudioPath string
	Language  string
	Config    *ASRConfig
}

func (ati AsrTranscribeInput) toCPtr() *C.ml_AsrTranscribeInput {
	cPtr := (*C.ml_AsrTranscribeInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrTranscribeInput{}))))
	*cPtr = C.ml_AsrTranscribeInput{}

	cPtr.audio_path = C.CString(ati.AudioPath)
	if ati.Language != "" {
		cPtr.language = C.CString(ati.Language)
	} else {
		cPtr.language = nil
	}
	if ati.Config != nil {
		cPtr.config = ati.Config.toCPtr()
	} else {
		cPtr.config = nil
	}

	return cPtr
}

func freeAsrTranscribeInput(cPtr *C.ml_AsrTranscribeInput) {
	if cPtr != nil {
		if cPtr.audio_path != nil {
			C.free(unsafe.Pointer(cPtr.audio_path))
		}
		if cPtr.language != nil {
			C.free(unsafe.Pointer(cPtr.language))
		}
		if cPtr.config != nil {
			freeASRConfig(cPtr.config)
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// AsrTranscribeOutput represents output from ASR transcription
type AsrTranscribeOutput struct {
	Result      ASRResult
	ProfileData ProfileData
}

func newAsrTranscribeOutputFromCPtr(c *C.ml_AsrTranscribeOutput) AsrTranscribeOutput {
	output := AsrTranscribeOutput{}

	if c == nil {
		return output
	}

	output.Result = newASRResultFromCPtr(&c.result)
	output.ProfileData = newProfileDataFromCPtr(c.profile_data)
	return output
}

func freeAsrTranscribeOutput(ptr *C.ml_AsrTranscribeOutput) {
	if ptr == nil {
		return
	}
	freeASRResult(&ptr.result)
}

// AsrListSupportedLanguagesInput represents input for listing supported languages
type AsrListSupportedLanguagesInput struct {
	Reserved interface{}
}

func (asli AsrListSupportedLanguagesInput) toCPtr() *C.ml_AsrListSupportedLanguagesInput {
	cPtr := (*C.ml_AsrListSupportedLanguagesInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrListSupportedLanguagesInput{}))))
	*cPtr = C.ml_AsrListSupportedLanguagesInput{}
	return cPtr
}

func freeAsrListSupportedLanguagesInput(cPtr *C.ml_AsrListSupportedLanguagesInput) {
	if cPtr != nil {
		C.free(unsafe.Pointer(cPtr))
	}
}

// AsrListSupportedLanguagesOutput represents output for listing supported languages
type AsrListSupportedLanguagesOutput struct {
	LanguageCodes []string
}

func newAsrListSupportedLanguagesOutputFromCPtr(c *C.ml_AsrListSupportedLanguagesOutput) AsrListSupportedLanguagesOutput {
	output := AsrListSupportedLanguagesOutput{}

	if c == nil {
		return output
	}

	output.LanguageCodes = cCharArrayToSlice(c.language_codes, c.language_count)

	return output
}

func freeAsrListSupportedLanguagesOutput(ptr *C.ml_AsrListSupportedLanguagesOutput) {
	if ptr == nil {
		return
	}
	if ptr.language_codes != nil {
		mlFreeCCharArray(ptr.language_codes, ptr.language_count)
	}
}

// ASR represents an ASR instance
type ASR struct {
	ptr *C.ml_ASR
}

// NewASR creates a new ASR instance
func NewASR(input AsrCreateInput) (*ASR, error) {
	slog.Debug("NewASR called", "input", input)

	cInput := input.toCPtr()
	defer freeAsrCreateInput(cInput)

	var cHandle *C.ml_ASR
	res := C.ml_asr_create(cInput, &cHandle)
	if res < 0 {
		return nil, SDKError(res)
	}

	return &ASR{ptr: cHandle}, nil
}

// Destroy destroys the ASR instance and frees associated resources
func (a *ASR) Destroy() error {
	slog.Debug("Destroy called", "ptr", a.ptr)

	if a.ptr == nil {
		return nil
	}

	res := C.ml_asr_destroy(a.ptr)
	if res < 0 {
		return SDKError(res)
	}
	a.ptr = nil
	return nil
}

// Transcribe transcribes audio file to text with specified language
func (a *ASR) Transcribe(input AsrTranscribeInput) (AsrTranscribeOutput, error) {
	slog.Debug("Transcribe called", "input", input)

	cInput := input.toCPtr()
	defer freeAsrTranscribeInput(cInput)

	var cOutput C.ml_AsrTranscribeOutput
	defer freeAsrTranscribeOutput(&cOutput)

	res := C.ml_asr_transcribe(a.ptr, cInput, &cOutput)
	if res < 0 {
		return AsrTranscribeOutput{}, SDKError(res)
	}

	output := newAsrTranscribeOutputFromCPtr(&cOutput)
	return output, nil
}

// ListSupportedLanguages gets list of supported languages for ASR model
func (a *ASR) ListSupportedLanguages() (AsrListSupportedLanguagesOutput, error) {
	slog.Debug("ListSupportedLanguages called")

	input := AsrListSupportedLanguagesInput{}
	cInput := input.toCPtr()
	defer freeAsrListSupportedLanguagesInput(cInput)

	var cOutput C.ml_AsrListSupportedLanguagesOutput
	defer freeAsrListSupportedLanguagesOutput(&cOutput)

	res := C.ml_asr_list_supported_languages(a.ptr, cInput, &cOutput)
	if res < 0 {
		return AsrListSupportedLanguagesOutput{}, SDKError(res)
	}

	output := newAsrListSupportedLanguagesOutputFromCPtr(&cOutput)
	return output, nil
}
