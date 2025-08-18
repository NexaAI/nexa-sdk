package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"

extern bool go_generate_stream_on_token(char*, void*);
*/
import "C"

import (
	"log/slog"
	"path/filepath"
	"strings"
	"unsafe"
)

// TODO: use same role const with VLM and others
// LLMRole represents different roles in a chat conversation
type LLMRole string

const (
	LLMRoleSystem    LLMRole = "system"    // System role for instructions
	LLMRoleUser      LLMRole = "user"      // User role for queries
	LLMRoleAssistant LLMRole = "assistant" // Assistant role for responses
)

type LlmCreateInput struct {
	ModelPath     string
	TokenizerPath string
	Config        ModelConfig
	PluginID      string
	DeviceID      string
}

func (lci LlmCreateInput) toCPtr() *C.ml_LlmCreateInput {
	cPtr := (*C.ml_LlmCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_LlmCreateInput{}))))
	*cPtr = C.ml_LlmCreateInput{}

	if lci.ModelPath != "" {
		cPtr.model_path = C.CString(lci.ModelPath)
	}
	if lci.TokenizerPath != "" {
		cPtr.tokenizer_path = C.CString(lci.TokenizerPath)
	}
	if lci.PluginID != "" {
		cPtr.plugin_id = C.CString(lci.PluginID)
	}
	if lci.DeviceID != "" {
		cPtr.device_id = C.CString(lci.DeviceID)
	}

	// no need to use toCPtr() here, because we are not using the config pointer
	cPtr.config = C.ml_ModelConfig{
		n_ctx:           C.int32_t(lci.Config.NCtx),
		n_threads:       C.int32_t(lci.Config.NThreads),
		n_threads_batch: C.int32_t(lci.Config.NThreadsBatch),
		n_batch:         C.int32_t(lci.Config.NBatch),
		n_ubatch:        C.int32_t(lci.Config.NUbatch),
		n_seq_max:       C.int32_t(lci.Config.NSeqMax),
		//n_gpu_layers:    C.int32_t(lci.Config.NGpuLayers),
	}
	if lci.Config.ChatTemplatePath != "" {
		cPtr.config.chat_template_path = C.CString(lci.Config.ChatTemplatePath)
	}
	if lci.Config.ChatTemplateContent != "" {
		cPtr.config.chat_template_content = C.CString(lci.Config.ChatTemplateContent)
	}

	return cPtr
}

func freeLlmCreateInput(cPtr *C.ml_LlmCreateInput) {
	if cPtr != nil {
		// Free string fields
		if cPtr.model_path != nil {
			C.free(unsafe.Pointer(cPtr.model_path))
		}
		if cPtr.tokenizer_path != nil {
			C.free(unsafe.Pointer(cPtr.tokenizer_path))
		}
		if cPtr.plugin_id != nil {
			C.free(unsafe.Pointer(cPtr.plugin_id))
		}
		if cPtr.device_id != nil {
			C.free(unsafe.Pointer(cPtr.device_id))
		}

		// Free nested ModelConfig - config is a value, not a pointer
		// We need to free the string fields manually
		if cPtr.config.chat_template_path != nil {
			C.free(unsafe.Pointer(cPtr.config.chat_template_path))
		}
		if cPtr.config.chat_template_content != nil {
			C.free(unsafe.Pointer(cPtr.config.chat_template_content))
		}

		// Free the main structure
		C.free(unsafe.Pointer(cPtr))
	}
}

type LlmGenerateInput struct {
	PromptUTF8 string
	Config     *GenerationConfig
	OnToken    OnTokenCallback
	// UserData   unsafe.Pointer
}

func (lgi LlmGenerateInput) toCPtr() *C.ml_LlmGenerateInput {
	cPtr := (*C.ml_LlmGenerateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_LlmGenerateInput{}))))
	*cPtr = C.ml_LlmGenerateInput{}

	cPtr.prompt_utf8 = C.CString(lgi.PromptUTF8)
	if lgi.Config != nil {
		cPtr.config = lgi.Config.toCPtr()
	} else {
		cPtr.config = nil
	}

	// Note: on_token and user_data should be set by the caller
	cPtr.on_token = nil
	cPtr.user_data = nil

	return cPtr
}

func freeLlmGenerateInput(cPtr *C.ml_LlmGenerateInput) {
	if cPtr != nil {
		if cPtr.prompt_utf8 != nil {
			C.free(unsafe.Pointer(cPtr.prompt_utf8))
		}

		if cPtr.config != nil {
			freeGenerationConfig(cPtr.config)
		}

		C.free(unsafe.Pointer(cPtr))
	}
}

type LlmGenerateOutput struct {
	FullText    string
	ProfileData ProfileData
}

