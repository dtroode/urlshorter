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

// Fatal is equivalent to [Error] followed by a call to [os.Exit](1).
func (l *Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}
