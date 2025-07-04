package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"

extern bool go_generate_stream_on_token(char*, void*);
*/
import "C"

import (
	"context"
	"runtime"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

// VLM wraps the C library VLM structure and provides Go interface
type VLM struct {
	ptr *C.ml_VLM // Pointer to the underlying C VLM structure
}

// NewLLM creates a new VLM instance with the specified model and configuration
func NewVLM(model string, tokenizer *string, ctxLen int32, devices *string) *VLM {
	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	var cTokenizer *C.char
	if tokenizer != nil {
		cTokenizer = C.CString(*tokenizer)
		defer C.free(unsafe.Pointer(cTokenizer))
	}

	return &VLM{
		ptr: C.ml_vlm_create(cModel, cTokenizer, C.int32_t(ctxLen), nil),
	}
}

// Destroy frees the memory allocated for the VLM instance
func (p *VLM) Destroy() {
	C.ml_vlm_destroy(p.ptr)
	p.ptr = nil
}

// Reset clears the VLM's internal state and context
func (p *VLM) Reset() {
	C.ml_vlm_reset(p.ptr)
}

// Encode converts a text message into token IDs using the model's tokenizer
func (p *VLM) Encode(msg string) ([]int32, error) {
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cMsg))

	var res *C.int32_t
	resLen := C.ml_vlm_encode(p.ptr, cMsg, &res)
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
func (p *VLM) Decode(ids []int32) (string, error) {
	var res *C.char
	resLen := C.ml_vlm_decode(
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

// Generate produces text completion for the given prompt
func (p *VLM) Generate(prompt string, images []string, audios []string) (string, error) {
	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	var pinnner runtime.Pinner
	defer pinnner.Unpin()

	// Configure generation parameters
	config := C.ml_GenerationConfig{}
	config.max_tokens = 4096
	if len(images) > 0 {
		cImages := make([]C.ml_Path, len(images))
		for i, image := range images {
			cImage := C.CString(string(image))
			cImages[i] = cImage
			defer C.free(unsafe.Pointer(cImage))
		}
		config.image_paths = &cImages[0]
		config.image_count = C.int32_t(len(images))
		pinnner.Pin(&cImages[0])
	}
	if len(audios) > 0 {
		cAudios := make([]C.ml_Path, len(audios))
		for i, audio := range audios {
			cAudio := C.CString(string(audio))
			cAudios[i] = cAudio
			defer C.free(unsafe.Pointer(cAudio))
		}
		config.audio_paths = &cAudios[0]
		config.audio_count = C.int32_t(len(audios))
		pinnner.Pin(&cAudios[0])
	}

	var res *C.char
	resLen := C.ml_vlm_generate(p.ptr, cPrompt, &config, &res)
	if resLen <= 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// GetChatTemplate retrieves the chat template for formatting conversations
func (p *VLM) GetChatTemplate(name *string) (string, error) {
	var cName *C.char
	if name != nil {
		cName = C.CString(*name)
		defer C.free(unsafe.Pointer(cName))
	}

	var res *C.char
	resLen := C.ml_vlm_get_chat_template(p.ptr, cName, &res)
	if resLen < 0 {
		return "", ErrCommon
	}

	return C.GoString(res), nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *VLM) ApplyChatTemplate(msgs []ChatMessage) (string, error) {
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
	resLen := C.ml_vlm_apply_chat_template(p.ptr, &cMsgs[0], C.int32_t(len(msgs)), &res)
	if resLen < 0 {
		return "", ErrCommon
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *VLM) ApplyJinjaTemplate(param ChatTemplateParam) (string, error) {
	chatTmpl, e := p.GetChatTemplate(nil)
	if e != nil {
		return "", e
	}

	tmpl, e := gonja.FromString(chatTmpl)
	if e != nil {
		return "", e
	}

	msgData, _ := sonic.Marshal(param) // won't fail
	m := make(map[string]any)
	sonic.Unmarshal(msgData, &m) // won't fail

	return tmpl.ExecuteToString(exec.NewContext(m))
}

// GenerateStream generates text in streaming mode, returning tokens as they are produced
// Returns two channels: one for receiving tokens and one for errors
// Note: Currently does not support parallel streaming due to global channel usage
func (p *VLM) GenerateStream(ctx context.Context, prompt string, images []string, audios []string) (<-chan string, <-chan error) {
	cPrompt := C.CString(prompt)

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

		var pinnner runtime.Pinner
		defer pinnner.Unpin()
		// Configure generation parameters
		config := C.ml_GenerationConfig{}
		config.max_tokens = 4096
		if len(images) > 0 {
			cImages := make([]C.ml_Path, len(images))
			for i, image := range images {
				cImage := C.CString(string(image))
				cImages[i] = cImage
				defer C.free(unsafe.Pointer(cImage))
			}
			config.image_paths = &cImages[0]
			config.image_count = C.int32_t(len(images))
			pinnner.Pin(&cImages[0])
		}
		if len(audios) > 0 {
			cAudios := make([]C.ml_Path, len(audios))
			for i, audio := range audios {
				cAudio := C.CString(string(audio))
				cAudios[i] = cAudio
				defer C.free(unsafe.Pointer(cAudio))
			}
			config.audio_paths = &cAudios[0]
			config.audio_count = C.int32_t(len(audios))
			pinnner.Pin(&cAudios[0])
		}

		// Start streaming generation
		resLen := C.ml_vlm_generate_stream(p.ptr, cPrompt, &config,
			(C.ml_llm_token_callback)(C.go_generate_stream_on_token),
			nil, nil)
		if resLen < 0 {
			err <- ErrCommon
		}
	}()

	return stream, err
}
