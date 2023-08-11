package loggerfactory

import (
	logrus "github.com/sirupsen/logrus"
)

type loggerFactory struct {
	Level  logrus.Level
	Format logrus.Formatter
	Fields logrus.Fields
}


func NewDefaultLoggerFactory(level logrus.Level, component string) *loggerFactory {
	
	return &loggerFactory{
		Level:  level,
		Format: nil, //default logger format will be json
		Fields: logrus.Fields{"component": component},
	}
}

func CustomLoggerFactory(fields logrus.Fields, 	format logrus.Formatter, level logrus.Level) *loggerFactory {
	return &loggerFactory{
		Level:  logrus.InfoLevel,
		Format: format, //default logger format will be json
		Fields: fields,
	}
}

func (lf *loggerFactory) NewLogger() *logrus.Logger {
	logger := logrus.New()

	logger.SetLevel(lf.Level)

	if lf.Format != nil {
		logger.SetFormatter(lf.Format)
	} else {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}
	logger.WithFields(lf.Fields)
	return logger
}
