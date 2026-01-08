package readline

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/mattn/go-runewidth"
)

// TODO: placeholder
type Buffer struct {
	// configuration
	prompt    string
	altPrompt string
	getWidth  func() (int, error)

	// state
	data         []rune
	height       int
	cursor       int
	cursorHeight int
}

func NewBuffer() *Buffer {
	return &Buffer{
		data: make([]rune, 0),
	}
}

func (rl *Buffer) resetState() {
	rl.data = rl.data[:0]
	rl.height = 1
	rl.cursor = 0
	rl.cursorHeight = 1
}

func (rl *Buffer) refresh() {
	width, err := rl.getWidth()
	if err != nil {
		width = 80
		slog.Warn("failed to get terminal width", "error", err)
	}

	// check min width
	if width <= runewidth.StringWidth(rl.prompt)+4 || width <= runewidth.StringWidth(rl.altPrompt)+4 {
		print("\x1b[H\x1b[2J")
		print("terminal width is too small!")
		return
	}

	buffer := ""

	// move cursor to the top
	if rl.cursorHeight != 1 {
		buffer += fmt.Sprintf("\x1b[%dA", rl.cursorHeight-1)
	}

	// render lines

	curWidth := 0
	rl.height = 1
	cursorWidth := 0
	cursorHeight := 1

	buffer += "\x1b[1G" // move cursor to beginning
	buffer += "\x1b[J"  // clean after
	buffer += rl.prompt
	curWidth += calcANSIWidth(rl.prompt)
	cursorWidth = curWidth

	for i, r := range rl.data {
		// line wrap
		rw := runewidth.RuneWidth(r)
		if r == CtrlJ {
			// new line
			buffer += "\n"
			buffer += rl.altPrompt
			rl.height++
			curWidth = calcANSIWidth(rl.altPrompt)
		} else if curWidth+rw == width {
			// exactly fit
			buffer += string(r)
			buffer += "\n"
			buffer += rl.altPrompt
			rl.height++
			curWidth = calcANSIWidth(rl.altPrompt)
		} else if curWidth+rw > width {
			// over flow
			buffer += "\n"
			buffer += rl.altPrompt
			rl.height++
			buffer += string(r)
			curWidth += rw
		} else {
			// normal char
			buffer += string(r)
			curWidth += rw
		}
		// record cursor position
		if i == rl.cursor-1 {
			cursorHeight = rl.height
			cursorWidth = curWidth
		}
	}

	// move cursor to the position

	rl.cursorHeight = cursorHeight
	if rl.height > cursorHeight {
		buffer += fmt.Sprintf("\x1b[%dA", rl.height-cursorHeight)
	}
	buffer += "\x1b[1G" // move cursor to beginning
	if cursorHeight > 1 {
		buffer += fmt.Sprintf("\x1b[%dC", cursorWidth)
	} else {
		buffer += fmt.Sprintf("\x1b[%dC", cursorWidth)
	}

	print(buffer)
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func calcANSIWidth(s string) int {
	return runewidth.StringWidth(ansiRegexp.ReplaceAllString(s, ""))
}
