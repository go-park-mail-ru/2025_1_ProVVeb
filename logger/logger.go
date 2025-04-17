package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Init(accessLogPath string) error {
	Log = logrus.New()
	Log.SetFormatter(&logrus.JSONFormatter{})

	// Настройка access log файла
	accessFile, err := os.OpenFile(accessLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	Log.SetOutput(accessFile)
	Log.SetLevel(logrus.InfoLevel)

	return nil
}

func GetLogger() *logrus.Logger {
	return Log
}