func newLlmGenerateOutputFromCPtr(c *C.ml_LlmGenerateOutput) LlmGenerateOutput {
	output := LlmGenerateOutput{}

	if c == nil {
		return output
	}

	if c.full_text != nil {
		output.FullText = C.GoString(c.full_text)
	}
	output.ProfileData = newProfileDataFromCPtr(c.profile_data)
	return output
}

func freeLlmGenerateOutput(ptr *C.ml_LlmGenerateOutput) {
	if ptr == nil {
		return
	}
	if ptr.full_text != nil {
		mlFree(unsafe.Pointer(ptr.full_text))
	}
}

type LlmChatMessage struct {
	Role    LLMRole
	Content string
}

type llmChatMessages []LlmChatMessage

func (lcm llmChatMessages) toCPtr() (*C.ml_LlmChatMessage, C.int32_t) {
	if len(lcm) == 0 {
		return nil, 0
	}

	count := len(lcm)
	raw := C.malloc(C.size_t(count * C.sizeof_ml_LlmChatMessage))
	cMessages := unsafe.Slice((*C.ml_LlmChatMessage)(raw), count)

	for i, msg := range lcm {
		if msg.Role != "" {
			cMessages[i].role = C.CString(string(msg.Role))
		}
		if msg.Content != "" {
			cMessages[i].content = C.CString(msg.Content)
		}
	}

	return (*C.ml_LlmChatMessage)(raw), C.int32_t(count)
}

func freeLlmChatMessages(cPtr *C.ml_LlmChatMessage, count C.int32_t) {
	if cPtr == nil || count == 0 {
		return
	}

	cMessages := unsafe.Slice(cPtr, int(count))
	for i := range count {
		if cMessages[i].role != nil {
			C.free(unsafe.Pointer(cMessages[i].role))
		}
		if cMessages[i].content != nil {
			C.free(unsafe.Pointer(cMessages[i].content))
		}
	}

	C.free(unsafe.Pointer(cPtr))
}

type LlmApplyChatTemplateInput struct {
	Messages    []LlmChatMessage
	Tools       []Tool
	EnableThink bool
}

func (lati LlmApplyChatTemplateInput) toCPtr() *C.ml_LlmApplyChatTemplateInput {
	cPtr := (*C.ml_LlmApplyChatTemplateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_LlmApplyChatTemplateInput{}))))
	*cPtr = C.ml_LlmApplyChatTemplateInput{}

	if len(lati.Messages) > 0 {
		cMessages, messageCount := llmChatMessages(lati.Messages).toCPtr()
		cPtr.messages = cMessages
		cPtr.message_count = C.int32_t(messageCount)
	}

	if len(lati.Tools) > 0 {
		cTools, toolCount := tools(lati.Tools).toCPtr()
		cPtr.tools = cTools
		cPtr.tool_count = C.int32_t(toolCount)
	}

	cPtr.enable_thinking = C.bool(lati.EnableThink)

	return cPtr
}

func freeLlmApplyChatTemplateInput(cPtr *C.ml_LlmApplyChatTemplateInput) {
	if cPtr == nil {
		return
	}
	freeLlmChatMessages(cPtr.messages, cPtr.message_count)
	freeTools(cPtr.tools, cPtr.tool_count)
	C.free(unsafe.Pointer(cPtr))
}

type LlmApplyChatTemplateOutput struct {
	FormattedText string
}

func newLlmApplyChatTemplateOutputFromCPtr(c *C.ml_LlmApplyChatTemplateOutput) LlmApplyChatTemplateOutput {
	output := LlmApplyChatTemplateOutput{}

	if c == nil {
		return output
	}

	if c.formatted_text != nil {
		output.FormattedText = C.GoString(c.formatted_text)
	}

	return output
}

func freeLlmApplyChatTemplateOutput(cPtr *C.ml_LlmApplyChatTemplateOutput) {
	if cPtr == nil {
		return
	}
	if cPtr.formatted_text != nil {
		mlFree(unsafe.Pointer(cPtr.formatted_text))
	}
}

type LLM struct {
	ptr *C.ml_LLM
}

