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
	"log/slog"
	"strings"
)

type SDKError int32

func (s SDKError) Error() string {
	return fmt.Sprintf("SDKError(%s)", C.GoString(C.ml_get_error_message(C.ml_ErrorCode(s))))
}

const (
	SDKErrorUnknown   = SDKError(C.ML_ERROR_COMMON_UNKNOWN)
	SDKErrorModelLoad = SDKError(C.ML_ERROR_COMMON_MODEL_LOAD)
)

// ProfilingData contains performance metrics from ML operations
type ProfilingData struct {
	TTFTUs          int64   // Time to first token (us)
	TotalTimeUs     int64   // Total generation time (us)
	PromptTimeUs    int64   // Prompt processing time (us)
	DecodeTimeUs    int64   // Token generation time (us)
	TokensPerSecond float64 // Decoding speed (tokens/sec)
	TotalTokens     int64   // Total tokens generated
	PromptTokens    int64   // Number of prompt tokens
	GeneratedTokens int64   // Number of generated tokens
	StopReason      string  // Stop reason: "eos", "length", "user", "stop_sequence"
}

// NewProfilingDataFromC creates a new ProfilingData from C ml_ProfilingData
func NewProfilingDataFromC(cData C.ml_ProfilingData) *ProfilingData {
	return &ProfilingData{
		TTFTUs:          int64(cData.ttft_us),
		TotalTimeUs:     int64(cData.total_time_us),
		PromptTimeUs:    int64(cData.prompt_time_us),
		DecodeTimeUs:    int64(cData.decode_time_us),
		TokensPerSecond: float64(cData.tokens_per_second),
		TotalTokens:     int64(cData.total_tokens),
		PromptTokens:    int64(cData.prompt_tokens),
		GeneratedTokens: int64(cData.generated_tokens),
		StopReason:      C.GoString(cData.stop_reason),
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

// Get SDK Version
func Version() string {
	return C.GoString(C.ml_version())
}

var bridgeLogEnabled = false

// go_log_wrap is exported to C and handles log messages from the C library
// It converts C strings to Go strings and prints them to stdout
//
//export go_log_wrap
func go_log_wrap(msg *C.char) {
	if bridgeLogEnabled {
		slog.Debug("[Bridge] " + strings.TrimSpace(C.GoString(msg)))
	}
}

// EnableBridgeLog enables or disables the bridge log
func EnableBridgeLog(enable bool) {
	bridgeLogEnabled = enable
}
