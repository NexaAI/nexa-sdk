// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package geniex_sdk

// This file deliberately does not `import "C"` — cmd/go forbids cgo in
// in-package _test.go files. C interop goes through wrappers (cCString,
// cGoString, cFreeIfSet) defined in helpers.go.

import (
	"runtime/cgo"
	"testing"
	"unsafe"
)

// Encodes the "empty Go string → NULL C pointer" invariant every binding relies on.
func TestCStringIfSet(t *testing.T) {
	if got := cStringIfSet(""); got != nil {
		t.Errorf("empty string: want nil, got %p", got)
		cFreeIfSet(unsafe.Pointer(got))
	}
	got := cStringIfSet("hello")
	if got == nil {
		t.Fatal("non-empty string returned nil")
	}
	defer cFreeIfSet(unsafe.Pointer(got))
	if s := cGoString(got); s != "hello" {
		t.Errorf("roundtrip mismatch: got %q, want %q", s, "hello")
	}
}

// NULL user_data must keep generation alive; an installed callback's return
// value must propagate so the user can stop.
func TestOnTokenCallbackDispatch(t *testing.T) {
	tok := cCString("abc")
	defer cFreeIfSet(unsafe.Pointer(tok))

	if got := bool(go_generate_stream_on_token(tok, nil)); !got {
		t.Error("nil user_data: expected true to keep SDK generating")
	}

	var seen string
	cb := OnTokenCallback(func(s string) bool {
		seen = s
		return false
	})
	h := cgo.NewHandle(cb)
	defer h.Delete()

	got := bool(go_generate_stream_on_token(tok, handleToUserData(h)))
	if got {
		t.Error("installed cb returning false should propagate to SDK")
	}
	if seen != "abc" {
		t.Errorf("callback payload: got %q, want %q", seen, "abc")
	}
}
