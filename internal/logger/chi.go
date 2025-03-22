package logger

import (
	"net/http"
	"time"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type LogEntry struct {
	l Log
}

func (e LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	e.l.Infoln(
		"status", status,
		"size", bytes,
		"header", header,
		"elapsed", elapsed,
	)
}

func (e LogEntry) Panic(v interface{}, stack []byte) {
	chiMiddleware.PrintPrettyStack(v)
}

type LogFormatter struct {
	l Log
}

func (f *LogFormatter) NewLogEntry(r *http.Request) chiMiddleware.LogEntry {
	return LogEntry{
		l: f.l,
	}
}

func NewLogFormatter(l Log) *LogFormatter {
	return &LogFormatter{
		l: l,
	}
}
