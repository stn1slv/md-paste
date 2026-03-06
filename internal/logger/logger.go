// Package logger provides initialization for structured logging.
package logger

import (
	"log/slog"
	"os"
)

// Init initializes the default logger for the application.
// We configure it to write to stderr so that stdout can be used purely for data output.
func Init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	// Use TextHandler to stderr to keep stdout clean for piping.
	handler := slog.NewTextHandler(os.Stderr, opts)
	slog.SetDefault(slog.New(handler))
}
