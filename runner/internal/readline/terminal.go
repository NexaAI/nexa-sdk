//go:build !windows

package readline

import (
	"bufio"
	"os"
	"syscall"
)

type Terminal struct {
	oldTermios syscall.Termios
	termios    syscall.Termios
	r          *bufio.Reader
}

func NewTerminal() (*Terminal, error) {
	t := &Terminal{}

	termios, err := getTermios()
	if err != nil {
		return nil, err
	}

	t.oldTermios = *termios

	t.termios = *termios
	t.termios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON
	t.termios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	t.termios.Cflag &^= syscall.CSIZE | syscall.PARENB
	t.termios.Cflag |= syscall.CS8
	t.termios.Cc[syscall.VMIN] = 1
	t.termios.Cc[syscall.VTIME] = 0

	t.r = bufio.NewReader(os.Stdin)

	return t, nil
}

func (t *Terminal) Read() (rune, error) {
	r, _, err := t.r.ReadRune()
	return r, err
}

func (t *Terminal) Close() error {
	return t.ExitRaw()
}

func (t *Terminal) EnterRaw() error {
	err := setTermios(&t.termios)
	if err != nil {
		return err
	}
	print("\x1b[?2004h") // enable bracketed paste mode
	return nil
}

func (t *Terminal) ExitRaw() error {
	print("\x1b[?2004l") // disable bracketed paste mode
	return setTermios(&t.oldTermios)
}
