package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"

import (
	"unsafe"

	"log/slog"
)

type Reranker struct {
	ptr *C.ml_Reranker
}

func NewReranker(model string, tokenizer *string, devices *string) (*Reranker, error) {
	slog.Debug("NewReranker called", "model", model, "tokenizer", tokenizer, "devices", devices)

	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	var cTokenizer *C.char

	if tokenizer != nil {
		cTokenizer = C.CString(*tokenizer)
		defer C.free(unsafe.Pointer(cTokenizer))

	}

	ptr := C.ml_reranker_create(cModel, cTokenizer, nil)
	if ptr == nil {
		return nil, SDKErrorModelLoad
	}
	return &Reranker{ptr: ptr}, nil
}

func (p *Reranker) Destroy() {
	slog.Debug("Destroy called", "ptr", p.ptr)

	C.ml_reranker_destroy(p.ptr)
	p.ptr = nil
}

// Reset implements service.keepable.
func (p *Reranker) Reset() {
	slog.Debug("Reset called", "ptr", p.ptr)

}

func (p *Reranker) Rerank(query string, texts []string) ([]float32, error) {
	slog.Debug("Rerank called", "query", query, "texts", texts)

	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))

	cTexts := make([]*C.char, len(texts))
	for i, text := range texts {
		cText := &cTexts[i]
		*cText = C.CString(text)
		defer C.free(unsafe.Pointer(*cText))
	}

	config := C.ml_RerankConfig{}

	var res *C.float
	resLen := C.ml_reranker_rerank(p.ptr, cQuery, &cTexts[0], C.int32_t(len(cTexts)), &config, &res)
	if resLen <= 0 {
		return nil, SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	ret := make([]float32, resLen)
	copy(ret, (*[1 << 30]float32)(unsafe.Pointer(res))[:resLen])
	return ret, nil
}

// GetProfilingData retrieves performance metrics from the Reranker instance
func (p *Reranker) GetProfilingData() (*ProfilingData, error) {
	slog.Debug("GetProfilingData called")

	var cData C.ml_ProfilingData
	res := C.ml_reranker_get_profiling_data(p.ptr, &cData)
	if res < 0 {
		return nil, SDKError(res)
	}

	return NewProfilingDataFromC(cData), nil
}
