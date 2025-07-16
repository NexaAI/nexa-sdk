package main

import (
	"log/slog"

	"github.com/NexaAI/nexa-sdk/internal/config"
	nexa_sdk "github.com/NexaAI/nexa-sdk/nexa-sdk"
)

const (
	LogLevelTrace string = "trace"
	LogLevelDebug string = "debug"
	LogLevelInfo  string = "info"
	LogLevelWarn  string = "warn"
	LogLevelError string = "error"
)

func applyLogLevel() {

	if config.Get().Log == LogLevelTrace {
		nexa_sdk.SetLog(func(msg string) {
			slog.Debug("[Bridge] " + msg)
		})
	}

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
