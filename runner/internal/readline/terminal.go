package readline

import (
	"os"
	"syscall"
)

type Terminal struct {
	oldTermios *syscall.Termios
}

func NewTerminal() (*Terminal, error) {
	t := &Terminal{}

	termios, err := getTermios(os.Stdin.Fd())
	if err != nil {
		return nil, err
	}

	t.oldTermios = termios

	newTermios := *termios
	newTermios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	newTermios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newTermios.Cflag &^= syscall.CSIZE | syscall.PARENB
	newTermios.Cflag |= syscall.CS8
	newTermios.Cc[syscall.VMIN] = 1
	newTermios.Cc[syscall.VTIME] = 0

	setTermios(os.Stdin.Fd(), &newTermios)

	return t, nil
}

func (t Terminal) Close() error {
	return setTermios(os.Stdin.Fd(), t.oldTermios)
}
