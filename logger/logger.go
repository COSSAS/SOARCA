package logger

import (
	"fmt"
	"os"

	logrus "github.com/sirupsen/logrus"
)

type Severity int
type Format int
type FileName string

// Custom type struct so we only need to import this logging package
const (
	Panic Severity = iota
	Fatal
	Error
	Warning
	Info
	Debug
	Trace
)

const (
	Json Format = iota
	Text
)

type Log struct {
	*logrus.Entry
}

func Logger(name string, severity Severity, fileName FileName, format Format) *Log {
	logger := logrus.New()

	if fileName != "" {
		file, err := os.OpenFile(string(fileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Failed to open file ")
		} else {
			logger.Out = file
		}
	}

	logger.SetLevel(logrus.Level(severity))

	if format == Json {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	entry := logger.WithFields(logrus.Fields{"component": name})

	return &Log{entry}
}
