package main

import (
	"testing"
)

//  cd runner && LD_LIBRARY_PATH=$PWD/../build/lib go test ./cmd/nexa-cli -v -run TestQuantRegix_MatchMixedCaseQuantLevels

func TestQuantRegix_MatchAllQuantLevels(t *testing.T) {
	// Reference: https://huggingface.co/docs/hub/en/gguf?#quantization-types
	// Test all case combinations: uppercase, lowercase, and mixed case
	quantLevels := []string{
		// Uppercase
		"F64", "I64", "F32", "I32", "F16", "BF16", "I16", "Q8_0", "Q8_1", "Q8_K",
		"I8", "Q6_K", "Q5_0", "Q5_1", "Q5_K", "Q4_0", "Q4_1", "Q4_K", "Q3_K", "Q2_K",
		"IQ4_NL", "IQ4_XS", "IQ3_S", "IQ3_XXS", "IQ2_XXS", "IQ2_S", "IQ2_XS", "IQ1_S", "IQ1_M",
		// Lowercase
		"f64", "i64", "f32", "i32", "f16", "bf16", "i16", "q8_0", "q8_1", "q8_k",
		"i8", "q6_k", "q5_0", "q5_1", "q5_k", "q4_0", "q4_1", "q4_k", "q3_k", "q2_k",
		"iq4_nl", "iq4_xs", "iq3_s", "iq3_xxs", "iq2_xxs", "iq2_s", "iq2_xs", "iq1_s", "iq1_m",
		// Mixed case
		"f64", "I64", "F32", "i32", "f16", "Bf16", "I16", "q8_0", "Q8_1", "q8_K",
		"i8", "Q6_k", "q5_0", "Q5_1", "q5_K", "Q4_0", "q4_1", "Q4_k", "q3_K", "Q2_k",
		"Iq4_nl", "iQ4_xs", "Iq3_s", "iQ3_xxs", "Iq2_xxs", "iQ2_s", "Iq2_xs", "iQ1_s", "Iq1_m",
	}

	for _, level := range quantLevels {
		matched := quantRegix.FindString(level)
		if matched == "" {
			t.Errorf("quantRegix did not match: %s", level)
		}
	}
}
