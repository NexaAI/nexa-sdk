package readline

import (
	"fmt"
	"log/slog"
	"regexp"
	"strings"

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

	var buffer strings.Builder

	// move cursor to the top
	if rl.cursorHeight != 1 {
		fmt.Fprintf(&buffer, "\x1b[%dA", rl.cursorHeight-1)
	}

	// render lines

	curWidth := 0
	curHeight := 1
	cursorWidth := 0
	cursorHeight := 1

	buffer.WriteString("\x1b[1G") // move cursor to beginning
	buffer.WriteString("\x1b[J")  // clean after
	buffer.WriteString(rl.prompt)
	curWidth += calcANSIWidth(rl.prompt)
	cursorWidth = curWidth

	for i, r := range rl.data {
		// line wrap
		rw := runewidth.RuneWidth(r)
		if r == CtrlJ {
			// new line
			buffer.WriteString("\n")
			buffer.WriteString(rl.altPrompt)
			curHeight++
			curWidth = calcANSIWidth(rl.altPrompt)
		} else if curWidth+rw == width {
			// exactly fit
			buffer.WriteString(string(r))
			buffer.WriteString("\n")
			buffer.WriteString(rl.altPrompt)
			curHeight++
			curWidth = calcANSIWidth(rl.altPrompt)
		} else if curWidth+rw > width {
			// over flow
			buffer.WriteString("\n")
			buffer.WriteString(rl.altPrompt)
			curHeight++
			buffer.WriteString(string(r))
			curWidth += rw
		} else {
			// normal char
			buffer.WriteString(string(r))
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
		fmt.Fprintf(&buffer, "\x1b[%dA", curHeight-cursorHeight)
	}
	buffer.WriteString("\x1b[1G") // move cursor to beginning
	if cursorHeight > 1 {
		fmt.Fprintf(&buffer, "\x1b[%dC", cursorWidth)
	} else {
		fmt.Fprintf(&buffer, "\x1b[%dC", cursorWidth)
	}

	print(buffer.String())
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func calcANSIWidth(s string) int {
	return runewidth.StringWidth(ansiRegexp.ReplaceAllString(s, ""))
}
