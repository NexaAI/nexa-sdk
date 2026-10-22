// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGrammarForRun(t *testing.T) {
	const grammar = `root ::= "YES"`
	dir := t.TempDir()
	path := filepath.Join(dir, "grammar.gbnf")
	missing := filepath.Join(dir, "missing.gbnf")
	if err := os.WriteFile(path, []byte(grammar), 0600); err != nil {
		t.Fatal(err)
	}

	for _, tt := range []struct {
		name, path, inline, want string
		wantErr                  bool
	}{
		{name: "no grammar"},
		{name: "inline grammar", inline: grammar, want: grammar},
		{name: "local file", path: path, want: grammar},
		{name: "inline overrides file", path: missing, inline: grammar, want: grammar},
		{name: "missing file", path: missing, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := grammarForRun(tt.path, tt.inline)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("grammarForRun(%q, %q) = (%q, %v), want (%q, error=%v)",
					tt.path, tt.inline, got, err, tt.want, tt.wantErr)
			}
		})
	}
}
