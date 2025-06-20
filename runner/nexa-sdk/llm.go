package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"

extern void go_generate_stream_on_token(char*, void*);
*/
import "C"

import (
	"unsafe"
)

type LLMRole string

const (
	LLMRoleSystem    = "system"
	LLMRoleUser      = "user"
	LLMRoleAssistant = "assistant"
)

type ChatMessage struct {
	Role    LLMRole
	Content string
}

type LLM struct {
	pointer *C.struct_ml_LLM
}

func NewLLM(model string, tokenizer *string, ctxLen int32, devices *string) LLM {
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	return LLM{
		pointer: C.ml_llm_create(cModel, nil, C.int32_t(ctxLen), nil),
	}
}

func (p *LLM) Destroy() {
	C.ml_llm_destroy(p.pointer)
	p.pointer = nil
}

func (p *LLM) Reset() {
	C.ml_llm_reset(p.pointer)
}

func (p *LLM) Encode(msg string) ([]int32, error) {
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cMsg))

	var res *C.int32_t
	resLen := C.ml_llm_encode(p.pointer, cMsg, &res)
	if resLen < 0 {
		return nil, ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	ids := make([]int32, resLen)
	copy(ids, (*[1 << 30]int32)(unsafe.Pointer(res))[:resLen])

	return ids, nil
}

func (p *LLM) Decode(ids []int32) (string, error) {
	var res *C.char
	resLen := C.ml_llm_decode(
		p.pointer,
		(*C.int32_t)(unsafe.Pointer(&ids[0])),
		C.int32_t(len(ids)),
		&res)
	if resLen < 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

func (p *LLM) SaveKVCache(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_llm_save_kv_cache(p.pointer, cPath, 0)
	if res < 0 {
		return ErrCommon
	}
	return nil
}

func (p *LLM) LoadKVCache(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_llm_load_kv_cache(p.pointer, cPath, 0)
	if res < 0 {
		return ErrCommon
	}
	return nil
}

func (p *LLM) Generate(prompt string) (string, error) {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	var res *C.char
	resLen := C.ml_llm_generate(p.pointer, cPrompt, nil, &res)
	if resLen <= 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

func (p *LLM) GetChatTemplate(name *string) (string, error) {
	var cName *C.char
	if name != nil {
		cName = C.CString(*name)
		defer C.free(unsafe.Pointer(cName))
	}

	var res *C.char
	resLen := C.ml_llm_get_chat_template(p.pointer, cName, &res)
	if resLen < 0 {
		return "", ErrCommon
	}

	return C.GoString(res), nil
}

// deprecated
func (p *LLM) ApplyChatTemplate(msgs []ChatMessage) (string, error) {
	cMsgs := make([]C.ml_ChatMessage, len(msgs))

	for i, msg := range msgs {
		cMsg := &cMsgs[i]
		cMsg.role = C.CString(string(msg.Role))
		defer C.free(unsafe.Pointer(cMsg.role))
		cMsg.content = C.CString(msg.Content)
		defer C.free(unsafe.Pointer(cMsg.content))
	}

	var res *C.char
	resLen := C.ml_llm_apply_chat_template(p.pointer, &cMsgs[0], C.int32_t(len(msgs)), &res)
	if resLen < 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// TODO: global channel mapping
var channel chan<- string

//export go_generate_stream_on_token
func go_generate_stream_on_token(token *C.char, _ *C.void) {
	channel <- C.GoString(token)
}

func (p *LLM) GenerateStream(prompt string) (<-chan string, <-chan error) {
	cPrompt := C.CString(prompt)

	config := C.ml_GenerationConfig{}
	config.max_tokens = 32

	stream := make(chan string, 10)
	err := make(chan error, 1)
	if channel != nil {
		panic("not support GenerateStream in parallel")
	}
	channel = stream

	// Start a goroutine to handle the streaming
	go func() {
		defer func() { channel = nil }()
		defer close(stream)
		defer close(err)
		defer C.free(unsafe.Pointer(cPrompt))

		// Call the C function to start streaming
		resLen := C.ml_llm_generate_stream(p.pointer, cPrompt, &config,
			(C.ml_llm_token_callback)(C.go_generate_stream_on_token), nil, nil)
		if resLen < 0 {
			err <- ErrCommon
		}
	}()

	return stream, err
}