func NewLLM(input LlmCreateInput) (*LLM, error) {

	slog.Debug("NewLLM called", "input", input)

	cInput := input.toCPtr()
	defer freeLlmCreateInput(cInput)

	// Qnn
	if cInput.model_path != nil {
		C.free(unsafe.Pointer(cInput.model_path))
	}
	if cInput.tokenizer_path != nil {
		C.free(unsafe.Pointer(cInput.tokenizer_path))
	}
	basePath := filepath.Dir(input.ModelPath)
	if strings.HasSuffix(basePath, "qwen3-npu") {
		cInput.model_path = C.CString(filepath.Join(basePath, "qwen3-npu", "weight_sharing_model_1_of_1_w8.serialized.bin"))
		cInput.tokenizer_path = C.CString(filepath.Join(basePath, "qwen3-npu", "tokenizer.json"))
		cInput.config.system_library_path = C.CString(filepath.Join(basePath, "htp-files-2.36", "QnnSystem.dll"))
		cInput.config.backend_library_path = C.CString(filepath.Join(basePath, "htp-files-2.36", "QnnHtp.dll"))
		cInput.config.extension_library_path = C.CString(filepath.Join(basePath, "htp-files-2.36", "QnnHtpNetRunExtensions.dll"))
		cInput.config.config_file_path = C.CString(filepath.Join(basePath, "qwen3-npu", "htp_backend_ext_config.json"))
		cInput.config.embedded_tokens_path = C.CString(filepath.Join(basePath, "qwen3-npu", "qwen3_embedding_layer.npy"))
	} else if strings.HasSuffix(basePath, "qwen3-4B-npu") {
		basePath := filepath.Dir(input.ModelPath)
		cInput.model_path = C.CString(filepath.Join(basePath, "qwen3-npu", "weight_sharing_model_1_of_2.serialized.bin"))
		cInput.tokenizer_path = C.CString(filepath.Join(basePath, "qwen3-npu", "tokenizer.json"))
		cInput.config.model_path_1 = C.CString(filepath.Join(basePath, "qwen3-npu", "weight_sharing_model_2_of_2.serialized.bin"))
		cInput.config.system_library_path = C.CString(filepath.Join(basePath, "htp-files-2.36", "QnnSystem.dll"))
		cInput.config.backend_library_path = C.CString(filepath.Join(basePath, "htp-files-2.36", "QnnHtp.dll"))
		cInput.config.extension_library_path = C.CString(filepath.Join(basePath, "htp-files-2.36", "QnnHtpNetRunExtensions.dll"))
		cInput.config.config_file_path = C.CString(filepath.Join(basePath, "qwen3-npu", "htp_backend_ext_config.json"))
		cInput.config.embedded_tokens_path = C.CString(filepath.Join(basePath, "qwen3-npu", "qwen3_embedding_layer.npy"))
		defer C.free(unsafe.Pointer(cInput.config.model_path_1))
	}
	defer C.free(unsafe.Pointer(cInput.config.system_library_path))
	defer C.free(unsafe.Pointer(cInput.config.backend_library_path))
	defer C.free(unsafe.Pointer(cInput.config.extension_library_path))
	defer C.free(unsafe.Pointer(cInput.config.config_file_path))
	defer C.free(unsafe.Pointer(cInput.config.embedded_tokens_path))
	cInput.config.max_tokens = 256
	cInput.config.enable_thinking = true
	cInput.config.verbose = false
	// Qnn
	var cHandle *C.ml_LLM
	res := C.ml_llm_create(cInput, &cHandle)
	if res < 0 {
		slog.Debug("Failed to create LLM", "error_code", res)
		return nil, SDKError(res)
	}

	slog.Debug("LLM created successfully", "ptr", cHandle)

	return &LLM{ptr: cHandle}, nil
}

func (l *LLM) Destroy() error {
	slog.Debug("Destroy called", "ptr", l.ptr)

	if l.ptr == nil {
		return nil
	}

	res := C.ml_llm_destroy(l.ptr)
	if res < 0 {
		return SDKError(res)
	}
	l.ptr = nil
	return nil
}

func (l *LLM) Reset() error {
	slog.Debug("Reset called", "ptr", l.ptr)
	if l.ptr == nil {
		return nil
	}

	res := C.ml_llm_reset(l.ptr)
	if res < 0 {
		return SDKError(res)
	}
	return nil
}

// LlmSaveKVCacheInput represents input for saving LLM KV cache
type LlmSaveKVCacheInput struct {
	Path string
}

func (lsci LlmSaveKVCacheInput) toCPtr() *C.ml_KvCacheSaveInput {
	cPtr := (*C.ml_KvCacheSaveInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_KvCacheSaveInput{}))))
	*cPtr = C.ml_KvCacheSaveInput{}

	if lsci.Path != "" {
		cPtr.path = C.CString(lsci.Path)
	}

	return cPtr
}

