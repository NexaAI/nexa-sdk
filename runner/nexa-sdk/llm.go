package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"

extern bool go_generate_stream_on_token(char*, void*);
*/
import "C"

import (
	"context"
	"log/slog"
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
	slog.Debug("NewLLM called", "model", model, "tokenizer", tokenizer, "ctxLen", ctxLen, "devices", devices)
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	ptr := C.ml_llm_create(cModel, nil, C.ml_ModelConfig{n_ctx: C.int32_t(ctxLen)}, nil)
	if ptr == nil {
		return nil, SDKErrorModelLoad
	}
	return &LLM{ptr: ptr}, nil
}

// Destroy frees the memory allocated for the LLM instance
func (p *LLM) Destroy() {
	slog.Debug("Destroy called", "ptr", p.ptr)
	C.ml_llm_destroy(p.ptr)
	p.ptr = nil
}

// Reset clears the LLM's internal state and context
func (p *LLM) Reset() {
	slog.Debug("Reset called", "ptr", p.ptr)
	C.ml_llm_reset(p.ptr)
}

// Encode converts a text message into token IDs using the model's tokenizer
func (p *LLM) Encode(msg string) ([]int32, error) {
	slog.Debug("Encode called", "msg", msg)
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cMsg))

	var res *C.int32_t
	resLen := C.ml_llm_encode(p.ptr, cMsg, &res)
	if resLen < 0 {
		return nil, SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	ids := make([]int32, resLen)
	copy(ids, (*[1 << 30]int32)(unsafe.Pointer(res))[:resLen])
	return ids, nil
}

// Decode converts token IDs back into text using the model's tokenizer
func (p *LLM) Decode(ids []int32) (string, error) {
	slog.Debug("Decode called", "ids", ids)
	var res *C.char
	resLen := C.ml_llm_decode(
		p.ptr,
		(*C.int32_t)(unsafe.Pointer(&ids[0])),
		C.int32_t(len(ids)),
		&res)
	if resLen < 0 {
		return "", SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	result := C.GoString(res)
	return result, nil
}

// SaveKVCache saves the model's key-value cache to disk for later reuse
func (p *LLM) SaveKVCache(path string) error {
	slog.Debug("SaveKVCache called", "path", path)
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_llm_save_kv_cache(p.ptr, cPath)
	if res < 0 {
		return SDKError(res)
	}
	return nil
}

// LoadKVCache loads a previously saved key-value cache from disk
func (p *LLM) LoadKVCache(path string) error {
	slog.Debug("LoadKVCache called", "path", path)
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	res := C.ml_llm_load_kv_cache(p.ptr, cPath)
	if res < 0 {
		return SDKError(res)
	}
	return nil
}

// Generate produces text completion for the given prompt
func (p *LLM) Generate(prompt string) (string, error) {
	slog.Debug("Generate called", "prompt", prompt)
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	config := C.ml_GenerationConfig{}
	config.max_tokens = 2048

	var res *C.char
	resLen := C.ml_llm_generate(p.ptr, cPrompt, &config, &res)
	if resLen <= 0 {
		return "", SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	result := C.GoString(res)
	return result, nil
}

// GetChatTemplate retrieves the chat template for formatting conversations
func (p *LLM) GetChatTemplate(name *string) (string, error) {
	slog.Debug("GetChatTemplate called", "name", name)
	var cName *C.char
	if name != nil {
		cName = C.CString(*name)
		defer C.free(unsafe.Pointer(cName))
	}

	var res *C.char
	resLen := C.ml_llm_get_chat_template(p.ptr, cName, &res)
	if resLen < 0 {
		return "", SDKError(resLen)
	}

	result := C.GoString(res)
	return result, nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *LLM) ApplyChatTemplate(msgs []ChatMessage) (string, error) {
	slog.Debug("ApplyChatTemplate called", "msgs", msgs)
	cMsgs := make([]C.ml_ChatMessage, len(msgs))

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
		return "", SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	result := C.GoString(res)
	return result, nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *LLM) ApplyJinjaTemplate(param ChatTemplateParam) (string, error) {
	slog.Debug("ApplyJinjaTemplate called", "param", param)
	chatTmpl, e := p.GetChatTemplate(nil)
	if e != nil {
		return "", e
	}

	chatTmpl = strings.ReplaceAll(chatTmpl, `messages|length`, `(messages|length)`)
	chatTmpl = strings.ReplaceAll(chatTmpl, `not(`, `not (`)
	tmpl, e := gonja.FromString(chatTmpl)
	if e != nil {
		return "", e
	}

	msgData, _ := sonic.Marshal(param)
	m := make(map[string]any)
	sonic.Unmarshal(msgData, &m)

	result, err := tmpl.ExecuteToString(exec.NewContext(m))
	if err != nil {
		return "", err
	}
	return result, nil
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
	slog.Debug("GenerateStream called", "prompt", prompt)
	cPrompt := C.CString(prompt)

	config := C.ml_GenerationConfig{}
	config.max_tokens = 2048

	if streamTokenCh != nil {
		panic("not support GenerateStream in parallel")
	}
	stream := make(chan string, 10)
	err := make(chan error, 1)
	streamTokenCh = stream
	streamTokenCtx = ctx

	go func() {
		defer func() {
			streamTokenCh = nil
			streamTokenCtx = nil
			close(err)
			close(stream)
			C.free(unsafe.Pointer(cPrompt))
		}()

		resLen := C.ml_llm_generate_stream(p.ptr, cPrompt, &config,
			(C.ml_llm_token_callback)(C.go_generate_stream_on_token), nil, nil)
		if resLen < 0 {
			err <- SDKError(resLen)
		}
	}()

	return stream, err
}

// GetProfilingData retrieves performance metrics from the LLM instance
func (p *LLM) GetProfilingData() (*ProfilingData, error) {
	slog.Debug("GetProfilingData called")
	var cData C.ml_ProfilingData
	res := C.ml_llm_get_profiling_data(p.ptr, &cData)
	if res < 0 {
		return nil, SDKError(res)
	}

	profiling := NewProfilingDataFromC(cData)
	return profiling, nil
}
