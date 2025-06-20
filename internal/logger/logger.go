package logger

import (
	"log/slog"
	"os"
)

// Logger represents application logger.
type Logger struct {
	*slog.Logger
}

// NewLog creates new Logger instance.
func NewLog(level string) *Logger {
	return &Logger{
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
}
