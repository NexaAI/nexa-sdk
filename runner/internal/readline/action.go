package readline

import (
	"errors"
	"io"
)

var (
	ErrComplete = errors.New("Complete") // readline complete

	ErrInterrupt = errors.New("Interrupt")
)

const (
	CharNull      = 0
	CharLineStart = 1
	CharBackward  = 2
	CharInterrupt = 3
	CharCtrlD     = 4
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
	CharBackspace = 127
)

const (
	Esc = "\x1b"
)

func (rl *Readline) initializeEventMaps() {
	rl.eventMap = map[rune]func() error{
		CharNull:      rl.noop,
		CharLineStart: rl.noop,
		CharBackward:  rl.noop,
		CharInterrupt: rl.interrupt,
		CharCtrlD:     rl.eof,
		CharLineEnd:   rl.noop,
		CharForward:   rl.noop,
		CharBell:      rl.noop,
		CharCtrlH:     rl.noop,
		CharTab:       rl.noop,
		CharCtrlJ:     rl.lf,
		CharKill:      rl.noop,
		CharCtrlL:     rl.noop,
		CharEnter:     rl.enter,
		CharNext:      rl.noop,
		CharPrev:      rl.noop,
		CharBckSearch: rl.noop,
		CharFwdSearch: rl.noop,
		CharTranspose: rl.noop,
		CharCtrlU:     rl.noop,
		CharCtrlW:     rl.noop,
		CharCtrlY:     rl.noop,
		CharCtrlZ:     rl.noop,
		CharEsc:       rl.esc,
		CharBackspace: rl.backspace,
	}
	rl.csiEventMap = map[string]func() error{
		"200~": func() error { rl.isPaste = true; return nil },
		"201~": func() error { rl.isPaste = false; return nil },
		"3~":   rl.delete,
	}
}

func (rl *Readline) interrupt() error {
	if len(rl.buf.data) == 0 {
		println("^C")
		return ErrInterrupt
	}

	rl.buf.resetState()
	println()
	return nil
}

func (rl *Readline) eof() error {
	if len(rl.buf.data) == 0 {
		return io.EOF
	}
	return nil
}

func (rl *Readline) delete() error {
	if len(rl.buf.data) > 0 {
		rl.buf.data = rl.buf.data[:len(rl.buf.data)-1]
	}
	return nil
}

func (rl *Readline) lf() error {
	if rl.isPaste {
		rl.buf.data = append(rl.buf.data, CharCtrlJ)
		return nil
	}
	println()
	return ErrComplete
}

func (rl *Readline) enter() error {
	println()
	return ErrComplete
}

func (rl *Readline) esc() error {
	rl.isEsc = true
	return nil
}

func (rl *Readline) backspace() error {
	if len(rl.buf.data) > 0 {
		rl.buf.data = rl.buf.data[:len(rl.buf.data)-1]
	}
	return nil
}

func (rl *Readline) noop() error {
	return nil
}
