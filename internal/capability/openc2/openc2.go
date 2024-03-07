package http

import (
	"reflect"

	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
)

type OpenC2Capability struct{}

type Empty struct{}

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// What to do if there is no agent or target?
// And maybe no auth info either?

func (httpCapability *OpenC2Capability) Execute(
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.VariableMap) (cacao.VariableMap, error) {
}
