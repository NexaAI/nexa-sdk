package common

import (
	"io"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"github.com/NexaAI/nexa-sdk/runner/internal/config"
	nexa_sdk "github.com/NexaAI/nexa-sdk/runner/nexa-sdk"
)

const (
	LogLevelNone  string = "none"
	LogLevelTrace string = "trace"
	LogLevelDebug string = "debug"
	LogLevelInfo  string = "info"
	LogLevelWarn  string = "warn"
	LogLevelError string = "error"
)

func ApplyLogLevel() {
	options := tint.Options{AddSource: true}

	if os.Getenv("NO_COLOR") == "1" {
		options.NoColor = true
	}

	switch config.GetLog() {
	case LogLevelNone:
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
		return
	case LogLevelTrace:
		nexa_sdk.EnableBridgeLog(true)
		options.Level = slog.LevelDebug
	case LogLevelDebug:
		options.Level = slog.LevelDebug
	case LogLevelInfo:
		options.Level = slog.LevelInfo
	case LogLevelWarn:
		options.Level = slog.LevelWarn
	case LogLevelError:
		options.Level = slog.LevelError
	}

	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &options)))
}
