package nexa_sdk

/*
#include <stdlib.h>
#include "ml.h"
*/
import "C"
import (
	"log/slog"
	"unsafe"
)

// ASRConfig represents ASR processing configuration
type ASRConfig struct {
	Timestamps string
	BeamSize   int32
	Stream     bool
}

// ASRStreamConfig represents ASR streaming configuration
type ASRStreamConfig struct {
	ChunkDuration   float32
	OverlapDuration float32
	SampleRate      int32
	MaxQueueSize    int32
	BufferSize      int32
	Timestamps      string
	BeamSize        int32
}

func (ac ASRConfig) toCPtr() *C.ml_ASRConfig {
	cPtr := (*C.ml_ASRConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_ASRConfig{}))))
	*cPtr = C.ml_ASRConfig{}

	if ac.Timestamps != "" {
		cPtr.timestamps = C.CString(ac.Timestamps)
	}
	cPtr.beam_size = C.int32_t(ac.BeamSize)
	cPtr.stream = C.bool(ac.Stream)

	return cPtr
}

func (asc ASRStreamConfig) toCPtr() *C.ml_ASRStreamConfig {
	cPtr := (*C.ml_ASRStreamConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_ASRStreamConfig{}))))
	*cPtr = C.ml_ASRStreamConfig{}

	cPtr.chunk_duration = C.float(asc.ChunkDuration)
	cPtr.overlap_duration = C.float(asc.OverlapDuration)
	cPtr.sample_rate = C.int32_t(asc.SampleRate)
	cPtr.max_queue_size = C.int32_t(asc.MaxQueueSize)
	cPtr.buffer_size = C.int32_t(asc.BufferSize)
	if asc.Timestamps != "" {
		cPtr.timestamps = C.CString(asc.Timestamps)
	}
	cPtr.beam_size = C.int32_t(asc.BeamSize)

	return cPtr
}

