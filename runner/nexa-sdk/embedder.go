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

type Embedder struct {
	ptr *C.ml_Embedder
}

func NewEmbedder(model string, tokenizer *string, devices *string) (*Embedder, error) {
	slog.Debug("NewEmbedder called", "model", model, "tokenizer", tokenizer, "devices", devices)
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	ptr := C.ml_embedder_create(cModel, nil, nil)
	if ptr == nil {
		slog.Debug("NewEmbedder failed", "error", SDKErrorModelLoad)
		return nil, SDKErrorModelLoad
	}
	slog.Debug("NewEmbedder success", "ptr", ptr)
	return &Embedder{ptr: ptr}, nil
}

func (p *Embedder) Destroy() {
	slog.Debug("Destroy called", "ptr", p.ptr)
	C.ml_embedder_destroy(p.ptr)
	p.ptr = nil
}

// Reset implements service.keepable.
func (p *Embedder) Reset() {
	slog.Debug("Reset called", "ptr", p.ptr)
}

func (p *Embedder) Embed(texts []string) ([]float32, error) {
	slog.Debug("Embed called", "texts", texts)
	cTexts := make([]*C.char, len(texts))
	for i, text := range texts {
		cText := &cTexts[i]
		*cText = C.CString(text)
		defer C.free(unsafe.Pointer(*cText))
	}

	config := C.ml_EmbeddingConfig{}

	var res *C.float
	resLen := C.ml_embedder_embed(p.ptr, &cTexts[0], C.int32_t(len(cTexts)), &config, &res)
	if resLen <= 0 {
		slog.Debug("Embed failed", "error", SDKError(resLen))
		return nil, SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	ret := make([]float32, resLen)
	copy(ret, (*[1 << 30]float32)(unsafe.Pointer(res))[:resLen])
	slog.Debug("Embed success", "result", ret)
	return ret, nil
}

// GetProfilingData retrieves performance metrics from the Embedder instance
func (p *Embedder) GetProfilingData() (*ProfilingData, error) {
	slog.Debug("GetProfilingData called")
	var cData C.ml_ProfilingData
	res := C.ml_embedder_get_profiling_data(p.ptr, &cData)
	if res < 0 {
		slog.Debug("GetProfilingData failed", "error", SDKError(res))
		return nil, SDKError(res)
	}

	profiling := NewProfilingDataFromC(cData)
	slog.Debug("GetProfilingData success", "profiling", profiling)
	return profiling, nil
}