func freeLlmSaveKVCacheInput(cPtr *C.ml_KvCacheSaveInput) {
	if cPtr != nil {
		if cPtr.path != nil {
			C.free(unsafe.Pointer(cPtr.path))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// LlmSaveKVCacheOutput represents output from saving LLM KV cache
type LlmSaveKVCacheOutput struct {
	Reserved interface{}
}

func newLlmSaveKVCacheOutputFromCPtr(c *C.ml_KvCacheSaveOutput) LlmSaveKVCacheOutput {
	output := LlmSaveKVCacheOutput{}

	if c == nil {
		return output
	}

	// Currently no fields to copy from C structure
	return output
}

func freeLlmSaveKVCacheOutput(cPtr *C.ml_KvCacheSaveOutput) {
	if cPtr == nil {
		return
	}
	// Currently no fields to free
}

// LlmLoadKVCacheInput represents input for loading LLM KV cache
type LlmLoadKVCacheInput struct {
	Path string
}

func (llci LlmLoadKVCacheInput) toCPtr() *C.ml_KvCacheLoadInput {
	cPtr := (*C.ml_KvCacheLoadInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_KvCacheLoadInput{}))))
	*cPtr = C.ml_KvCacheLoadInput{}

	if llci.Path != "" {
		cPtr.path = C.CString(llci.Path)
	}

	return cPtr
}

func freeLlmLoadKVCacheInput(cPtr *C.ml_KvCacheLoadInput) {
	if cPtr != nil {
		if cPtr.path != nil {
			C.free(unsafe.Pointer(cPtr.path))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// LlmLoadKVCacheOutput represents output from loading LLM KV cache
type LlmLoadKVCacheOutput struct {
	Reserved any
}

func newLlmLoadKVCacheOutputFromCPtr(c *C.ml_KvCacheLoadOutput) LlmLoadKVCacheOutput {
	output := LlmLoadKVCacheOutput{}

	if c == nil {
		return output
	}

	// Currently no fields to copy from C structure
	return output
}

func freeLlmLoadKVCacheOutput(cPtr *C.ml_KvCacheLoadOutput) {
	if cPtr == nil {
		return
	}
	// Currently no fields to free
}

func (l *LLM) ApplyChatTemplate(input LlmApplyChatTemplateInput) (*LlmApplyChatTemplateOutput, error) {
	slog.Debug("ApplyChatTemplate called", "input", input)

	cinput := input.toCPtr()
	defer freeLlmApplyChatTemplateInput(cinput)

	var cOutput C.ml_LlmApplyChatTemplateOutput
	defer freeLlmApplyChatTemplateOutput(&cOutput)

	res := C.ml_llm_apply_chat_template(l.ptr, cinput, &cOutput)
	if res < 0 {
		return nil, SDKError(res)
	}

	output := newLlmApplyChatTemplateOutputFromCPtr(&cOutput)

	return &output, nil
}

func (l *LLM) SaveKVCache(input LlmSaveKVCacheInput) (*LlmSaveKVCacheOutput, error) {
	slog.Debug("SaveKVCache called", "input", input)

	cInput := input.toCPtr()
	defer freeLlmSaveKVCacheInput(cInput)

	var cOutput C.ml_KvCacheSaveOutput
	defer freeLlmSaveKVCacheOutput(&cOutput)

	res := C.ml_llm_save_kv_cache(l.ptr, cInput, &cOutput)
	if res < 0 {
		return nil, SDKError(res)
	}

	output := newLlmSaveKVCacheOutputFromCPtr(&cOutput)

	return &output, nil
}

func (l *LLM) LoadKVCache(input LlmLoadKVCacheInput) (*LlmLoadKVCacheOutput, error) {
	slog.Debug("LoadKVCache called", "input", input)

	cInput := input.toCPtr()
	defer freeLlmLoadKVCacheInput(cInput)

	var cOutput C.ml_KvCacheLoadOutput
	defer freeLlmLoadKVCacheOutput(&cOutput)

	res := C.ml_llm_load_kv_cache(l.ptr, cInput, &cOutput)
	if res < 0 {
		return nil, SDKError(res)
	}

	output := newLlmLoadKVCacheOutputFromCPtr(&cOutput)

	return &output, nil
}

func (l *LLM) Generate(input LlmGenerateInput) (LlmGenerateOutput, error) {
	slog.Debug("Generate called", "input", input)

	cInput := input.toCPtr()
	defer freeLlmGenerateInput(cInput)

	// set the callback
	onToken = input.OnToken
	defer func() {
		onToken = nil // reset to default
	}()
	cInput.on_token = C.ml_token_callback(C.go_generate_stream_on_token)

	var cOutput C.ml_LlmGenerateOutput
	defer freeLlmGenerateOutput(&cOutput)

	res := C.ml_llm_generate(l.ptr, cInput, &cOutput)
	if res < 0 {
		return LlmGenerateOutput{}, SDKError(res)
	}

	output := newLlmGenerateOutputFromCPtr(&cOutput)

	return output, nil
}
