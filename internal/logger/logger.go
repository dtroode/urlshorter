package logger

import (
	"go.uber.org/zap"
)

type Log struct {
	*zap.SugaredLogger
}

func NewLog(level string) (*Log, error) {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &Log{
		SugaredLogger: zl.Sugar(),
	}, nil
}

func (l *Log) Print(v ...interface{}) {
	l.Infoln(v)
}
