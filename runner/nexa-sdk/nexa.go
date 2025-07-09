package nexa_sdk

/*
#cgo CFLAGS: -I./stub
#cgo LDFLAGS: -L./stub -lnexa_bridge
#cgo linux LDFLAGS: -Wl,-rpath,${SRCDIR}/../../build/lib
#cgo darwin LDFLAGS: -Wl,-rpath,${SRCDIR}/../../build/lib

#include <stdlib.h>
#include "ml.h"

extern void go_log_wrap(char *msg);
*/
import "C"

import (
	"fmt"
	"log"
	"strings"
)

type SDKError int32

func (s SDKError) Error() string {
	return fmt.Sprintf("SDKError(%s)", C.GoString(C.ml_get_error_message(C.ml_ErrorCode(s))))
}

const (
	SDKErrorUnknown      = SDKError(C.ML_ERROR_COMMON_UNKNOWN)
	SDKErrorModelLoad    = SDKError(C.ML_ERROR_COMMON_MODEL_LOAD)
	SDKErrorFileNotFound = SDKError(C.ML_ERROR_COMMON_FILE_NOT_FOUND)
)

// ProfilingData contains performance metrics from ML operations
type ProfilingData struct {
	TTFTMs          float64 // Time to first token (ms)
	TotalTokens     int32   // Total tokens generated
	StopReason      string  // Stop reason: "eos", "length", "user", "stop_sequence"
	TokensPerSecond float64 // Decoding speed (tokens/sec)
	TotalTimeMs     float64 // Total generation time
	PromptTimeMs    float64 // Prompt processing time
	DecodeTimeMs    float64 // Token generation time
	PromptTokens    int32   // Number of prompt tokens
	GeneratedTokens int32   // Number of generated tokens
}

// NewProfilingDataFromC creates a new ProfilingData from C ml_ProfilingData
func NewProfilingDataFromC(cData C.ml_ProfilingData) *ProfilingData {
	return &ProfilingData{
		TTFTMs:          float64(cData.ttft_ms),
		TotalTokens:     int32(cData.total_tokens),
		StopReason:      C.GoString(cData.stop_reason),
		TokensPerSecond: float64(cData.tokens_per_second),
		TotalTimeMs:     float64(cData.total_time_ms),
		PromptTimeMs:    float64(cData.prompt_time_ms),
		DecodeTimeMs:    float64(cData.decode_time_ms),
		PromptTokens:    int32(cData.prompt_tokens),
		GeneratedTokens: int32(cData.generated_tokens),
	}
}

// Init initializes the Nexa SDK by calling the underlying C library initialization
// This must be called before using any other SDK functions
func Init() {
	C.ml_init()
	C.ml_set_log((C.ml_log_callback)(C.go_log_wrap))
}

// DeInit cleans up resources allocated by the Nexa SDK
// This should be called when the SDK is no longer needed
func DeInit() {
	C.ml_deinit()
}

// go_log_wrap is exported to C and handles log messages from the C library
// It converts C strings to Go strings and prints them to stdout
//
//export go_log_wrap
func go_log_wrap(msg *C.char) {
	log.Default().Println(strings.TrimSpace(C.GoString(msg)))
}
