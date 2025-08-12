package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"

import (
	"unsafe"
)

/* ========================================================================== */
/*                              Utilities					  				 */
/* ========================================================================== */

func mlFree(ptr unsafe.Pointer) {
	C.ml_free(ptr)
}

func sliceToCCharArray(slice []string) (**C.char, C.int32_t) {
	if len(slice) == 0 {
		return nil, 0
	}
	count := C.size_t(len(slice))
	raw := C.malloc(C.size_t(count) * C.size_t(unsafe.Sizeof(uintptr(0))))
	if raw == nil {
		panic("C.malloc failed")
	}

	cArray := unsafe.Slice((**C.char)(raw), len(slice))
	for i, s := range slice {
		cArray[i] = C.CString(s)
	}
	return (**C.char)(raw), C.int32_t(count)
}

func cCharArrayToSlice(ptr **C.char, count C.int32_t) []string {
	if ptr == nil || count == 0 {
		return nil
	}
	arr := unsafe.Slice(ptr, int(count))
	out := make([]string, count)
	for i, cstr := range arr {
		out[i] = C.GoString(cstr)
	}
	return out
}

// free the C char array in Go side allocated
func freeCCharArray(ptr **C.char, count C.int32_t) {
	if ptr == nil || count == 0 {
		return
	}
	arr := unsafe.Slice(ptr, int(count))
	for _, p := range arr {
		C.free(unsafe.Pointer(p))
	}
	C.free(unsafe.Pointer(ptr))
}

func mlFreeCCharArray(ptr **C.char, count C.int32_t) {
	if ptr == nil || count == 0 {
		return
	}
	arr := unsafe.Slice(ptr, int(count))
	for _, p := range arr {
		mlFree(unsafe.Pointer(p))
	}
	mlFree(unsafe.Pointer(ptr))
}

/* ========================================================================== */
/*                              Callbacks								  */
/* ========================================================================== */

// TODO: use this to manage callbacks
// type CallbackManager struct {
// 	mu      sync.RWMutex
// 	callbacks map[unsafe.Pointer]OnTokenCallback
// }

type OnTokenCallback func(token string) bool

var onToken OnTokenCallback = nil

//export go_generate_stream_on_token
func go_generate_stream_on_token(token *C.char, _ *C.void) C.bool {
	if onToken == nil {
		return C.bool(true)
	}
	return C.bool(onToken(C.GoString(token)))
}

/* ========================================================================== */
/*                              Common Types & Utilities					  */
/* ========================================================================== */

type ProfileData struct {
	TTFT            int64
	PromptTime      int64
	DecodeTime      int64
	PromptTokens    int64
	GeneratedTokens int64
	AudioDuration   int64
	PrefillSpeed    float64
	DecodingSpeed   float64
	RealTimeFactor  float64
	StopReason      string
}

func (p ProfileData) TotalTokens() int64 {
	return p.PromptTokens + p.GeneratedTokens
}

func (p ProfileData) TotalTimeUs() int64 {
	return p.PromptTime + p.DecodeTime
}

func newProfileDataFromCPtr(c C.ml_ProfileData) ProfileData {
	return ProfileData{
		TTFT:            int64(c.ttft),
		PromptTime:      int64(c.prompt_time),
		DecodeTime:      int64(c.decode_time),
		PromptTokens:    int64(c.prompt_tokens),
		GeneratedTokens: int64(c.generated_tokens),
		AudioDuration:   int64(c.audio_duration),
		PrefillSpeed:    float64(c.prefill_speed),
		DecodingSpeed:   float64(c.decoding_speed),
		RealTimeFactor:  float64(c.real_time_factor),
		StopReason:      C.GoString(c.stop_reason),
	}
}

/* ========================================================================== */
/*                              LANGUAGE MODELS (LLM) */
/* ========================================================================== */
type SamplerConfig struct {
	Temperature       float32
	TopP              float32
	TopK              int32
	MinP              float32
	RepetitionPenalty float32
	PresencePenalty   float32
	FrequencyPenalty  float32
	Seed              int32
	GrammarPath       string
	GrammarString     string

	C *C.ml_SamplerConfig
}

func (sc SamplerConfig) toCPtr() *C.ml_SamplerConfig {
	// Allocate C structure
	cPtr := (*C.ml_SamplerConfig)(C.malloc(C.sizeof_ml_SamplerConfig))
	*cPtr = C.ml_SamplerConfig{}

	cPtr.temperature = C.float(sc.Temperature)
	cPtr.top_p = C.float(sc.TopP)
	cPtr.top_k = C.int32_t(sc.TopK)
	cPtr.min_p = C.float(sc.MinP)
	cPtr.repetition_penalty = C.float(sc.RepetitionPenalty)
	cPtr.presence_penalty = C.float(sc.PresencePenalty)
	cPtr.frequency_penalty = C.float(sc.FrequencyPenalty)
	cPtr.seed = C.int32_t(sc.Seed)

	if sc.GrammarPath != "" {
		cPtr.grammar_path = C.CString(sc.GrammarPath)
	}
	if sc.GrammarString != "" {
		cPtr.grammar_string = C.CString(sc.GrammarString)
	}

	return cPtr
}

