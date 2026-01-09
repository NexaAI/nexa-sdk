package readline

import (
	"bufio"
	"os"
)

type Terminal struct {
	oldState Termios
	state    Termios
	r        *bufio.Reader
}

func NewTerminal() (*Terminal, error) {
	t := &Terminal{}

	termios, err := getTermios()
	if err != nil {
		return nil, err
	}

	t.oldState = *termios
	t.state = *termios
	applyRawMode(&t.state)

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
	err := setTermios(&t.state)
	if err != nil {
		return err
	}
	print("\x1b[?2004h") // enable bracketed paste mode
	return nil
}

func (t *Terminal) ExitRaw() error {
	print("\x1b[?2004l") // disable bracketed paste mode
	return setTermios(&t.oldState)
}
