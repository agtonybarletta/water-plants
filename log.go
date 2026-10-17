package main

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func GetLogger() *slog.Logger {
	if Logger == nil {
		opts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		Logger = slog.New(slog.NewTextHandler(os.Stderr, opts))

	}
	return Logger
}
