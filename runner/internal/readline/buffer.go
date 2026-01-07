package readline

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/mattn/go-runewidth"
)

type Buffer struct {
	prompt    string
	altPrompt string
	getWidth  func() (int, error)

	data []rune
	r    *bufio.Reader

	cursor int
	height int
}

func NewBuffer() *Buffer {
	return &Buffer{
		data: make([]rune, 0),
	}
}

func (b *Buffer) Read() (string, error) {
	b.data = b.data[:0]
	b.r = bufio.NewReader(os.Stdin)
	b.cursor = 0
	b.height = 1
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
	event, exists := eventMap[r]
	if !exists {
		b.data = append(b.data, r)
		b.refresh()
		return nil
	}

	err := event(b)
	if err != nil {
		return err
	}

	b.refresh()
	return nil
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

	fmt.Printf("\x1b[2K") // clean current line
	fmt.Printf("\r")      // move cursor to beginning
	print(b.prompt)
	curLine += runewidth.StringWidth(b.prompt)

	for _, r := range b.data {

		// line wrap
		rw := runewidth.RuneWidth(r)
		if curLine+rw > width {
			print("\n")
			fmt.Printf("\x1b[2K")
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
