// Copyright 2024-2026 Qualcomm Technologies, Inc. and/or its subsidiaries.
// SPDX-License-Identifier: BSD-3-Clause

package decision

import (
	"testing"

	geniex_sdk "github.com/qualcomm/GenieX/bindings/go"
)

type fakeVLM struct {
	templateInput geniex_sdk.VlmApplyChatTemplateInput
	generateInput geniex_sdk.VlmGenerateInput
	response      *geniex_sdk.VlmGenerateOutput
	generateErr   error
}

func (f *fakeVLM) ApplyChatTemplate(input geniex_sdk.VlmApplyChatTemplateInput) (*geniex_sdk.VlmApplyChatTemplateOutput, error) {
	f.templateInput = input
	return &geniex_sdk.VlmApplyChatTemplateOutput{FormattedText: "formatted"}, nil
}

func (f *fakeVLM) Generate(input geniex_sdk.VlmGenerateInput) (*geniex_sdk.VlmGenerateOutput, error) {
	f.generateInput = input
	return f.response, f.generateErr
}

func TestGenerationOptionsCopiesCallerStopSlice(t *testing.T) {
	stop := []string{"</s>"}
	config := (GenerationOptions{Stop: stop}).generationConfig(0, "grammar", geniex_sdk.RuntimeLlamaCpp, nil)
	stop[0] = "mutated"
	if config.MaxTokens != 128 || config.Stop[0] != "</s>" {
		t.Fatalf("config = %#v", config)
	}
}
