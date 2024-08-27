package powershell

import (
	"context"
	"errors"
	"reflect"
	"strconv"

	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/masterzen/winrm"
)

type PowershellCapability struct {
}

type Empty struct{}

const (
	resultVariableName = "__soarca_powershell_result__"
	capabilityName     = "soarca-powershell"
)

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New() *PowershellCapability {
	return &PowershellCapability{}
}

func (capability *PowershellCapability) GetType() string {
	return capabilityName
}

func (capability *PowershellCapability) Execute(
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.Variables,
) (cacao.Variables, error) {
	log.Trace(metadata.ExecutionId)

	port, err := strconv.Atoi(target.Port)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}

	address, err := determineTargetAddress(target)
	if err != nil {
		return cacao.NewVariables(), err
	}

	endpoint := winrm.NewEndpoint(address, port, false, false, nil, nil, nil, 0)
	client, err := winrm.NewClient(endpoint, authentication.Username, authentication.Password)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, stdErr, _, err := client.RunPSWithContext(ctx, command.Command)
	if err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	pwshResult := cacao.Variable{Type: cacao.VariableTypeString, Name: resultVariableName, Value: result}
	pwshError := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var__", Value: stdErr}
	results := cacao.NewVariables(pwshResult, pwshError)

	return results, nil
}

func determineTargetAddress(target cacao.AgentTarget) (string, error) {

	if len(target.Address) == 0 {
		return "", errors.New("address map is empty or nil please provide a valid address map")
	}
	if len(target.Address["ipv4"]) > 0 {
		return target.Address["ipv4"][0], nil
	}
	if len(target.Address["ipv6"]) > 0 {
		return target.Address["ipv6"][0], nil
	}
	if len(target.Address["url"]) > 0 {
		return target.Address["url"][0], nil
	}

	return "", errors.New("unsupported target address type, not implemented")
}