func freeASRConfig(cPtr *C.ml_ASRConfig) {
	if cPtr != nil {
		if cPtr.timestamps != nil {
			C.free(unsafe.Pointer(cPtr.timestamps))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

func freeASRStreamConfig(cPtr *C.ml_ASRStreamConfig) {
	if cPtr != nil {
		if cPtr.timestamps != nil {
			C.free(unsafe.Pointer(cPtr.timestamps))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// ASRResult represents ASR transcription result
type ASRResult struct {
	Transcript       string
	ConfidenceScores []float32
	Timestamps       []float32
}

func newASRResultFromCPtr(c *C.ml_ASRResult) ASRResult {
	result := ASRResult{}

	if c == nil {
		return result
	}

	if c.transcript != nil {
		result.Transcript = C.GoString(c.transcript)
	}

	// Convert confidence scores array
	if c.confidence_scores != nil && c.confidence_count > 0 {
		scores := unsafe.Slice((*C.float)(unsafe.Pointer(c.confidence_scores)), int(c.confidence_count))
		result.ConfidenceScores = make([]float32, c.confidence_count)
		for i := range result.ConfidenceScores {
			result.ConfidenceScores[i] = float32(scores[i])
		}
	}

	// Convert timestamps array
	if c.timestamps != nil && c.timestamp_count > 0 {
		timestamps := unsafe.Slice((*C.float)(unsafe.Pointer(c.timestamps)), int(c.timestamp_count))
		result.Timestamps = make([]float32, c.timestamp_count)
		for i := range result.Timestamps {
			result.Timestamps[i] = float32(timestamps[i])
		}
	}

	return result
}

func freeASRResult(ptr *C.ml_ASRResult) {
	if ptr == nil {
		return
	}
	if ptr.transcript != nil {
		mlFree(unsafe.Pointer(ptr.transcript))
	}
	if ptr.confidence_scores != nil {
		mlFree(unsafe.Pointer(ptr.confidence_scores))
	}
	if ptr.timestamps != nil {
		mlFree(unsafe.Pointer(ptr.timestamps))
	}
}

// ASRModelConfig represents ASR model configuration
type ASRModelConfig struct {
	NCtx              int32
	NThreads          int32
	NThreadsBatch     int32
	NBatch            int32
	NUbatch           int32
	NSeqMax           int32
	NGpuLayers        int32
	ChatTemplatePath  string
	ChatTemplateContent string
	EnableSampling    bool
	GrammarStr        string
	MaxTokens         int32
	EnableThinking    bool
	Verbose           bool
	QnnModelFolderPath string
	QnnLibFolderPath  string
}

// AsrCreateInput represents input parameters for ASR creation
type AsrCreateInput struct {
	ModelName     string
	ModelPath     string
	TokenizerPath string
	Config        ASRModelConfig
	Language      string
	PluginID      string
	DeviceID      string
	LicenseID     string
	LicenseKey    string
}

func (mc ASRModelConfig) toCPtr() *C.ml_ModelConfig {
	cPtr := (*C.ml_ModelConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_ModelConfig{}))))
	*cPtr = C.ml_ModelConfig{}

	cPtr.n_ctx = C.int32_t(mc.NCtx)
	cPtr.n_threads = C.int32_t(mc.NThreads)
	cPtr.n_threads_batch = C.int32_t(mc.NThreadsBatch)
	cPtr.n_batch = C.int32_t(mc.NBatch)
	cPtr.n_ubatch = C.int32_t(mc.NUbatch)
	cPtr.n_seq_max = C.int32_t(mc.NSeqMax)
	cPtr.n_gpu_layers = C.int32_t(mc.NGpuLayers)

	if mc.ChatTemplatePath != "" {
		cPtr.chat_template_path = C.CString(mc.ChatTemplatePath)
	}
	if mc.ChatTemplateContent != "" {
		cPtr.chat_template_content = C.CString(mc.ChatTemplateContent)
	}
	cPtr.enable_sampling = C.bool(mc.EnableSampling)
	if mc.GrammarStr != "" {
		cPtr.grammar_str = C.CString(mc.GrammarStr)
	}
	cPtr.max_tokens = C.int32_t(mc.MaxTokens)
	cPtr.enable_thinking = C.bool(mc.EnableThinking)
	cPtr.verbose = C.bool(mc.Verbose)

	if mc.QnnModelFolderPath != "" {
		cPtr.qnn_model_folder_path = C.CString(mc.QnnModelFolderPath)
	}
	if mc.QnnLibFolderPath != "" {
		cPtr.qnn_lib_folder_path = C.CString(mc.QnnLibFolderPath)
	}

	return cPtr
}

func freeModelConfig(cPtr *C.ml_ModelConfig) {
	if cPtr != nil {
		if cPtr.chat_template_path != nil {
			C.free(unsafe.Pointer(cPtr.chat_template_path))
		}
		if cPtr.chat_template_content != nil {
			C.free(unsafe.Pointer(cPtr.chat_template_content))
		}
		if cPtr.grammar_str != nil {
			C.free(unsafe.Pointer(cPtr.grammar_str))
		}
		if cPtr.qnn_model_folder_path != nil {
			C.free(unsafe.Pointer(cPtr.qnn_model_folder_path))
		}
		if cPtr.qnn_lib_folder_path != nil {
			C.free(unsafe.Pointer(cPtr.qnn_lib_folder_path))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

func (aci AsrCreateInput) toCPtr() *C.ml_AsrCreateInput {
	cPtr := (*C.ml_AsrCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrCreateInput{}))))
	*cPtr = C.ml_AsrCreateInput{}

	if aci.ModelName != "" {
		cPtr.model_name = C.CString(aci.ModelName)
	}
	if aci.ModelPath != "" {
		cPtr.model_path = C.CString(aci.ModelPath)
	}
	if aci.TokenizerPath != "" {
		cPtr.tokenizer_path = C.CString(aci.TokenizerPath)
	}
	cPtr.config = *aci.Config.toCPtr()
	if aci.Language != "" {
		cPtr.language = C.CString(aci.Language)
	}
	if aci.PluginID != "" {
		cPtr.plugin_id = C.CString(aci.PluginID)
	}
	if aci.DeviceID != "" {
		cPtr.device_id = C.CString(aci.DeviceID)
	}
	if aci.LicenseID != "" {
		cPtr.license_id = C.CString(aci.LicenseID)
	}
	if aci.LicenseKey != "" {
		cPtr.license_key = C.CString(aci.LicenseKey)
	}

	return cPtr
}

func freeAsrCreateInput(cPtr *C.ml_AsrCreateInput) {
	if cPtr != nil {
		if cPtr.model_name != nil {
			C.free(unsafe.Pointer(cPtr.model_name))
		}
		if cPtr.model_path != nil {
			C.free(unsafe.Pointer(cPtr.model_path))
		}
		if cPtr.tokenizer_path != nil {
			C.free(unsafe.Pointer(cPtr.tokenizer_path))
		}
		// config is a struct, not a pointer, so no need to free it
		if cPtr.language != nil {
			C.free(unsafe.Pointer(cPtr.language))
		}
		if cPtr.plugin_id != nil {
			C.free(unsafe.Pointer(cPtr.plugin_id))
		}
		if cPtr.device_id != nil {
			C.free(unsafe.Pointer(cPtr.device_id))
		}
		if cPtr.license_id != nil {
			C.free(unsafe.Pointer(cPtr.license_id))
		}
		if cPtr.license_key != nil {
			C.free(unsafe.Pointer(cPtr.license_key))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// AsrTranscribeInput represents input parameters for ASR transcription
type AsrTranscribeInput struct {
	AudioPath string
	Language  string
	Config    *ASRConfig
}

func (ati AsrTranscribeInput) toCPtr() *C.ml_AsrTranscribeInput {
	cPtr := (*C.ml_AsrTranscribeInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrTranscribeInput{}))))
	*cPtr = C.ml_AsrTranscribeInput{}

	cPtr.audio_path = C.CString(ati.AudioPath)
	if ati.Language != "" {
		cPtr.language = C.CString(ati.Language)
	} else {
		cPtr.language = nil
	}
	if ati.Config != nil {
		cPtr.config = ati.Config.toCPtr()
	} else {
		cPtr.config = nil
	}

	return cPtr
}

func freeAsrTranscribeInput(cPtr *C.ml_AsrTranscribeInput) {
	if cPtr != nil {
		if cPtr.audio_path != nil {
			C.free(unsafe.Pointer(cPtr.audio_path))
		}
		if cPtr.language != nil {
			C.free(unsafe.Pointer(cPtr.language))
		}
		if cPtr.config != nil {
			freeASRConfig(cPtr.config)
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// AsrTranscribeOutput represents output from ASR transcription
type AsrTranscribeOutput struct {
	Result      ASRResult
	ProfileData ProfileData
}

func newAsrTranscribeOutputFromCPtr(c *C.ml_AsrTranscribeOutput) AsrTranscribeOutput {
	output := AsrTranscribeOutput{}

	if c == nil {
		return output
	}

	output.Result = newASRResultFromCPtr(&c.result)
	output.ProfileData = newProfileDataFromCPtr(c.profile_data)
	return output
}

func freeAsrTranscribeOutput(ptr *C.ml_AsrTranscribeOutput) {
	if ptr == nil {
		return
	}
	freeASRResult(&ptr.result)
}

// AsrListSupportedLanguagesInput represents input for listing supported languages
type AsrListSupportedLanguagesInput struct {
	Reserved interface{}
}

func (asli AsrListSupportedLanguagesInput) toCPtr() *C.ml_AsrListSupportedLanguagesInput {
	cPtr := (*C.ml_AsrListSupportedLanguagesInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrListSupportedLanguagesInput{}))))
	*cPtr = C.ml_AsrListSupportedLanguagesInput{}
	return cPtr
}

func freeAsrListSupportedLanguagesInput(cPtr *C.ml_AsrListSupportedLanguagesInput) {
	if cPtr != nil {
		C.free(unsafe.Pointer(cPtr))
	}
}

// AsrListSupportedLanguagesOutput represents output for listing supported languages
type AsrListSupportedLanguagesOutput struct {
	LanguageCodes []string
}

func newAsrListSupportedLanguagesOutputFromCPtr(c *C.ml_AsrListSupportedLanguagesOutput) AsrListSupportedLanguagesOutput {
	output := AsrListSupportedLanguagesOutput{}

	if c == nil {
		return output
	}

	output.LanguageCodes = cCharArrayToSlice(c.language_codes, c.language_count)

	return output
}

func freeAsrListSupportedLanguagesOutput(ptr *C.ml_AsrListSupportedLanguagesOutput) {
	if ptr == nil {
		return
	}
	if ptr.language_codes != nil {
		mlFreeCCharArray(ptr.language_codes, ptr.language_count)
	}
}

// ASRTranscriptionCallback represents callback function for streaming transcription updates
type ASRTranscriptionCallback func(text string, userData interface{})

// AsrStreamBeginInput represents input for beginning ASR streaming
type AsrStreamBeginInput struct {
	StreamConfig     *ASRStreamConfig
	Language         string
	OnTranscription  ASRTranscriptionCallback
	UserData         interface{}
}

// AsrStreamBeginOutput represents output for streaming begin
type AsrStreamBeginOutput struct {
	Reserved interface{}
}

// AsrStreamPushAudioInput represents input for processing audio data
type AsrStreamPushAudioInput struct {
	AudioData []float32
	Length    int32
}

// AsrStreamStopInput represents input for stopping streaming
type AsrStreamStopInput struct {
	Graceful bool
}

// ASR represents an ASR instance
type ASR struct {
	ptr *C.ml_ASR
}

// NewASR creates a new ASR instance
func NewASR(input AsrCreateInput) (*ASR, error) {
	slog.Debug("NewASR called", "input", input)

	cInput := input.toCPtr()
	defer freeAsrCreateInput(cInput)

	var cHandle *C.ml_ASR
	res := C.ml_asr_create(cInput, &cHandle)
	if res < 0 {
		return nil, SDKError(res)
	}

	return &ASR{ptr: cHandle}, nil
}

// Destroy destroys the ASR instance and frees associated resources
func (a *ASR) Destroy() error {
	slog.Debug("Destroy called", "ptr", a.ptr)

	if a.ptr == nil {
		return nil
	}

	res := C.ml_asr_destroy(a.ptr)
	if res < 0 {
		return SDKError(res)
	}
	a.ptr = nil
	return nil
}

// Transcribe transcribes audio file to text with specified language
func (a *ASR) Transcribe(input AsrTranscribeInput) (AsrTranscribeOutput, error) {
	slog.Debug("Transcribe called", "input", input)

	cInput := input.toCPtr()
	defer freeAsrTranscribeInput(cInput)

	var cOutput C.ml_AsrTranscribeOutput
	defer freeAsrTranscribeOutput(&cOutput)

	res := C.ml_asr_transcribe(a.ptr, cInput, &cOutput)
	if res < 0 {
		return AsrTranscribeOutput{}, SDKError(res)
	}

	output := newAsrTranscribeOutputFromCPtr(&cOutput)
	return output, nil
}

// ListSupportedLanguages gets list of supported languages for ASR model
func (a *ASR) ListSupportedLanguages() (AsrListSupportedLanguagesOutput, error) {
	slog.Debug("ListSupportedLanguages called")

	input := AsrListSupportedLanguagesInput{}
	cInput := input.toCPtr()
	defer freeAsrListSupportedLanguagesInput(cInput)

	var cOutput C.ml_AsrListSupportedLanguagesOutput
	defer freeAsrListSupportedLanguagesOutput(&cOutput)

	res := C.ml_asr_list_supported_languages(a.ptr, cInput, &cOutput)
	if res < 0 {
		return AsrListSupportedLanguagesOutput{}, SDKError(res)
	}

	output := newAsrListSupportedLanguagesOutputFromCPtr(&cOutput)
	return output, nil
}

// StreamBegin begins streaming ASR with specified callbacks
func (a *ASR) StreamBegin(input AsrStreamBeginInput) (AsrStreamBeginOutput, error) {
	slog.Debug("StreamBegin called", "input", input)

	cInput := input.toCPtr()
	defer freeAsrStreamBeginInput(cInput)

	var cOutput C.ml_AsrStreamBeginOutput
	defer freeAsrStreamBeginOutput(&cOutput)

	res := C.ml_asr_stream_begin(a.ptr, cInput, &cOutput)
	if res < 0 {
		return AsrStreamBeginOutput{}, SDKError(res)
	}

	output := newAsrStreamBeginOutputFromCPtr(&cOutput)
	return output, nil
}

// StreamPushAudio pushes audio data to streaming ASR for processing
func (a *ASR) StreamPushAudio(input AsrStreamPushAudioInput) error {
	slog.Debug("StreamPushAudio called", "length", input.Length)

	cInput := input.toCPtr()
	defer freeAsrStreamPushAudioInput(cInput)

	res := C.ml_asr_stream_push_audio(a.ptr, cInput)
	if res < 0 {
		return SDKError(res)
	}

	return nil
}

// StreamStop stops streaming ASR
func (a *ASR) StreamStop(input AsrStreamStopInput) error {
	slog.Debug("StreamStop called", "graceful", input.Graceful)

	cInput := input.toCPtr()
	defer freeAsrStreamStopInput(cInput)

	res := C.ml_asr_stream_stop(a.ptr, cInput)
	if res < 0 {
		return SDKError(res)
	}

	return nil
}

// Note: C callback wrapper for ASR transcription would be implemented here
// when the C library supports proper callback mechanisms

// callbackData holds the Go callback and user data for C callbacks
type callbackData struct {
	callback ASRTranscriptionCallback
	userData interface{}
}

func (asbi AsrStreamBeginInput) toCPtr() *C.ml_AsrStreamBeginInput {
	cPtr := (*C.ml_AsrStreamBeginInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrStreamBeginInput{}))))
	*cPtr = C.ml_AsrStreamBeginInput{}
	
	if asbi.StreamConfig != nil {
		cPtr.stream_config = asbi.StreamConfig.toCPtr()
	}
	if asbi.Language != "" {
		cPtr.language = C.CString(asbi.Language)
	}
	
	// Set up callback
	if asbi.OnTranscription != nil {
		callbackData := &callbackData{
			callback: asbi.OnTranscription,
			userData: asbi.UserData,
		}
		// Note: C callback setup would need to be implemented differently
		// For now, we'll set it to nil and handle callbacks in Go
		cPtr.on_transcription = nil
		cPtr.user_data = unsafe.Pointer(callbackData)
	}
	
	return cPtr
}

