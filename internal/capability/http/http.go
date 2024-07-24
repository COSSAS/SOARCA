package http

import (
	"reflect"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
	"soarca/utils/http"
	"soarca/utils/mapper"
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
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.Variables,
	inputVariableKeys []string,
	outputVariablesKeys []string) (cacao.Variables, error) {

	soarca_http_options := http.HttpOptions{
		Target:  &target,
		Command: &command,
		Auth:    &authentication,
	}

	responseBytes, err := httpCapability.soarca_http_request.Request(soarca_http_options)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}

	response := string(responseBytes)

	results := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString,
		Name:  httpApiResultVariableName,
		Value: string(response)})
	log.Trace("Finished https execution, will return the variables: ", results)

	return mapper.Variables(variables, outputVariablesKeys, results, []string{httpApiResultVariableName})

}
