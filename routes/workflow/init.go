package workflow

import (
	loggerfactory "soarca/loggerfactory"

	logrus "github.com/sirupsen/logrus"
)

const component  = "WORKFLOW"
var logger *logrus.Logger

func init(){
	loggerFactory := loggerfactory.NewDefaultLoggerFactory(logrus.InfoLevel, component)
	logger = loggerFactory.NewLogger()
}