package http

import (
	"reflect"

	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/utils/http"
)

type OpenC2Capability struct{}

type Empty struct{}

const capabilityName = "soarca-openc2-capability"

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func (OpenC2Capability *OpenC2Capability) GetType() string {
	return capabilityName
}

func (OpenC2Capability *OpenC2Capability) Execute(
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.VariableMap,
) (cacao.VariableMap, error) {
	httpRequest := http.HttpRequest{}
	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
	}
	response, err := httpRequest.Request(httpOptions)
	if err != nil {
		log.Error(err)
		return cacao.VariableMap{}, err
	}
	return cacao.VariableMap{
		capabilityName: {Name: "result", Value: string(response)},
	}, nil
}
