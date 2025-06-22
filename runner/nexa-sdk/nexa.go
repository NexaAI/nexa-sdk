package nexa_sdk

/*
#cgo CFLAGS: -I../../build/include
#cgo LDFLAGS: -L../../build/lib -lnexa_bridge
#cgo linux LDFLAGS: -Wl,--unresolved-symbols=ignore-in-shared-libs

#include <stdlib.h>
#include "ml.h"

extern void go_log_wrap(char *msg);
*/
import "C"
import (
	"errors"
	"fmt"
)

var (
	// ErrCommon represents a generic error from the C library
	ErrCommon = errors.New("error")
)

// Init initializes the Nexa SDK by calling the underlying C library initialization
// This must be called before using any other SDK functions
func Init() {
	C.ml_init()
}

// DeInit cleans up resources allocated by the Nexa SDK
// This should be called when the SDK is no longer needed
func DeInit() {
	C.ml_deinit()
}

// go_log_wrap is exported to C and handles log messages from the C library
// It converts C strings to Go strings and prints them to stdout
//export go_log_wrap
func go_log_wrap(msg *C.char) {
	fmt.Println(C.GoString(msg))
}
