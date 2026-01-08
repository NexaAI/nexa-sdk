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
	Null      = 0
	CtrlA     = 1
	CtrlB     = 2
	CtrlC     = 3
	CtrlD     = 4
	CtrlE     = 5
	CtrlF     = 6
	Bell      = 7
	CtrlH     = 8
	Tab       = 9
	CtrlJ     = 10
	Kill      = 11
	CtrlL     = 12
	Enter     = 13
	Next      = 14
	Prev      = 16
	BckSearch = 18
	FwdSearch = 19
	Transpose = 20
	CtrlU     = 21
	CtrlW     = 23
	CtrlY     = 25
	CtrlZ     = 26
	Esc       = 27
	Backspace = 127
)

func (rl *Readline) initializeEventMaps() {
	rl.eventMap = map[rune]func() error{
		Null:  rl.noop,
		CtrlA: rl.begin,
		CtrlB: rl.left,
		CtrlC: rl.interrupt,
		CtrlD: rl.eof,
		CtrlE: rl.end,
		CtrlF: rl.right,
		Bell:  rl.noop,
		CtrlH: rl.noop,
		Tab:   rl.noop,
		// CtrlJ:     rl.lf,
		Kill:      rl.noop,
		CtrlL:     rl.clear,
		Enter:     rl.enter,
		Next:      rl.noop,
		Prev:      rl.noop,
		BckSearch: rl.noop,
		FwdSearch: rl.noop,
		Transpose: rl.noop,
		CtrlU:     rl.noop,
		CtrlW:     rl.noop,
		CtrlY:     rl.noop,
		CtrlZ:     rl.noop,
		Esc:       rl.esc,
		Backspace: rl.backspace,
	}
	rl.csiEventMap = map[string]func() error{
		"200~": func() error { rl.isPaste = true; return nil },
		"201~": func() error { rl.isPaste = false; return nil },
		"3~":   rl.delete,
		"D":    rl.left,
		"C":    rl.right,
	}
}

func (rl *Readline) noop() error {
	return nil
}

// control actions

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

func (rl *Readline) esc() error {
	rl.isEsc = true
	return nil
}

// cursor move

func (rl *Readline) left() error {
	if rl.buf.cursor > 0 {
		rl.buf.cursor--
	}
	return nil
}

func (rl *Readline) right() error {
	if rl.buf.cursor < len(rl.buf.data) {
		rl.buf.cursor++
	}
	return nil
}

func (rl *Readline) begin() error {
	rl.buf.cursor = 0
	return nil
}

func (rl *Readline) end() error {
	rl.buf.cursor = len(rl.buf.data)
	return nil
}

// edit actions

func (rl *Readline) backspace() error {
	if len(rl.buf.data) > 0 {
		rl.buf.data = append(rl.buf.data[:rl.buf.cursor-1], rl.buf.data[rl.buf.cursor:]...)
		rl.buf.cursor--
	}
	return nil
}

func (rl *Readline) delete() error {
	if rl.buf.cursor < len(rl.buf.data) {
		rl.buf.data = append(rl.buf.data[:rl.buf.cursor], rl.buf.data[rl.buf.cursor+1:]...)
	}
	return nil
}

func (rl *Readline) clear() error {
	print("\x1b[H\x1b[2J") // clear screen
	return nil
}

func (rl *Readline) enter() error {
	println()
	return ErrComplete
}