func freeSamplerConfig(cPtr *C.ml_SamplerConfig) {
	if cPtr != nil {
		if cPtr.grammar_path != nil {
			C.free(unsafe.Pointer(cPtr.grammar_path))
		}
		if cPtr.grammar_string != nil {
			C.free(unsafe.Pointer(cPtr.grammar_string))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

type GenerationConfig struct {
	MaxTokens     int32
	Stop          []string
	NPast         int32
	SamplerConfig *SamplerConfig
	ImagePaths    []string
	AudioPaths    []string
}

func (gc GenerationConfig) toCPtr() *C.ml_GenerationConfig {
	// Allocate C structure
	cPtr := (*C.ml_GenerationConfig)(C.malloc(C.sizeof_ml_GenerationConfig))
	*cPtr = C.ml_GenerationConfig{}

	cPtr.max_tokens = C.int32_t(gc.MaxTokens)
	cPtr.n_past = C.int32_t(gc.NPast)

	if len(gc.Stop) > 0 {
		cPtr.stop, cPtr.stop_count = sliceToCCharArray(gc.Stop)
	}

	if len(gc.ImagePaths) > 0 {
		imagePaths, imageCount := sliceToCCharArray(gc.ImagePaths)
		cPtr.image_paths = (*C.ml_Path)(imagePaths)
		cPtr.image_count = C.int32_t(imageCount)
	}

	if len(gc.AudioPaths) > 0 {
		audioPaths, audioCount := sliceToCCharArray(gc.AudioPaths)
		cPtr.audio_paths = (*C.ml_Path)(audioPaths)
		cPtr.audio_count = C.int32_t(audioCount)
	}

	if gc.SamplerConfig != nil {
		cPtr.sampler_config = gc.SamplerConfig.toCPtr()
	}

	return cPtr
}

func freeGenerationConfig(ptr *C.ml_GenerationConfig) {
	if ptr == nil {
		return
	}

	freeCCharArray(ptr.stop, ptr.stop_count)
	freeCCharArray((**C.char)(unsafe.Pointer(ptr.image_paths)), ptr.image_count)
	freeCCharArray((**C.char)(unsafe.Pointer(ptr.audio_paths)), ptr.audio_count)

	if ptr.sampler_config != nil {
		freeSamplerConfig(ptr.sampler_config)
	}
}

type ModelConfig struct {
	NCtx                int32
	NThreads            int32
	NThreadsBatch       int32
	NBatch              int32
	NUbatch             int32
	NSeqMax             int32
	ChatTemplatePath    string
	ChatTemplateContent string
}

// TODO: check this if it's needed, llm has it self.
type KvCacheSaveInput struct {
	Path string
}

func (kci KvCacheSaveInput) toCPtr() *C.ml_KvCacheSaveInput {
	// Allocate C structure
	cPtr := (*C.ml_KvCacheSaveInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_KvCacheSaveInput{}))))
	*cPtr = C.ml_KvCacheSaveInput{}

	cPtr.path = C.CString(kci.Path)

	return cPtr
}

func freeKvCacheSaveInput(cPtr *C.ml_KvCacheSaveInput) {
	if cPtr != nil {
		if cPtr.path != nil {
			C.free(unsafe.Pointer(cPtr.path))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

type KvCacheSaveOutput struct {
	Reserved interface{}
}

func freeKvCacheSaveOutput(ptr *C.ml_KvCacheSaveOutput) {
	if ptr == nil {
		return
	}
}

type KVCacheLoadInput struct {
	Path string
}

func (kci KVCacheLoadInput) toCPtr() *C.ml_KvCacheLoadInput {
	// Allocate C structure
	cPtr := (*C.ml_KvCacheLoadInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_KvCacheLoadInput{}))))
	*cPtr = C.ml_KvCacheLoadInput{}
	cPtr.path = C.CString(kci.Path)
	return cPtr
}

func freeKVCacheLoadInput(cPtr *C.ml_KvCacheLoadInput) {
	if cPtr != nil {
		if cPtr.path != nil {
			C.free(unsafe.Pointer(cPtr.path))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

type KVCacheLoadOutput struct {
	// Reserved interface{}
}

// func freeKVCacheLoadOutput(ptr *C.ml_KvCacheLoadOutput) {
// 	if ptr == nil {
// 		return
// 	}
// }

// Tool represents a tool definition
type Tool struct {
	Type     string
	Function *ToolFunction
}

type tools []Tool

func (vts tools) toCPtr() (*C.ml_Tool, C.int32_t) {
	if len(vts) == 0 {
		return nil, 0
	}

	count := len(vts)
	raw := C.malloc(C.size_t(count * C.sizeof_ml_Tool))
	cTools := unsafe.Slice((*C.ml_Tool)(raw), count)

	for i, vt := range vts {
		if vt.Type != "" {
			cTools[i]._type = C.CString(vt.Type)
		}
		if vt.Function != nil {
			cTools[i].function = vt.Function.toCPtr()
		}
	}

	return (*C.ml_Tool)(raw), C.int32_t(count)
}

func freeTools(cPtr *C.ml_Tool, count C.int32_t) {
	if cPtr == nil || count == 0 {
		return
	}

	cTools := unsafe.Slice(cPtr, int(count))
	for i := range count {
		if cTools[i]._type != nil {
			C.free(unsafe.Pointer(cTools[i]._type))
		}
		if cTools[i].function != nil {
			freeToolFunction(cTools[i].function)
		}
	}

	C.free(unsafe.Pointer(cPtr))
}
