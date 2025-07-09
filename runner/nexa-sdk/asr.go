package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"

import (
	"unsafe"
)

// ASRConfig represents ASR transcription configuration
type ASRConfig struct {
	Timestamps string // Timestamp mode: "none", "segment", "word"
	BeamSize   int32  // Beam search size
	Stream     bool   // Enable streaming mode
}

// ASRResult represents ASR transcription result
type ASRResult struct {
	Transcript       string    // Transcribed text (UTF-8)
	ConfidenceScores []float32 // Confidence scores for each unit
	Timestamps       []float32 // Timestamp pairs: [start, end] for each unit
}

// ASR wraps the C library ASR structure and provides Go interface
type ASR struct {
	ptr *C.ml_ASR // Pointer to the underlying C ASR structure
}

// NewASR creates a new ASR instance with the specified model and configuration
func NewASR(model string, tokenizer *string, language *string, devices *string) (*ASR, error) {
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	var cTokenizer *C.char
	if tokenizer != nil {
		cTokenizer = C.CString(*tokenizer)
		defer C.free(unsafe.Pointer(cTokenizer))
	}

	var cLanguage *C.char
	if language != nil {
		cLanguage = C.CString(*language)
		defer C.free(unsafe.Pointer(cLanguage))
	}

	ptr := C.ml_asr_create(cModel, cTokenizer, cLanguage, nil)
	if ptr == nil {
		return nil, ErrCreateFailed
	}

	return &ASR{ptr: ptr}, nil
}

// Destroy frees the memory allocated for the ASR instance
func (p *ASR) Destroy() {
	C.ml_asr_destroy(p.ptr)
	p.ptr = nil
}

// LoadModel loads ASR model from path with optional extra configuration data
func (p *ASR) LoadModel(modelPath string, extraData unsafe.Pointer) error {
	cPath := C.CString(modelPath)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_asr_load_model(p.ptr, cPath, extraData)
	if !res {
		return ErrCommon
	}
	return nil
}

// Close cleanup ASR resources
func (p *ASR) Close() {
	C.ml_asr_close(p.ptr)
}

// Transcribe converts audio to text using ASR
func (p *ASR) Transcribe(audio []float32, sampleRate int32, config *ASRConfig) (*ASRResult, error) {
	if len(audio) == 0 {
		return nil, ErrCommon
	}

	var cConfig C.ml_ASRConfig
	if config != nil {
		cTimestamps := C.CString(config.Timestamps)
		defer C.free(unsafe.Pointer(cTimestamps))

		cConfig = C.ml_ASRConfig{
			timestamps: cTimestamps,
			beam_size:  C.int32_t(config.BeamSize),
			stream:     C.bool(config.Stream),
		}
	}

	res := C.ml_asr_transcribe(
		p.ptr,
		(*C.float)(unsafe.Pointer(&audio[0])),
		C.int32_t(len(audio)),
		C.int32_t(sampleRate),
		&cConfig,
	)

	if res.transcript == nil {
		return nil, ErrCommon
	}

	// Convert C result to Go result
	result := &ASRResult{
		Transcript: C.GoString(res.transcript),
	}

	// Copy confidence scores
	if res.confidence_scores != nil && res.confidence_count > 0 {
		confidenceCount := int(res.confidence_count)
		result.ConfidenceScores = make([]float32, confidenceCount)
		copy(result.ConfidenceScores, (*[1 << 30]float32)(unsafe.Pointer(res.confidence_scores))[:confidenceCount])
	}

	// Copy timestamps
	if res.timestamps != nil && res.timestamp_count > 0 {
		timestampCount := int(res.timestamp_count)
		result.Timestamps = make([]float32, timestampCount)
		copy(result.Timestamps, (*[1 << 30]float32)(unsafe.Pointer(res.timestamps))[:timestampCount])
	}

	// Free C memory
	C.ml_asr_free_result(&res)

	return result, nil
}

