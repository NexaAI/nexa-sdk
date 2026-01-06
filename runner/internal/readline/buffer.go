package readline

import (
	"bufio"
	"errors"
	"os"
)

type Buffer struct {
	prompt string

	data []rune
	r    *bufio.Reader
}

func NewBuffer() *Buffer {
	return &Buffer{
		data: make([]rune, 0),
	}
}

func (b *Buffer) Read() (string, error) {
	b.data = b.data[:0]
	b.r = bufio.NewReader(os.Stdin)
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
	print("\r")
	print(b.prompt)
	print(string(b.data))
}
