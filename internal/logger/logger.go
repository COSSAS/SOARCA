package logger

import (
	"fmt"
	"os"
	"soarca/pkg/utils"
	"strings"

	logrus "github.com/sirupsen/logrus"
)

type Severity int
type Format int
type FileName string

const production = "production"
const development = "development"
const defaultSeverity = "info"
const defaultLogPath = ""
const formatStringJSon = "json"
const formatStringText = "text"

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

func severityFromString(name string) Severity {
	nameToLower := strings.ToLower(name)
	switch nameToLower {
	case "panic":
		return Panic
	case "fatal":
		return Fatal
	case "error":
		return Error
	case "warning":
		return Warning
	case "info":
		return Info
	case "debug":
		return Debug
	case "trace":
		return Trace

	default:
		return Info
	}
}

const (
	Json Format = iota
	Text
)

func (format Format) fromString(name string) Format {
	nameToLower := strings.ToLower(name)
	switch nameToLower {
	case formatStringJSon:
		return Json
	case formatStringText:
		return Text

	default:
		return Json
	}
}

func setFormat(instance *logrus.Logger, format Format) {
	switch format {
	case Text:
		instance.SetFormatter(&logrus.TextFormatter{})
	case Json:
		instance.SetFormatter(&logrus.JSONFormatter{})

	}
}

type Log struct {
	*logrus.Entry
}

func Logger(name string, severity Severity, fileName FileName, format Format) *Log {

	globalLogSeverity := utils.GetEnv("LOG_GLOBAL_LEVEL", defaultSeverity)
	globalOperationMode := utils.GetEnv("LOG_MODE", production)
	globalLogFilePath := utils.GetEnv("LOG_FILE_PATH", defaultLogPath)
	globalLogFormat := format.fromString(utils.GetEnv("LOG_FORMAT", formatStringJSon))

	instance := logrus.New()

	if globalLogFilePath != "" {
		file, err := os.OpenFile(string(globalLogFilePath), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Failed to open file ")
		} else {
			instance.Out = file
		}
	}

	setFormat(instance, globalLogFormat)

	globalSeverityLevel := severityFromString(globalLogSeverity)
	if globalSeverityLevel > severity {
		instance.SetLevel(logrus.Level(globalSeverityLevel))
	} else {
		instance.SetLevel(logrus.Level(severity))
	}

	if globalOperationMode == development {
		if fileName != "" {
			file, err := os.OpenFile(string(fileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("Failed to open file ")
			} else {
				instance.Out = file
			}
		}

		instance.SetLevel(logrus.Level(severity))
		setFormat(instance, format)
	}

	entry := instance.WithFields(logrus.Fields{"component": name})

	return &Log{entry}
}
