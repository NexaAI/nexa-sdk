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
	cursorIndex  int
	cursorHeight int
}

func NewBuffer(prompt, altPrompt string, getWidth func() (int, error)) *Buffer {
	return &Buffer{
		prompt:    prompt,
		altPrompt: altPrompt,
		getWidth:  getWidth,
		data:      make([]rune, 0),
	}
}

func (rl *Buffer) resetState() {
	rl.data = rl.data[:0]
	rl.cursorIndex = 0
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
	curHeight := 1
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
			curHeight++
			curWidth = calcANSIWidth(rl.altPrompt)
		} else if curWidth+rw == width {
			// exactly fit
			buffer += string(r)
			buffer += "\n"
			buffer += rl.altPrompt
			curHeight++
			curWidth = calcANSIWidth(rl.altPrompt)
		} else if curWidth+rw > width {
			// over flow
			buffer += "\n"
			buffer += rl.altPrompt
			curHeight++
			buffer += string(r)
			curWidth += rw
		} else {
			// normal char
			buffer += string(r)
			curWidth += rw
		}
		// record cursor position
		if i == rl.cursorIndex-1 {
			cursorHeight = curHeight
			cursorWidth = curWidth
		}
	}

	// move cursor to the position

	rl.cursorHeight = cursorHeight
	if curHeight > cursorHeight {
		buffer += fmt.Sprintf("\x1b[%dA", curHeight-cursorHeight)
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
