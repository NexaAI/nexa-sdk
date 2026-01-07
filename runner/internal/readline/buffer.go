package readline

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/mattn/go-runewidth"
)

// TODO: placeholder
type Buffer struct {
	// configuration
	prompt    string
	altPrompt string
	getWidth  func() (int, error)

	r *bufio.Reader

	// state
	data  []rune
	esc   bool
	escEx bool

	// rendering state
	cursor int
	height int
}

func NewBuffer() *Buffer {
	return &Buffer{
		data: make([]rune, 0),
	}
}

func (b *Buffer) Read() (string, error) {
	b.r = bufio.NewReader(os.Stdin)
	b.resetState()
	b.refresh()

	for {
		r, _, err := b.r.ReadRune()
		if err != nil {
			return "", err
		}

		if err := b.parse(r); err != nil {
			if errors.Is(err, ErrComplete) {
				return string(b.data), nil
			}
			return "", err
		}
	}
}

func (b *Buffer) parse(r rune) error {
	var event func(*Buffer) error

	if b.escEx {
		b.escEx = false
		if ev, ok := escExEventMap[r]; ok {
			event = ev
		}
	} else if b.esc {
		b.esc = false
		if ev, ok := escEventMap[r]; ok {
			event = ev
		}
	} else {
		if ev, ok := eventMap[r]; ok {
			event = ev
		} else {
			b.data = append(b.data, r)
		}
	}

	if event != nil {
		err := event(b)
		if err != nil {
			return err
		}
	}

	b.refresh()
	return nil
}

func (b *Buffer) resetState() {
	b.data = b.data[:0]
	b.cursor = 0
	b.height = 1
}

func (b *Buffer) refresh() {
	width, err := b.getWidth()
	if err != nil {
		width = 80
		slog.Warn("failed to get terminal width", "error", err)
	}

	// check min width
	if width <= runewidth.StringWidth(b.prompt)+4 || width <= runewidth.StringWidth(b.altPrompt)+4 {
		print("terminal width is too small\n")
		return
	}

	// move cursor to the top
	if b.height > 1 {
		fmt.Printf("\x1b[%dA", b.height-1)
	}

	// render lines

	curLine := 0
	b.height = 1

	fmt.Printf("\r")     // move cursor to beginning
	fmt.Printf("\x1b[J") // clean after
	print(b.prompt)
	curLine += runewidth.StringWidth(b.prompt)

	for _, r := range b.data {

		// line wrap
		rw := runewidth.RuneWidth(r)
		if curLine+rw > width {
			print("\n")
			b.height++
			curLine = 0
			print(b.altPrompt)
			curLine += runewidth.StringWidth(b.altPrompt)
		} else {
			print(string(r))
			curLine += rw
		}
	}
}
