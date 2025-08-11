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

// RerankConfig represents reranking configuration
type RerankConfig struct {
	BatchSize       int32
	Normalize       bool
	NormalizeMethod string
}

func (rc RerankConfig) toCPtr() *C.ml_RerankConfig {
	cPtr := (*C.ml_RerankConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_RerankConfig{}))))
	*cPtr = C.ml_RerankConfig{}

	cPtr.batch_size = C.int32_t(rc.BatchSize)
	cPtr.normalize = C.bool(rc.Normalize)
	if rc.NormalizeMethod != "" {
		cPtr.normalize_method = C.CString(rc.NormalizeMethod)
	}

	return cPtr
}

func freeRerankConfig(cPtr *C.ml_RerankConfig) {
	if cPtr != nil {
		if cPtr.normalize_method != nil {
			C.free(unsafe.Pointer(cPtr.normalize_method))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// RerankerCreateInput represents input parameters for reranker creation
type RerankerCreateInput struct {
	ModelPath     string
	TokenizerPath string
	PluginID      string
}

func (rci RerankerCreateInput) toCPtr() *C.ml_RerankerCreateInput {
	cPtr := (*C.ml_RerankerCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_RerankerCreateInput{}))))
	*cPtr = C.ml_RerankerCreateInput{}

	if rci.ModelPath != "" {
		cPtr.model_path = C.CString(rci.ModelPath)
	}
	if rci.TokenizerPath != "" {
		cPtr.tokenizer_path = C.CString(rci.TokenizerPath)
	}
	if rci.PluginID != "" {
		cPtr.plugin_id = C.CString(rci.PluginID)
	}

	return cPtr
}

func freeRerankerCreateInput(cPtr *C.ml_RerankerCreateInput) {
	if cPtr != nil {
		if cPtr.model_path != nil {
			C.free(unsafe.Pointer(cPtr.model_path))
		}
		if cPtr.tokenizer_path != nil {
			C.free(unsafe.Pointer(cPtr.tokenizer_path))
		}
		if cPtr.plugin_id != nil {
			C.free(unsafe.Pointer(cPtr.plugin_id))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// RerankerRerankInput represents input parameters for reranking operation
type RerankerRerankInput struct {
	Query     string
	Documents []string
	Config    *RerankConfig
}

func (rri RerankerRerankInput) toCPtr() *C.ml_RerankerRerankInput {
	cPtr := (*C.ml_RerankerRerankInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_RerankerRerankInput{}))))
	*cPtr = C.ml_RerankerRerankInput{}

	if rri.Query != "" {
		cPtr.query = C.CString(rri.Query)
	}

	// Convert documents array
	if len(rri.Documents) > 0 {
		documentsArray, documentCount := sliceToCCharArray(rri.Documents)
		cPtr.documents = documentsArray
		cPtr.documents_count = documentCount
	} else {
		cPtr.documents = nil
		cPtr.documents_count = 0
	}

	// Set config
	if rri.Config != nil {
		cPtr.config = rri.Config.toCPtr()
	} else {
		cPtr.config = nil
	}

	return cPtr
}

func freeRerankerRerankInput(cPtr *C.ml_RerankerRerankInput) {
	if cPtr == nil {
		return
	}

	if cPtr.query != nil {
		C.free(unsafe.Pointer(cPtr.query))
	}

	if cPtr.documents != nil {
		freeCCharArray(cPtr.documents, cPtr.documents_count)
	}

	if cPtr.config != nil {
		freeRerankConfig(cPtr.config)
	}

	C.free(unsafe.Pointer(cPtr))
}

// RerankerRerankOutput represents output from reranking operation
type RerankerRerankOutput struct {
	Scores      []float32
	ScoreCount  int32
	ProfileData ProfileData
}

func newRerankerRerankOutputFromCPtr(c *C.ml_RerankerRerankOutput) RerankerRerankOutput {
	output := RerankerRerankOutput{}

	if c == nil {
		return output
	}

	// Convert scores array
	if c.scores != nil && c.score_count > 0 {
		scores := unsafe.Slice((*C.float)(unsafe.Pointer(c.scores)), int(c.score_count))
		output.Scores = make([]float32, c.score_count)
		for i := range output.Scores {
			output.Scores[i] = float32(scores[i])
		}
	}

	output.ScoreCount = int32(c.score_count)
	output.ProfileData = newProfileDataFromCPtr(c.profile_data)

	return output
}

func freeRerankerRerankOutput(ptr *C.ml_RerankerRerankOutput) {
	if ptr == nil {
		return
	}
	if ptr.scores != nil {
		mlFree(unsafe.Pointer(ptr.scores))
	}
}

// Reranker represents a reranker instance
type Reranker struct {
	ptr *C.ml_Reranker
}

// NewReranker creates a new reranker instance
func NewReranker(input RerankerCreateInput) (*Reranker, error) {
	slog.Debug("NewReranker called", "input", input)

	cInput := input.toCPtr()
	defer freeRerankerCreateInput(cInput)

	var cHandle *C.ml_Reranker
	res := C.ml_reranker_create(cInput, &cHandle)
	if res < 0 {
		return nil, SDKError(res)
	}

	return &Reranker{ptr: cHandle}, nil
}

// Destroy destroys the reranker instance and frees associated resources
func (r *Reranker) Destroy() error {
	slog.Debug("Destroy called", "ptr", r.ptr)

	if r.ptr == nil {
		return nil
	}

	res := C.ml_reranker_destroy(r.ptr)
	if res < 0 {
		return SDKError(res)
	}
	r.ptr = nil
	return nil
}

// Rerank performs reranking operation on documents against a query
func (r *Reranker) Rerank(input RerankerRerankInput) (RerankerRerankOutput, error) {
	slog.Debug("Rerank called", "input", input)

	cInput := input.toCPtr()
	defer freeRerankerRerankInput(cInput)

	var cOutput C.ml_RerankerRerankOutput
	defer freeRerankerRerankOutput(&cOutput)

	res := C.ml_reranker_rerank(r.ptr, cInput, &cOutput)
	if res < 0 {
		return RerankerRerankOutput{}, SDKError(res)
	}

	output := newRerankerRerankOutputFromCPtr(&cOutput)
	return output, nil
}
