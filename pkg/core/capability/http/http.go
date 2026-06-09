package http

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/utils/http"
)

// Receive HTTP API command data from decomposer/executor
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
	metadata execution.Metadata,
	context capability.Context) (cacao.Variables, error) {

	soarca_http_options := http.HttpOptions{
		Target:  &context.Target,
		Command: &context.Command,
		Auth:    &context.Authentication,
	}

	responseBytes, err := httpCapability.soarca_http_request.Request(soarca_http_options)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	respString := string(responseBytes)
	variable := cacao.Variable{Type: cacao.VariableTypeString,
		Name:  httpApiResultVariableName,
		Value: respString}

	return cacao.NewVariables(variable), nil

}
