package mongodb

import (
	"reflect"

	"soarca/internal/logger"
)

var log *logger.Log

type Empty struct{}

func LoadComponent() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Trace, "", logger.Json)
}