func freeAsrStreamBeginInput(cPtr *C.ml_AsrStreamBeginInput) {
	if cPtr != nil {
		if cPtr.stream_config != nil {
			freeASRStreamConfig(cPtr.stream_config)
		}
		if cPtr.language != nil {
			C.free(unsafe.Pointer(cPtr.language))
		}
		if cPtr.user_data != nil {
			// Free the callback data
			callbackData := (*callbackData)(cPtr.user_data)
			C.free(unsafe.Pointer(callbackData))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

func newAsrStreamBeginOutputFromCPtr(c *C.ml_AsrStreamBeginOutput) AsrStreamBeginOutput {
	output := AsrStreamBeginOutput{}
	if c == nil {
		return output
	}
	// Reserved field is empty for now
	return output
}

func freeAsrStreamBeginOutput(ptr *C.ml_AsrStreamBeginOutput) {
	// Nothing to free for now as it's reserved
}

func (aspai AsrStreamPushAudioInput) toCPtr() *C.ml_AsrStreamPushAudioInput {
	cPtr := (*C.ml_AsrStreamPushAudioInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrStreamPushAudioInput{}))))
	*cPtr = C.ml_AsrStreamPushAudioInput{}
	
	if len(aspai.AudioData) > 0 {
		cPtr.audio_data = (*C.float)(unsafe.Pointer(&aspai.AudioData[0]))
	}
	cPtr.length = C.int32_t(aspai.Length)
	
	return cPtr
}

func freeAsrStreamPushAudioInput(cPtr *C.ml_AsrStreamPushAudioInput) {
	if cPtr != nil {
		C.free(unsafe.Pointer(cPtr))
	}
}

func (assi AsrStreamStopInput) toCPtr() *C.ml_AsrStreamStopInput {
	cPtr := (*C.ml_AsrStreamStopInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_AsrStreamStopInput{}))))
	*cPtr = C.ml_AsrStreamStopInput{}
	
	cPtr.graceful = C.bool(assi.Graceful)
	
	return cPtr
}

func freeAsrStreamStopInput(cPtr *C.ml_AsrStreamStopInput) {
	if cPtr != nil {
		C.free(unsafe.Pointer(cPtr))
	}
}
