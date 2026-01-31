package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON-structured logger for the parser.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// Default is the default logger instance.
var Default = New()
