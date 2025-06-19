package nexa_sdk

/*
#cgo CFLAGS: -I../../build/include
#cgo LDFLAGS: -L../../build/lib -lbinding
#cgo LDFLAGS: -Wl,--unresolved-symbols=ignore-in-shared-libs

#include <stdlib.h>
#include "binding.h"
*/
import "C"

import (
	"errors"
	"unsafe"
)

type LLMPipeline struct {
	pointer *C.struct_LLMPipeline
	buffer  *C.char
}

func NewLLMPipeline() LLMPipeline {
	return LLMPipeline{
		pointer: C.llm_pipeline_new(),
		buffer:  (*C.char)(unsafe.Pointer(C.malloc(65535))),
	}
}

func (p LLMPipeline) LoadModel(path string) error {
	if !C.llm_pipeline_load_model(p.pointer, C.CString(path)) {
		return errors.New("LoadModel error")
	}
	return nil
}

func (p LLMPipeline) Generate(user string) (string, int, error) {
	user_c := C.CString(user)
	count := C.llm_pipeline_generate(p.pointer, user_c, p.buffer)
	C.free(unsafe.Pointer(user_c))
	if count <= 0 {
		return "", 0, errors.New("Generate error")
	}
	return C.GoString(p.buffer), int(count), nil
}

func (p LLMPipeline) GenerateStream(user string) error {
	if !C.llm_pipeline_generate_send(p.pointer, C.CString(user)) {
		return errors.New("Generate error")
	}
	return nil
}
func (p LLMPipeline) GenerateNextToken() (string, error) {
	count := C.llm_pipeline_generate_next_token(p.pointer, p.buffer)
	if count < 0 {
		return "", errors.New("Generate error")
	}
	if count == 0 {
		return "", nil
	}
	return C.GoString(p.buffer), nil
}

func (p LLMPipeline) Close() {
	C.llm_pipeline_close(p.pointer)
}

func (p LLMPipeline) Destroy() {
	C.llm_pipeline_free(p.pointer)
	p.pointer = nil
	C.free(unsafe.Pointer(p.buffer))
	p.buffer = nil
}
