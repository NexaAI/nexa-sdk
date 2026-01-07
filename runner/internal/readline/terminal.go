package readline

import (
	"bufio"
	"os"
	"syscall"
)

type Terminal struct {
	oldTermios *syscall.Termios
	r          *bufio.Reader
}

func NewTerminal() (*Terminal, error) {
	t := &Terminal{}

	termios, err := getTermios()
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

	setTermios(&newTermios)
	print("\x1b[?2004h") // enable bracketed paste mode"

	t.r = bufio.NewReader(os.Stdin)

	return t, nil
}

func (t *Terminal) Read() (rune, error) {
	r, _, err := t.r.ReadRune()
	return r, err
}

func (t *Terminal) Close() error {
	print("\x1b[?2004l") // disable bracketed paste mode
	return setTermios(t.oldTermios)
}
