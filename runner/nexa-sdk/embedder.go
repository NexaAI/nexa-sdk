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

// EmbeddingConfig represents embedding generation configuration
type EmbeddingConfig struct {
	BatchSize       int32
	Normalize       bool
	NormalizeMethod string
}

func (ec EmbeddingConfig) toCPtr() *C.ml_EmbeddingConfig {
	cPtr := (*C.ml_EmbeddingConfig)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_EmbeddingConfig{}))))
	*cPtr = C.ml_EmbeddingConfig{}

	cPtr.batch_size = C.int32_t(ec.BatchSize)
	cPtr.normalize = C.bool(ec.Normalize)
	if ec.NormalizeMethod != "" {
		cPtr.normalize_method = C.CString(ec.NormalizeMethod)
	}

	return cPtr
}

func freeEmbeddingConfig(cPtr *C.ml_EmbeddingConfig) {
	if cPtr != nil {
		if cPtr.normalize_method != nil {
			C.free(unsafe.Pointer(cPtr.normalize_method))
		}
		C.free(unsafe.Pointer(cPtr))
	}
}

// EmbedderCreateInput represents input parameters for embedder creation
type EmbedderCreateInput struct {
	ModelPath     string
	TokenizerPath string
	PluginID      string
}

func (eci EmbedderCreateInput) toCPtr() *C.ml_EmbedderCreateInput {
	cPtr := (*C.ml_EmbedderCreateInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_EmbedderCreateInput{}))))
	*cPtr = C.ml_EmbedderCreateInput{}

	if eci.ModelPath != "" {
		cPtr.model_path = C.CString(eci.ModelPath)
	}
	if eci.TokenizerPath != "" {
		cPtr.tokenizer_path = C.CString(eci.TokenizerPath)
	}
	if eci.PluginID != "" {
		cPtr.plugin_id = C.CString(eci.PluginID)
	}

	return cPtr
}

