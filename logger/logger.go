package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type Logger interface {
	WithFields(fields *logrus.Fields) *logrus.Entry
	WithError(err error) *logrus.Entry
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

type LogrusLogger struct {
	log *logrus.Logger
}

type Fields *logrus.Fields

func NewLogrusLogger(accessLogPath string) (*LogrusLogger, error) {
	var Log LogrusLogger
	Log.log = logrus.New()
	Log.log.SetFormatter(&logrus.JSONFormatter{})

	if err := os.MkdirAll(filepath.Dir(accessLogPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	accessFile, err := os.OpenFile(accessLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	Log.log.SetOutput(accessFile)
	Log.log.SetLevel(logrus.InfoLevel)

	return &Log, nil
}

func (l *LogrusLogger) WithFields(fields *logrus.Fields) *logrus.Entry {
	return l.log.WithFields(*fields)
}

func (l *LogrusLogger) WithError(err error) *logrus.Entry {
	return l.log.WithError(err)
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.log.Debug(args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.log.Info(args...)
}

func (l *LogrusLogger) Warn(args ...interface{}) {
	l.log.Warn(args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.log.Error(args...)
}
