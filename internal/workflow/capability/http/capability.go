package http

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"soarca/pkg/utils/http"
)

// Receive HTTP API command data from decomposer/executer
// Validate HTTP API call
// Run HTTP API call
// Return response

const (
	httpApiResultVariableName = "__soarca_http_api_result__"
	httpApiCapabilityName     = "soarca-http-api"
)

type HttpCapability struct {
	soarca_http_request http.IHttpRequest
}

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(httpRequest http.IHttpRequest) *HttpCapability {
	return &HttpCapability{soarca_http_request: httpRequest}
}

func (httpCapability *HttpCapability) GetType() string {
	return httpApiCapabilityName
}

func (httpCapability *HttpCapability) Execute(
	metadata run.Metadata,
	context capability.Context) (cacao.Variables, error) {

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
			soarca_http_options := http.HttpOptions{
				Target:  &target,
				Command: &command,
				Auth:    &auth,
			}

			responseBytes, err := httpCapability.soarca_http_request.Request(soarca_http_options)
			if err != nil {
				log.Error(err)
				stepErr = err
				// Abort this target's remaining commands on first failure,
				// but keep processing the other targets.
				break
			}
			respString := string(responseBytes)
			variable := cacao.Variable{Type: cacao.VariableTypeString,
				Name:  httpApiResultVariableName,
				Value: respString}

			returnVariables.Merge(cacao.NewVariables(variable))
		}
	}

	return returnVariables, stepErr
}