func freeEmbedderCreateInput(cPtr *C.ml_EmbedderCreateInput) {
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

// EmbedderEmbedInput represents input parameters for embedding generation
type EmbedderEmbedInput struct {
	Texts              []string
	Config             *EmbeddingConfig
	InputIDs2D         [][]int32
	InputIDsRowLengths []int32
}

func (eei EmbedderEmbedInput) toCPtr() *C.ml_EmbedderEmbedInput {
	cPtr := (*C.ml_EmbedderEmbedInput)(C.malloc(C.size_t(unsafe.Sizeof(C.ml_EmbedderEmbedInput{}))))
	*cPtr = C.ml_EmbedderEmbedInput{}

	// Convert texts array
	if len(eei.Texts) > 0 {
		textsArray, textCount := sliceToCCharArray(eei.Texts)
		cPtr.texts = textsArray
		cPtr.text_count = textCount
	} else {
		cPtr.texts = nil
		cPtr.text_count = 0
	}

	// Set config
	if eei.Config != nil {
		cPtr.config = eei.Config.toCPtr()
	} else {
		cPtr.config = nil
	}

	// Convert input_ids_2d array (if provided)
	if len(eei.InputIDs2D) > 0 {
		// Allocate array of pointers
		raw := C.malloc(C.size_t(len(eei.InputIDs2D)) * C.size_t(unsafe.Sizeof(uintptr(0))))
		inputIDs2D := unsafe.Slice((**C.int32_t)(raw), len(eei.InputIDs2D))

		// Allocate each row
		for i, row := range eei.InputIDs2D {
			if len(row) > 0 {
				rowPtr := C.malloc(C.size_t(len(row)) * C.size_t(unsafe.Sizeof(C.int32_t(0))))
				cRow := unsafe.Slice((*C.int32_t)(rowPtr), len(row))
				for j, val := range row {
					cRow[j] = C.int32_t(val)
				}
				inputIDs2D[i] = (*C.int32_t)(rowPtr)
			} else {
				inputIDs2D[i] = nil
			}
		}

		cPtr.input_ids_2d = (**C.int32_t)(raw)
		cPtr.input_ids_row_count = C.int32_t(len(eei.InputIDs2D))

		// Convert row lengths
		if len(eei.InputIDsRowLengths) > 0 {
			rowLengthsPtr := C.malloc(C.size_t(len(eei.InputIDsRowLengths)) * C.size_t(unsafe.Sizeof(C.int32_t(0))))
			cRowLengths := unsafe.Slice((*C.int32_t)(rowLengthsPtr), len(eei.InputIDsRowLengths))
			for i, length := range eei.InputIDsRowLengths {
				cRowLengths[i] = C.int32_t(length)
			}
			cPtr.input_ids_row_lengths = (*C.int32_t)(rowLengthsPtr)
		} else {
			cPtr.input_ids_row_lengths = nil
		}
	} else {
		cPtr.input_ids_2d = nil
		cPtr.input_ids_row_lengths = nil
		cPtr.input_ids_row_count = 0
	}

	return cPtr
}

func freeEmbedderEmbedInput(cPtr *C.ml_EmbedderEmbedInput) {
	if cPtr == nil {
		return
	}

	// Free texts array
	if cPtr.texts != nil {
		freeCCharArray(cPtr.texts, cPtr.text_count)
	}

	// Free config
	if cPtr.config != nil {
		freeEmbeddingConfig(cPtr.config)
	}

	// Free input_ids_2d array
	if cPtr.input_ids_2d != nil && cPtr.input_ids_row_count > 0 {
		inputIDs2D := unsafe.Slice(cPtr.input_ids_2d, int(cPtr.input_ids_row_count))
		for _, rowPtr := range inputIDs2D {
			if rowPtr != nil {
				C.free(unsafe.Pointer(rowPtr))
			}
		}
		C.free(unsafe.Pointer(cPtr.input_ids_2d))
	}

	// Free input_ids_row_lengths
	if cPtr.input_ids_row_lengths != nil {
		C.free(unsafe.Pointer(cPtr.input_ids_row_lengths))
	}

	C.free(unsafe.Pointer(cPtr))
}

// EmbedderEmbedOutput represents output from embedding generation
type EmbedderEmbedOutput struct {
	Embeddings  []float32
	ProfileData ProfileData
}

func newEmbedderEmbedOutputFromCPtr(c *C.ml_EmbedderEmbedOutput, embeddingDim int32) EmbedderEmbedOutput {
	output := EmbedderEmbedOutput{}

	if c == nil {
		return output
	}

	output.ProfileData = newProfileDataFromCPtr(c.profile_data)

	// Convert embeddings array
	// c.embedding_count = number of embeddings (texts)
	// c.embeddings = flat array of embedding_count * embedding_dimension floats
	if c.embeddings != nil && c.embedding_count > 0 {
		totalFloats := int(c.embedding_count * C.int32_t(embeddingDim))
		embeddings := unsafe.Slice((*C.float)(unsafe.Pointer(c.embeddings)), totalFloats)
		output.Embeddings = make([]float32, totalFloats)
		for i := range output.Embeddings {
			output.Embeddings[i] = float32(embeddings[i])
		}
	}

	return output
}

func freeEmbedderEmbedOutput(ptr *C.ml_EmbedderEmbedOutput) {
	if ptr == nil {
		return
	}
	if ptr.embeddings != nil {
		mlFree(unsafe.Pointer(ptr.embeddings))
	}
}

// EmbedderDimOutput represents output for getting embedding dimension
type EmbedderDimOutput struct {
	Dimension int32
}

func newEmbedderDimOutputFromCPtr(c *C.ml_EmbedderDimOutput) EmbedderDimOutput {
	output := EmbedderDimOutput{}

	if c == nil {
		return output
	}

	output.Dimension = int32(c.dimension)
	return output
}

// Embedder represents an embedder instance
type Embedder struct {
	ptr *C.ml_Embedder
}

// NewEmbedder creates a new embedder instance
func NewEmbedder(input EmbedderCreateInput) (*Embedder, error) {
	slog.Debug("NewEmbedder called", "input", input)

	cInput := input.toCPtr()
	defer freeEmbedderCreateInput(cInput)

	var cHandle *C.ml_Embedder
	res := C.ml_embedder_create(cInput, &cHandle)
	if res < 0 {
		return nil, SDKError(res)
	}

	return &Embedder{ptr: cHandle}, nil
}

// Destroy destroys the embedder instance and frees associated resources
func (e *Embedder) Destroy() error {
	slog.Debug("Destroy called", "ptr", e.ptr)

	if e.ptr == nil {
		return nil
	}

	res := C.ml_embedder_destroy(e.ptr)
	if res < 0 {
		return SDKError(res)
	}
	e.ptr = nil
	return nil
}

// Embed generates embeddings for input texts
func (e *Embedder) Embed(input EmbedderEmbedInput) (EmbedderEmbedOutput, error) {
	slog.Debug("Embed called", "input", input)

	// First get embedding dimension
	dimOutput, err := e.EmbeddingDimension()
	if err != nil {
		return EmbedderEmbedOutput{}, err
	}

	cInput := input.toCPtr()
	defer freeEmbedderEmbedInput(cInput)

	var cOutput C.ml_EmbedderEmbedOutput
	defer freeEmbedderEmbedOutput(&cOutput)

	res := C.ml_embedder_embed(e.ptr, cInput, &cOutput)
	if res < 0 {
		return EmbedderEmbedOutput{}, SDKError(res)
	}

	output := newEmbedderEmbedOutputFromCPtr(&cOutput, dimOutput.Dimension)
	return output, nil
}

// EmbeddingDimension gets embedding dimension from the model
func (e *Embedder) EmbeddingDimension() (EmbedderDimOutput, error) {
	slog.Debug("EmbeddingDimension called")

	var cOutput C.ml_EmbedderDimOutput

	res := C.ml_embedder_embedding_dim(e.ptr, &cOutput)
	if res < 0 {
		return EmbedderDimOutput{}, SDKError(res)
	}

	output := newEmbedderDimOutputFromCPtr(&cOutput)
	return output, nil
}
