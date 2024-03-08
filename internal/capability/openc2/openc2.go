package openc2

import (
	"reflect"

	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/utils/http"
)

type OpenC2Capability struct {
	httpRequest http.IHttpRequest
}

type Empty struct{}

const (
	openc2ResultVariableName = "__soarca_openc2_result__"
	openc2capabilityName     = "soarca-openc2-capability"
)

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(httpRequest http.IHttpRequest) *OpenC2Capability {
	return &OpenC2Capability{httpRequest: httpRequest}
}

func (OpenC2Capability *OpenC2Capability) GetType() string {
	return openc2capabilityName
}

func (OpenC2Capability *OpenC2Capability) Execute(
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.VariableMap,
) (cacao.VariableMap, error) {
	log.Trace(metadata.ExecutionId)

	httpOptions := http.HttpOptions{
		Command: &command,
		Target:  &target,
		Auth:    &authentication,
	}
	response, err := OpenC2Capability.httpRequest.Request(httpOptions)
	if err != nil {
		log.Error(err)
		return cacao.VariableMap{}, err
	}

	results := cacao.VariableMap{openc2ResultVariableName: {Name: "result", Value: string(response)}}
	log.Trace("Finished openc2 execution, will return the variables: ", results)
	return results, nil
}
