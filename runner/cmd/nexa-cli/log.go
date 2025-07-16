package main

import (
	"log/slog"

	"github.com/NexaAI/nexa-sdk/internal/config"
)

const (
	LogLevelTrace string = "trace"
	LogLevelDebug string = "debug"
	LogLevelInfo  string = "info"
	LogLevelWarn  string = "warn"
	LogLevelError string = "error"
)

func applyLogLevel() {
	switch config.Get().Log {
	case LogLevelTrace, LogLevelDebug:
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case LogLevelInfo:
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case LogLevelWarn:
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case LogLevelError:
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}
