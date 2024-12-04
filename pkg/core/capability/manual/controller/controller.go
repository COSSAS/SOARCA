package controller

import (
	"reflect"
	"soarca/internal/logger"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ManualCommandInfo struct {
	Name  string
	Id    string
	FinId string
}

type IManualController interface {
	GetPendingCommands() map[string]ManualCommandInfo
	GetPendingCommand() map[string]ManualCommandInfo
	PostContinue() map[string]ManualCommandInfo
}

type ManualController struct {
	manualCommandsRegistry map[string]ManualCommandInfo
}

func (controller *ManualController) GetPendingCommands()
