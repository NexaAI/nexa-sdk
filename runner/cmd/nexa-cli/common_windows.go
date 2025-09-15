package main

import (
	"os"

	"golang.org/x/sys/windows"
)

func getTerminalWidth() int {
	handle := windows.Handle(os.Stdout.Fd())
	var info windows.ConsoleScreenBufferInfo
	err := windows.GetConsoleScreenBufferInfo(handle, &info)
	if err != nil {
		return 80
	}
	width := int(info.Window.Right - info.Window.Left + 1)
	return width
}
