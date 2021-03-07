package config

import (
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, DisableColors: true})
}

// Logger is the logger type
type Logger struct {
	Writer        io.Writer
	ContextLogger *log.Entry
}

func newLogger(appName string) *Logger {
	writer := os.Stderr
	contextLogger := log.WithFields(log.Fields{
		"app_name": appName,
	})

	return &Logger{
		Writer:        writer,
		ContextLogger: contextLogger,
	}
}
