package routes

import (
	"soarca/logger"
)

var log *logger.Log

const component = "api"

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}
