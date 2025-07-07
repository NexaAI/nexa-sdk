package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"

extern bool go_generate_stream_on_token(char*, void*);
*/
import "C"

import (
	"context"
	"strings"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

// LLMRole represents different roles in a chat conversation
type LLMRole string

const (
	LLMRoleSystem    = "system"    // System role for instructions
	LLMRoleUser      = "user"      // User role for queries
	LLMRoleAssistant = "assistant" // Assistant role for responses
)

// ChatMessage represents a single message in a chat conversation
type ChatMessage struct {
	Role    LLMRole `json:"role"`    // The role of the message sender
	Content string  `json:"content"` // The actual message content
}

type ChatToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	Strict      bool           `json:"strict,omitempty"`
}

type ChatTool struct {
	Type     string           `json:"type"`
	Function ChatToolFunction `json:"function"`
}

type ChatTemplateParam struct {
	Messages []ChatMessage `json:"messages,omitempty"`
	Tools    []ChatTool    `json:"tools,omitempty"`
}

// LLM wraps the C library LLM structure and provides Go interface
type LLM struct {
	ptr *C.ml_LLM // Pointer to the underlying C LLM structure
}

// NewLLM creates a new LLM instance with the specified model and configuration
func NewLLM(model string, tokenizer *string, ctxLen int32, devices *string) (*LLM, error) {
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	ptr := C.ml_llm_create(cModel, nil, C.ml_ModelConfig{n_ctx: C.int32_t(ctxLen)}, nil)
	if ptr == nil {
		return nil, ErrCreateFailed
	}

	return &LLM{ptr: ptr}, nil
}

// Destroy frees the memory allocated for the LLM instance
func (p *LLM) Destroy() {
	C.ml_llm_destroy(p.ptr)
	p.ptr = nil
}

// Reset clears the LLM's internal state and context
func (p *LLM) Reset() {
	C.ml_llm_reset(p.ptr)
}

// Encode converts a text message into token IDs using the model's tokenizer
func (p *LLM) Encode(msg string) ([]int32, error) {
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cMsg))

	var res *C.int32_t
	resLen := C.ml_llm_encode(p.ptr, cMsg, &res)
	if resLen < 0 {
		return nil, ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	// Copy C array to Go slice
	ids := make([]int32, resLen)
	copy(ids, (*[1 << 30]int32)(unsafe.Pointer(res))[:resLen])

	return ids, nil
}

