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

// GetProfilingData retrieves performance metrics from the VLM instance
func (p *VLM) GetProfilingData() (*ProfilingData, error) {
	slog.Debug("GetProfilingData called")

	var cData C.ml_ProfilingData
	res := C.ml_vlm_get_profiling_data(p.ptr, &cData)
	if res < 0 {
		return nil, SDKError(res)
	}

	return NewProfilingDataFromC(cData), nil
}

// NewLLM creates a new VLM instance with the specified model and configuration
func NewVLM(model string, tokenizer *string, ctxLen int32, devices *string) (*VLM, error) {
	slog.Debug("NewVLM called", "model", model, "tokenizer", tokenizer, "ctxLen", ctxLen, "devices", devices)

	cModel := C.CString(model)
	defer C.free(unsafe.Pointer(cModel))

	var cTokenizer *C.char
	if tokenizer != nil {
		cTokenizer = C.CString(*tokenizer)
		defer C.free(unsafe.Pointer(cTokenizer))
	}

	ptr := C.ml_vlm_create(cModel, cTokenizer, C.ml_ModelConfig{n_ctx: C.int32_t(ctxLen)}, nil)
	if ptr == nil {
		return nil, SDKErrorModelLoad
	}
	return &VLM{ptr: ptr}, nil
}

// Destroy frees the memory allocated for the VLM instance
func (p *VLM) Destroy() {
	slog.Debug("Destroy called", "ptr", p.ptr)

	C.ml_vlm_destroy(p.ptr)
	p.ptr = nil
}

// Reset clears the VLM's internal state and context
func (p *VLM) Reset() {
	slog.Debug("Reset called", "ptr", p.ptr)

	C.ml_vlm_reset(p.ptr)
}

// Encode converts a text message into token IDs using the model's tokenizer
func (p *VLM) Encode(msg string) ([]int32, error) {
	slog.Debug("Encode called", "msg", msg)

	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cMsg))

	var res *C.int32_t
	resLen := C.ml_vlm_encode(p.ptr, cMsg, &res)
	if resLen < 0 {
		return nil, SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	ids := make([]int32, resLen)
	copy(ids, (*[1 << 30]int32)(unsafe.Pointer(res))[:resLen])
	return ids, nil
}

