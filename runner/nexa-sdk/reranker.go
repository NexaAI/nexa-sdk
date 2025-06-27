package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"

import (
	"unsafe"
)

type Reranker struct {
	ptr *C.ml_Reranker
}

func NewReranker(model string, tokenizer *string, devices *string) *Reranker {
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))
	cTokenizer := C.CString(*tokenizer) //TODO
	defer C.free(unsafe.Pointer(cTokenizer))

	return &Reranker{
		ptr: C.ml_reranker_create(cModel, cTokenizer, nil),
	}
}

func (p *Reranker) Destroy() {
	C.ml_reranker_destroy(p.ptr)
	p.ptr = nil
}

// Reset implements service.keepable.
func (p *Reranker) Reset() {
	panic("unimplemented")
}

func (p *Reranker) Rerank(query string, texts []string) ([]float32, error) {
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
		return nil, ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	ret := make([]float32, resLen)
	copy(ret, (*[1 << 30]float32)(unsafe.Pointer(res))[:resLen])
	return ret, nil
}
