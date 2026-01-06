package readline

import "errors"

var (
	ErrComplete = errors.New("Complete") // readline complete

	ErrInterrupt = errors.New("Interrupt")
)

const (
	CharNull      = 0
	CharLineStart = 1
	CharBackward  = 2
	CharInterrupt = 3
	CharDelete    = 4
	CharLineEnd   = 5
	CharForward   = 6
	CharBell      = 7
	CharCtrlH     = 8
	CharTab       = 9
	CharCtrlJ     = 10
	CharKill      = 11
	CharCtrlL     = 12
	CharEnter     = 13
	CharNext      = 14
	CharPrev      = 16
	CharBckSearch = 18
	CharFwdSearch = 19
	CharTranspose = 20
	CharCtrlU     = 21
	CharCtrlW     = 23
	CharCtrlY     = 25
	CharCtrlZ     = 26
	CharEsc       = 27
	CharEscapeEx  = 91
	CharBackspace = 127
)

var eventMap = map[rune]func(*Buffer) error{
	CharNull:      noop,
	CharLineStart: noop,
	CharBackward:  noop,
	CharInterrupt: interrupt,
	CharDelete:    delete,
	CharLineEnd:   noop,
	CharForward:   noop,
	CharBell:      noop,
	CharCtrlH:     noop,
	CharTab:       noop,
	CharCtrlJ:     noop,
	CharKill:      noop,
	CharCtrlL:     noop,
	CharEnter:     enter,
	CharNext:      noop,
	CharPrev:      noop,
	CharBckSearch: noop,
	CharFwdSearch: noop,
	CharTranspose: noop,
	CharCtrlU:     noop,
	CharCtrlW:     noop,
	CharCtrlY:     noop,
	CharCtrlZ:     noop,
	CharEsc:       noop,
	CharEscapeEx:  noop,
	CharBackspace: backspace,
}

func interrupt(b *Buffer) error {
	if len(b.data) == 0 {
		println("^C")
		return ErrInterrupt
	}

	b.data = b.data[:0]
	println()
	return nil
}

func delete(b *Buffer) error {
	if len(b.data) > 0 {
		b.data = b.data[:len(b.data)-1]
	}
	return nil
}

func enter(b *Buffer) error {
	println()
	return ErrComplete
}

func backspace(b *Buffer) error {
	if len(b.data) > 0 {
		b.data = b.data[:len(b.data)-1]
	}
	return nil
}

func noop(*Buffer) error {
	return nil
}
