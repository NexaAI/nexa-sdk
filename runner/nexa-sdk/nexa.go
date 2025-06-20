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
	ErrCommon = errors.New("error")
)

func Init() {
	C.ml_init()
}

func DeInit() {
	C.ml_deinit()
}

//export go_log_wrap
func go_log_wrap(msg *C.char) {
	fmt.Println(C.GoString(msg))
}
