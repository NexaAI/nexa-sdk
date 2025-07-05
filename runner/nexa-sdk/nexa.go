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
	"errors"
	"log"
	"strings"
)

var (
	// ErrCommon represents a generic error from the C library
	ErrCommon               = errors.New("SDK Error")
	ErrCreateFailed         = errors.New("Model Create Failed")
	ErrChatTemplateNotFound = errors.New("Chat Template Not Found")
)

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
