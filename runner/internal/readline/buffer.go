package readline

import (
	"fmt"
	"log/slog"

	"github.com/mattn/go-runewidth"
)

// TODO: placeholder
type Buffer struct {
	// configuration
	prompt    string
	altPrompt string
	getWidth  func() (int, error)

	data []rune

	// state
	cursor int
	height int
}

func NewBuffer() *Buffer {
	return &Buffer{
		data: make([]rune, 0),
	}
}

func (rl *Buffer) resetState() {
	rl.data = rl.data[:0]
	rl.cursor = 0
	rl.height = 1
}

func (rl *Buffer) refresh() {
	width, err := rl.getWidth()
	if err != nil {
		width = 80
		slog.Warn("failed to get terminal width", "error", err)
	}

	// check min width
	if width <= runewidth.StringWidth(rl.prompt)+4 || width <= runewidth.StringWidth(rl.altPrompt)+4 {
		print("terminal width is too small\n")
		return
	}

	// move cursor to the top
	if rl.height > 1 {
		fmt.Printf("\x1b[%dA", rl.height-1)
	}

	// render lines

	curLine := 0
	rl.height = 1

	fmt.Printf("\x1b[1G") // move cursor to beginning
	fmt.Printf("\x1b[J")  // clean after
	print(rl.prompt)
	curLine += runewidth.StringWidth(rl.prompt)

	for _, r := range rl.data {
		// line wrap
		rw := runewidth.RuneWidth(r)
		if r == CharCtrlJ || curLine+rw > width {
			print("\n")
			rl.height++
			curLine = 0
			print(rl.altPrompt)
			curLine += runewidth.StringWidth(rl.altPrompt)
		} else {
			print(string(r))
			curLine += rw
		}
	}
}
