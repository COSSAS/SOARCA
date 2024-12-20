package openc2

import (
	"reflect"

	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/utils/http"
)

type OpenC2Capability struct {
	httpRequest http.IHttpRequest
}

type Empty struct{}

const (
	openc2ResultVariableName = "__soarca_openc2_http_result__"
	openc2CapabilityName     = "soarca-openc2-http"
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
	return openc2CapabilityName
}

func (OpenC2Capability *OpenC2Capability) Execute(
	metadata execution.Metadata,
	context capability.Context,
) (cacao.Variables, error) {
	log.Trace(metadata.ExecutionId)

	httpOptions := http.HttpOptions{
		Command: &context.CommandData,
		Target:  &context.Target,
		Auth:    &context.Authentication,
	}
	response, err := OpenC2Capability.httpRequest.Request(httpOptions)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}

	results := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString,
		Name:  openc2ResultVariableName,
		Value: string(response)})
	log.Trace("Finished openc2 execution, will return the variables: ", results)
	return results, nil
}
