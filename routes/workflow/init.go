package workflow

import (
	"reflect"

	"soarca/logger"
)

var log *logger.Log

type Empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Debug, "", logger.Json)
}
