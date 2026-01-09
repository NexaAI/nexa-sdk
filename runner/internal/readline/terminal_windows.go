package readline

import (
	"errors"
)

type Terminal struct {
}

func NewTerminal() (*Terminal, error) {
	return nil, errors.New("terminal not supported on Windows")
}

func (t *Terminal) Read() (rune, error) {
	return 'a', errors.New("terminal not supported on Windows")
}

func (t *Terminal) Close() error {
	return t.ExitRaw()
}

func (t *Terminal) EnterRaw() error {
	return errors.New("terminal not supported on Windows")

}

func (t *Terminal) ExitRaw() error {
	return errors.New("terminal not supported on Windows")
}

func (t *Terminal) GetWidth() (int, error) {
	return 0, errors.New("terminal not supported on Windows")
}