// TranscribeBatch processes multiple audio samples in batch mode
func (p *ASR) TranscribeBatch(audios [][]float32, sampleRate int32, config *ASRConfig) ([]*ASRResult, error) {
	if len(audios) == 0 {
		return nil, ErrCommon
	}

	// Prepare C arrays
	cAudios := make([]*C.float, len(audios))
	numSamplesArray := make([]C.int32_t, len(audios))

	for i, audio := range audios {
		if len(audio) == 0 {
			return nil, ErrCommon
		}
		cAudios[i] = (*C.float)(unsafe.Pointer(&audio[0]))
		numSamplesArray[i] = C.int32_t(len(audio))
	}

	var cConfig C.ml_ASRConfig
	if config != nil {
		cTimestamps := C.CString(config.Timestamps)
		defer C.free(unsafe.Pointer(cTimestamps))

		cConfig = C.ml_ASRConfig{
			timestamps: cTimestamps,
			beam_size:  C.int32_t(config.BeamSize),
			stream:     C.bool(config.Stream),
		}
	}

	cRes := C.ml_asr_transcribe_batch(
		p.ptr,
		&cAudios[0],
		&numSamplesArray[0],
		C.int32_t(len(audios)),
		C.int32_t(sampleRate),
		&cConfig,
	)

	if cRes == nil {
		return nil, ErrCommon
	}

	// Convert C results to Go results
	results := make([]*ASRResult, len(audios))
	for i := 0; i < len(audios); i++ {
		res := (*C.ml_ASRResult)(unsafe.Pointer(uintptr(unsafe.Pointer(cRes)) + uintptr(i)*unsafe.Sizeof(C.ml_ASRResult{})))

		if res.transcript == nil {
			results[i] = &ASRResult{}
			continue
		}

		result := &ASRResult{
			Transcript: C.GoString(res.transcript),
		}

		// Copy confidence scores
		if res.confidence_scores != nil && res.confidence_count > 0 {
			confidenceCount := int(res.confidence_count)
			result.ConfidenceScores = make([]float32, confidenceCount)
			copy(result.ConfidenceScores, (*[1 << 30]float32)(unsafe.Pointer(res.confidence_scores))[:confidenceCount])
		}

		// Copy timestamps
		if res.timestamps != nil && res.timestamp_count > 0 {
			timestampCount := int(res.timestamp_count)
			result.Timestamps = make([]float32, timestampCount)
			copy(result.Timestamps, (*[1 << 30]float32)(unsafe.Pointer(res.timestamps))[:timestampCount])
		}

		results[i] = result
	}

	// Free C memory
	C.ml_asr_free_result(cRes)

	return results, nil
}

// TranscribeStep processes audio chunk for streaming transcription
func (p *ASR) TranscribeStep(audioChunk []float32, step int32, config *ASRConfig) (*ASRResult, error) {
	if len(audioChunk) == 0 {
		return nil, ErrCommon
	}

	var cConfig C.ml_ASRConfig
	if config != nil {
		cTimestamps := C.CString(config.Timestamps)
		defer C.free(unsafe.Pointer(cTimestamps))

		cConfig = C.ml_ASRConfig{
			timestamps: cTimestamps,
			beam_size:  C.int32_t(config.BeamSize),
			stream:     C.bool(config.Stream),
		}
	}

	res := C.ml_asr_transcribe_step(
		p.ptr,
		(*C.float)(unsafe.Pointer(&audioChunk[0])),
		C.int32_t(len(audioChunk)),
		C.int32_t(step),
		&cConfig,
	)

	if res.transcript == nil {
		return nil, ErrCommon
	}

	// Convert C result to Go result
	result := &ASRResult{
		Transcript: C.GoString(res.transcript),
	}

	// Copy confidence scores
	if res.confidence_scores != nil && res.confidence_count > 0 {
		confidenceCount := int(res.confidence_count)
		result.ConfidenceScores = make([]float32, confidenceCount)
		copy(result.ConfidenceScores, (*[1 << 30]float32)(unsafe.Pointer(res.confidence_scores))[:confidenceCount])
	}

	// Copy timestamps
	if res.timestamps != nil && res.timestamp_count > 0 {
		timestampCount := int(res.timestamp_count)
		result.Timestamps = make([]float32, timestampCount)
		copy(result.Timestamps, (*[1 << 30]float32)(unsafe.Pointer(res.timestamps))[:timestampCount])
	}

	// Free C memory
	C.ml_asr_free_result(&res)

	return result, nil
}

// PrintResult prints ASR result to stdout
func (p *ASR) PrintResult(result *ASRResult) {
	if result == nil {
		return
	}

	cResult := C.ml_ASRResult{
		transcript:        C.CString(result.Transcript),
		confidence_scores: (*C.float)(unsafe.Pointer(&result.ConfidenceScores[0])),
		confidence_count:  C.int32_t(len(result.ConfidenceScores)),
		timestamps:        (*C.float)(unsafe.Pointer(&result.Timestamps[0])),
		timestamp_count:   C.int32_t(len(result.Timestamps)),
	}
	defer C.free(unsafe.Pointer(cResult.transcript))

	C.ml_asr_print_result(&cResult)
}

// ListSupportedLanguages returns list of supported languages
func (p *ASR) ListSupportedLanguages() ([]string, error) {
	var count C.int32_t
	cLanguages := C.ml_asr_list_supported_languages(p.ptr, &count)
	if cLanguages == nil {
		return nil, ErrCommon
	}

	languages := make([]string, count)
	for i := 0; i < int(count); i++ {
		langPtr := (*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(cLanguages)) + uintptr(i)*unsafe.Sizeof((*C.char)(nil))))
		languages[i] = C.GoString(langPtr)
	}

	return languages, nil
}

// SetLanguage sets the language for ASR transcription
func (p *ASR) SetLanguage(language string) {
	cLanguage := C.CString(language)
	defer C.free(unsafe.Pointer(cLanguage))

	C.ml_asr_set_language(p.ptr, cLanguage)
}