// Decode converts token IDs back into text using the model's tokenizer
func (p *LLM) Decode(ids []int32) (string, error) {
	var res *C.char
	resLen := C.ml_llm_decode(
		p.ptr,
		(*C.int32_t)(unsafe.Pointer(&ids[0])),
		C.int32_t(len(ids)),
		&res)
	if resLen < 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// SaveKVCache saves the model's key-value cache to disk for later reuse
func (p *LLM) SaveKVCache(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_llm_save_kv_cache(p.ptr, cPath)
	if res < 0 {
		return ErrCommon
	}
	return nil
}

// LoadKVCache loads a previously saved key-value cache from disk
func (p *LLM) LoadKVCache(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_llm_load_kv_cache(p.ptr, cPath)
	if res < 0 {
		return ErrCommon
	}
	return nil
}

// Generate produces text completion for the given prompt
func (p *LLM) Generate(prompt string) (string, error) {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	// Configure generation parameters
	config := C.ml_GenerationConfig{}
	config.max_tokens = 2048

	var res *C.char
	resLen := C.ml_llm_generate(p.ptr, cPrompt, &config, &res)
	if resLen <= 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// GetChatTemplate retrieves the chat template for formatting conversations
func (p *LLM) GetChatTemplate(name *string) (string, error) {
	var cName *C.char
	if name != nil {
		cName = C.CString(*name)
		defer C.free(unsafe.Pointer(cName))
	}

	var res *C.char
	resLen := C.ml_llm_get_chat_template(p.ptr, cName, &res)
	if resLen < 0 {
		return "", ErrCommon
	}

	return C.GoString(res), nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *LLM) ApplyChatTemplate(msgs []ChatMessage) (string, error) {
	cMsgs := make([]C.ml_ChatMessage, len(msgs))

	// Convert Go chat messages to C structures
	for i, msg := range msgs {
		cMsg := &cMsgs[i]
		cMsg.role = C.CString(string(msg.Role))
		defer C.free(unsafe.Pointer(cMsg.role))
		cMsg.content = C.CString(msg.Content)
		defer C.free(unsafe.Pointer(cMsg.content))
	}

	var res *C.char
	resLen := C.ml_llm_apply_chat_template(p.ptr, &cMsgs[0], C.int32_t(len(msgs)), &res)
	if resLen < 0 {
		if resLen == -1 {
			return "", ErrChatTemplateNotFound
		}

		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *LLM) ApplyJinjaTemplate(param ChatTemplateParam) (string, error) {
	chatTmpl, e := p.GetChatTemplate(nil)
	if e != nil {
		return "", e
	}

	// TODO: Remove replace when gonja fixed.
	// Workaround for gonja template issues:
	// - https://github.com/NikolaLohinski/gonja/issues/48: wrap messages|length in parentheses
	// - https://github.com/NikolaLohinski/gonja/issues/49: ensure space after 'not('
	chatTmpl = strings.ReplaceAll(chatTmpl, `messages|length`, `(messages|length)`)
	chatTmpl = strings.ReplaceAll(chatTmpl, `not(`, `not (`)
	tmpl, e := gonja.FromString(chatTmpl)
	if e != nil {
		return "", e
	}

	msgData, _ := sonic.Marshal(param) // won't fail
	m := make(map[string]any)
	sonic.Unmarshal(msgData, &m) // won't fail

	return tmpl.ExecuteToString(exec.NewContext(m))
}

// Global streamTokenCh for streaming - TODO: implement proper streamTokenCh mapping for concurrent streams
var (
	streamTokenCh  chan<- string
	streamTokenCtx context.Context
)

// go_generate_stream_on_token is the callback function called by C code during streaming
// It sends each generated token to the Go channel
//
//export go_generate_stream_on_token
func go_generate_stream_on_token(token *C.char, _ *C.void) C.bool {
	select {
	case <-streamTokenCtx.Done():
		// fmt.Println("context done")
		return false
	default:
	}

	select {
	case streamTokenCh <- C.GoString(token):
		return true
	case <-streamTokenCtx.Done():
		// fmt.Println("cancel")
		return false
	}
}

// GenerateStream generates text in streaming mode, returning tokens as they are produced
// Returns two channels: one for receiving tokens and one for errors
// Note: Currently does not support parallel streaming due to global channel usage
func (p *LLM) GenerateStream(ctx context.Context, prompt string) (<-chan string, <-chan error) {
	cPrompt := C.CString(prompt)

	// Configure generation parameters
	config := C.ml_GenerationConfig{}
	config.max_tokens = 2048

	// check parallel call
	if streamTokenCh != nil {
		panic("not support GenerateStream in parallel")
	}
	// Create channels for streaming output
	stream := make(chan string, 10)
	err := make(chan error, 1)
	streamTokenCh = stream
	streamTokenCtx = ctx

	// Start streaming in a separate goroutine
	go func() {
		defer func() {
			streamTokenCh = nil
			streamTokenCtx = nil
			close(err)
			close(stream)
			C.free(unsafe.Pointer(cPrompt))
		}()

		// Call C function to start streaming generation
		resLen := C.ml_llm_generate_stream(p.ptr, cPrompt, &config,
			(C.ml_llm_token_callback)(C.go_generate_stream_on_token), nil, nil)
		if resLen < 0 {
			err <- ErrCommon
		}
	}()

	return stream, err
}

// GetProfilingData retrieves performance metrics from the LLM instance
func (p *LLM) GetProfilingData() (*ProfilingData, error) {
	var cData C.ml_ProfilingData
	res := C.ml_llm_get_profiling_data(p.ptr, &cData)
	if res < 0 {
		return nil, ErrCommon
	}

	return NewProfilingDataFromC(cData), nil
}
