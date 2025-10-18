// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package render

import (
	"fmt"
	"testing"
)

func TestThemeColor(t *testing.T) {
	theme := GetTheme()
	fmt.Println(theme.Normal.Sprint("This is a normal message."))
	fmt.Println(theme.Info.Sprint("This is an info message."))
	fmt.Println(theme.Success.Sprint("This is a success message."))
	fmt.Println(theme.Warning.Sprint("This is a warning message."))
	fmt.Println(theme.Error.Sprint("This is an error message."))

	fmt.Println(theme.Quant.Sprint("🔹 Quant=Q4_K_M"))
	fmt.Println(theme.Prompt.Sprint("> ") + "/path/to/image how are you")
	fmt.Println(theme.AddFiles.Sprint("add image: /path/to/image"))
	fmt.Println(theme.ThinkOutput.Sprint("<think>this is think message</think>"))
	fmt.Println(theme.ModelOutput.Sprint("hi how are you today"))
	fmt.Println(theme.Profile.Sprint("— 64.8 tok/s • 61 tok • 18.4 s first token —"))
}
