package common

import (
	"golang.org/x/sys/unix"
)

func GetTerminalWidth() int {
	fd := int(unix.Stdout)
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 80
	}
	return int(ws.Col)
}
