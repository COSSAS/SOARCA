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

	// This capability performs commands against a target; a step declaring
	// zero targets has nothing to run against, so skip without error.
	if len(context.Targets) == 0 {
		return cacao.NewVariables(), nil
	}
	targets := context.Targets

	returnVariables := cacao.NewVariables()
	var stepErr error

	for _, resolvedTarget := range targets {
		target := resolvedTarget.Target
		auth := resolvedTarget.Authentication

		for _, command := range context.Commands {
			httpOptions := http.HttpOptions{
				Command: &command,
				Target:  &target,
				Auth:    &auth,
			}
			response, err := OpenC2Capability.httpRequest.Request(httpOptions)
			if err != nil {
				log.Error(err)
				stepErr = err
				// Abort this target's remaining commands on first failure,
				// but keep processing the other targets.
				break
			}

			results := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString,
				Name:  openc2ResultVariableName,
				Value: string(response)})
			log.Trace("Finished openc2 execution, will return the variables: ", results)
			returnVariables.Merge(results)
		}
	}

	return returnVariables, stepErr
}
