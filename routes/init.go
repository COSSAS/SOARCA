package routes

import (
	loggerfactory "soarca/loggerfactory"

	logrus "github.com/sirupsen/logrus"
)

var logger *logrus.Logger
const component = "api"

func init(){
	loggerFactory := loggerfactory.NewDefaultLoggerFactory(logrus.InfoLevel, component)
	logger = loggerFactory.NewLogger()
}