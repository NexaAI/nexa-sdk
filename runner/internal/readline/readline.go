package readline

import (
	"errors"
)

type Config struct {
	Prompt    string
	AltPrompt string
}

type Readline struct {
	term *Terminal
	buf  *Buffer

	eventMap map[rune]func() error

	// csi
	isEsc       bool
	isEscEx     bool
	escBuf      string
	csiEventMap map[string]func() error
	isPaste     bool
}

func New(config *Config) (*Readline, error) {
	term, err := NewTerminal()
	if err != nil {
		return nil, err
	}

	buf := NewBuffer()
	buf.prompt = config.Prompt
	buf.altPrompt = config.AltPrompt
	buf.getWidth = term.GetWidth

	rl := Readline{
		term: term,
		buf:  buf,
	}
	rl.initializeEventMaps()
	return &rl, nil
}

func (rl *Readline) Read() (string, error) {
	rl.buf.resetState()
	rl.buf.refresh()

	if err := rl.term.EnterRaw(); err != nil {
		return "", err
	}
	defer rl.term.ExitRaw()

	for {
		r, err := rl.term.Read()
		if err != nil {
			return "", err
		}

		if err := rl.parse(r); err != nil {
			if errors.Is(err, ErrComplete) {
				return string(rl.buf.data), nil
			}
			return "", err
		}
	}
}

func (rl *Readline) parse(r rune) error {
	if rl.isEsc {
		// escape sequence

		rl.isEsc = false
		if r == '[' {
			rl.isEsc = false
			rl.isEscEx = true
			rl.escBuf = ""
			return nil
		}

	} else if rl.isEscEx {
		// escape ex sequence

		if r < 0x20 || r >= 0x80 {
			// invalid char, end escape ex
			rl.isEscEx = false
		}
		rl.escBuf += string(r)
		if r >= 0x40 {
			// end of escape ex
			rl.isEscEx = false
			if event, ok := rl.csiEventMap[rl.escBuf]; !ok {
				print("\nUnknown escape ex sequence:", rl.escBuf, "\n") // for debug
			} else if err := event(); err != nil {
				return err
			}
		}

	} else {
		// single char
		// print("read rune: ", r, "\n") // for debug

		if event, ok := rl.eventMap[r]; !ok {
			rl.buf.data = append(rl.buf.data, r)
		} else if err := event(); err != nil {
			return err
		}
	}

	rl.buf.refresh()
	return nil
}

func (rl *Readline) Close() error {
	return rl.term.Close()
}
