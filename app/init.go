package app

import (
	"reflect"

	"soarca/logger"
)

var log *logger.Log

type Empty struct{}

func LoadComponent() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}
