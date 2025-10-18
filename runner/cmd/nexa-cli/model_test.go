// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package main

import "testing"

func TestQuantRegix_MatchAllQuantLevels(t *testing.T) {
	// Test all quantization levels based on the provided list
	quantLevels := []string{
		// Uppercase
		"FP32", "FP16", "FP64",
		"F64", "F32", "F16",
		"I64", "I32", "I16", "I8",
		"BF16",
		"Q8_0", "Q8_1", "Q8_K", "Q6_K", "Q5_0", "Q5_1", "Q5_K", "Q4_0", "Q4_1", "Q4_K", "Q3_K", "Q2_K",
		"IQ4_NL", "IQ4_XS", "IQ3_S", "IQ3_XXS", "IQ2_XXS", "IQ2_S", "IQ2_XS", "IQ1_S", "IQ1_M",
		"1bit", "2bit", "3bit", "4bit", "16bit",

		// Lowercase versions
		"fp32", "fp16", "fp64",
		"f64", "f32", "f16",
		"i64", "i32", "i16", "i8",
		"bf16",
		"q8_0", "q8_1", "q8_k", "q6_k", "q5_0", "q5_1", "q5_k", "q4_0", "q4_1", "q4_k", "q3_k", "q2_k",
		"iq4_nl", "iq4_xs", "iq3_s", "iq3_xxs", "iq2_xxs", "iq2_s", "iq2_xs", "iq1_s", "iq1_m",
		"1bit", "2bit", "3bit", "4bit", "16bit",

		// Mixed case versions
		"Fp32", "fP16", "Fp64",
		"F64", "f32", "F16",
		"I64", "i32", "I16", "I8",
		"Bf16", "bF16",
		"Q8_0", "q8_1", "Q8_k", "q6_K", "Q5_0", "q5_1", "Q5_k", "Q4_0", "q4_1", "Q4_k", "q3_K", "Q2_k",
		"Iq4_nl", "iQ4_xs", "Iq3_s", "iQ3_xxs", "Iq2_xxs", "iQ2_s", "Iq2_xs", "iQ1_s", "Iq1_m",
		"1BIT", "2BIT", "3BIT", "4BIT", "16BIT",
		"1Bit", "2Bit", "3Bit", "4Bit", "16Bit",
		"1bIt", "2bIt", "3bIt", "4bIt", "16bIt",
	}

	for _, level := range quantLevels {
		matched := quantRegix.FindString(level)
		if matched != level {
			t.Errorf("quantRegix did not match: %s, %s", level, matched)
		}
	}
}
