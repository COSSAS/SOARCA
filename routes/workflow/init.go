package workflow

import (
	"soarca/logger"
)

const component = "WORKFLOW"

var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}
