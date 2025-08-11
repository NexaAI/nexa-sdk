package main

import (
	"io"
	"log/slog"

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

func applyLogLevel() {
	switch config.Get().Log {
	case LogLevelNone:
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	case LogLevelTrace:
		nexa_sdk.EnableBridgeLog(true)
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case LogLevelDebug:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case LogLevelInfo:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case LogLevelWarn:
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case LogLevelError:
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}