// Decode converts token IDs back into text using the model's tokenizer
func (p *VLM) Decode(ids []int32) (string, error) {
	slog.Debug("Decode called", "ids", ids)

	var res *C.char
	resLen := C.ml_vlm_decode(
		p.ptr,
		(*C.int32_t)(unsafe.Pointer(&ids[0])),
		C.int32_t(len(ids)),
		&res)
	if resLen < 0 {
		return "", SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// Generate produces text completion for the given prompt
func (p *VLM) Generate(prompt string, images []string, audios []string) (string, error) {
	slog.Debug("Generate called", "prompt", prompt, "images", images, "audios", audios)

	cPrompt := C.CString(prompt)
	defer C.free(unsafe.Pointer(cPrompt))

	var pinnner runtime.Pinner
	defer pinnner.Unpin()

	config := C.ml_GenerationConfig{}
	config.max_tokens = 2048
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
	resLen := C.ml_vlm_generate_stream(p.ptr, cPrompt, &config, nil, nil, &res)
	if resLen <= 0 {
		return "", SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// GetChatTemplate retrieves the chat template for formatting conversations
func (p *VLM) GetChatTemplate(name *string) (string, error) {
	slog.Debug("GetChatTemplate called", "name", name)

	var cName *C.char
	if name != nil {
		cName = C.CString(*name)
		defer C.free(unsafe.Pointer(cName))
	}

	var res *C.char
	resLen := C.ml_vlm_get_chat_template(p.ptr, cName, &res)
	if resLen < 0 {
		return "", SDKError(resLen)
	}

	return C.GoString(res), nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *VLM) ApplyChatTemplate(msgs []ChatMessage, images []string, audios []string) (string, error) {
	slog.Debug("ApplyChatTemplate called", "msgs", msgs, "images", images, "audios", audios)

	cMsgs := make([]C.ml_VlmChatMessage, len(msgs))

	// Calculate total content items needed per message
	// For the last message (user message), we add images and audios
	// For other messages, just text content
	totalContents := 0
	for i := range msgs {
		if i == len(msgs)-1 { // Last message gets media appended
			totalContents += 1 + len(images) + len(audios) // text + images + audios
		} else {
			totalContents += 1 // just text
		}
	}

	allContents := make([]C.ml_VlmContent, totalContents)

	var pinner runtime.Pinner
	defer pinner.Unpin()

	contentIndex := 0
	for i, msg := range msgs {
		cMsg := &cMsgs[i]
		cMsg.role = C.CString(string(msg.Role))
		defer C.free(unsafe.Pointer(cMsg.role))

		// Start with text content
		contentStartIndex := contentIndex

		// Create text content
		textContent := &allContents[contentIndex]
		textContent._type = C.CString("text")
		defer C.free(unsafe.Pointer(textContent._type))
		textContent.text = C.CString(msg.Content)
		defer C.free(unsafe.Pointer(textContent.text))
		contentIndex++

		// For the last message (typically user message), append images and audios
		if i == len(msgs)-1 {
			// Add image contents
			for _, imagePath := range images {
				imageContent := &allContents[contentIndex]
				imageContent._type = C.CString("image")
				defer C.free(unsafe.Pointer(imageContent._type))
				imageContent.text = C.CString(imagePath)
				defer C.free(unsafe.Pointer(imageContent.text))
				contentIndex++
			}

			// Add audio contents
			for _, audioPath := range audios {
				audioContent := &allContents[contentIndex]
				audioContent._type = C.CString("audio")
				defer C.free(unsafe.Pointer(audioContent._type))
				audioContent.text = C.CString(audioPath)
				defer C.free(unsafe.Pointer(audioContent.text))
				contentIndex++
			}
		}

		// Set up the message with its content items
		contentCount := contentIndex - contentStartIndex
		cMsg.content_count = C.int64_t(contentCount)
		cMsg.contents = &allContents[contentStartIndex]
	}

	// Pin the allContents slice to prevent GC from moving it
	if len(allContents) > 0 {
		pinner.Pin(&allContents[0])
	}

	var res *C.char
	resLen := C.ml_vlm_apply_chat_template(p.ptr, &cMsgs[0], C.int32_t(len(msgs)), nil, 0, &res)
	if resLen < 0 {
		return "", SDKError(resLen)
	}
	defer C.free(unsafe.Pointer(res))

	return C.GoString(res), nil
}

// ApplyChatTemplate formats chat messages using the model's chat template
func (p *VLM) ApplyJinjaTemplate(param ChatTemplateParam) (string, error) {
	slog.Debug("ApplyJinjaTemplate called", "param", param)

	chatTmpl, e := p.GetChatTemplate(nil)
	if e != nil {
		return "", e
	}
	tmpl, e := gonja.FromString(chatTmpl)
	if e != nil {
		return "", e
	}

	msgData, _ := sonic.Marshal(param)
	m := make(map[string]any)
	sonic.Unmarshal(msgData, &m)

	return tmpl.ExecuteToString(exec.NewContext(m))
}

// GenerateStream generates text in streaming mode, returning tokens as they are produced
// Returns two channels: one for receiving tokens and one for errors
// Note: Currently does not support parallel streaming due to global channel usage
func (p *VLM) GenerateStream(ctx context.Context, prompt string, images []string, audios []string) (<-chan string, <-chan error) {
	slog.Debug("GenerateStream called", "prompt", prompt, "images", images, "audios", audios)

	cPrompt := C.CString(prompt)

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

		var pinnner runtime.Pinner
		defer pinnner.Unpin()
		config := C.ml_GenerationConfig{}
		config.max_tokens = 2048
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

		resLen := C.ml_vlm_generate_stream(p.ptr, cPrompt, &config,
			(C.ml_llm_token_callback)(C.go_generate_stream_on_token),
			nil, nil)
		if resLen < 0 {
			err <- SDKError(resLen)
		}
	}()

	return stream, err
}
