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
