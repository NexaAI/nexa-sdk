// Copyright (c) 2025 Nexa AI
//
// LICENSE NOTICE - DUAL LICENSING:
// - NPU models and inference: CC-BY-NC 4.0 (NON-COMMERCIAL USE ONLY)
// - GPU/CPU models and inference: Apache 2.0 (FREE FOR ALL USE)

package render

import (
	"fmt"
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
		theme = defaultColorTheme()
	})
	return theme
}

type Style interface {
	Sprint(...any) string
	Sprintf(string, ...any) string
	Code() string
}

type NoColor struct{}

func (n NoColor) Sprint(a ...any) string                 { return fmt.Sprint(a...) }
func (n NoColor) Sprintf(format string, a ...any) string { return fmt.Sprintf(format, a...) }
func (n NoColor) Code() string                           { return "" }

func (t *Theme) Set(s Style) {
	color.SetTerminal(s.Code())
}

func (t *Theme) Reset() {
	color.ResetTerminal()
}

func noColorTheme() *Theme {
	return &Theme{
		Normal:      NoColor{},
		Info:        NoColor{},
		Success:     NoColor{},
		Warning:     NoColor{},
		Error:       NoColor{},
		Quant:       NoColor{},
		Prompt:      NoColor{},
		AddFiles:    NoColor{},
		ThinkOutput: NoColor{},
		ModelOutput: NoColor{},
		Profile:     NoColor{},
	}
}

func defaultColorTheme() *Theme {
	theme := noColorTheme()
	if color.SupportColor() {
		slog.Debug("apply 16 color")
		theme.Normal = NoColor{}
		theme.Info = color.Style{color.FgBlue}
		theme.Success = color.Style{color.FgGreen}
		theme.Warning = color.Style{color.FgYellow}
		theme.Error = color.Style{color.FgRed}

		theme.Quant = color.Style{color.FgLightBlue}
		theme.Prompt = color.Style{color.FgGreen, color.Bold}
		theme.AddFiles = color.Style{color.FgWhite}
		theme.ThinkOutput = color.Style{color.FgDarkGray}
		theme.ModelOutput = NoColor{}
		theme.Profile = color.Style{color.FgLightCyan}
	}
	if color.Support256Color() {
		slog.Debug("apply 256 color")
		theme.Quant = color.S256(31)
		//theme.AddFiles = color.S256(7)
		//theme.ModelOutput = color.S256(15)
		theme.Profile = color.S256(44)
	}
	if color.SupportTrueColor() {
		slog.Debug("apply true color")
		theme.Quant = color.NewRGBStyle(color.RGB(0, 135, 175))
		//theme.AddFiles = color.NewRGBStyle(color.RGB(192, 192, 192))
		theme.Profile = color.NewRGBStyle(color.RGB(0, 215, 215))
	}
	return theme
}
