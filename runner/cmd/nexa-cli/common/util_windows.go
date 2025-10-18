// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package common

import (
	"os"

	"golang.org/x/sys/windows"
)

func GetTerminalWidth() int {
	handle := windows.Handle(os.Stdout.Fd())
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(handle, &info)
	if err != nil {
		return 80
	}
	width := int(info.Window.Right - info.Window.Left + 1)
	return width
}
