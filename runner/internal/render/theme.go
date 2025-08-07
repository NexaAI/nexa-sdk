package render

import (
	"log/slog"
	"sync"

	"github.com/gookit/color"
)

var theme *Theme
var themeInit sync.Once

type Theme struct {
	Normal  Style
	Info    Style
	Success Style
	Warning Style
	Error   Style

	Quant       Style
	Prompt      Style
	AddFiles    Style
	ThinkOutput Style
	ModelOutput Style
	Profile     Style
}

func GetTheme() *Theme {
	themeInit.Do(func() {
		slog.Debug("Detect terminal theme")
		theme = &defaultTrueColorTheme
	})
	return theme
}

type Style interface {
	Sprint(...any) string
	Sprintf(string, ...any) string
	Code() string
}

func (t *Theme) Set(s Style) {
	color.SetTerminal(s.Code())
}

func (t *Theme) Reset() {
	color.ResetTerminal()
}

var defaultTrueColorTheme = Theme{
	Normal:  color.Style{color.FgWhite},
	Info:    color.Style{color.FgBlue},
	Success: color.Style{color.FgGreen},
	Warning: color.Style{color.FgYellow},
	Error:   color.Style{color.FgRed},

	Quant:       color.NewRGBStyle(color.RGB(0, 135, 175)),
	Prompt:      color.Style{color.FgGreen, color.Bold},
	AddFiles:    color.NewRGBStyle(color.RGB(192, 192, 192)),
	ThinkOutput: color.Style{color.FgDarkGray},
	ModelOutput: color.Style{color.FgWhite},
	Profile:     color.NewRGBStyle(color.RGB(0, 215, 215)),
}
