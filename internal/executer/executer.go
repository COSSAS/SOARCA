package executer

import (
	"errors"
	"reflect"
	"soarca/internal/capability"
	"soarca/logger"
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

type IExecuter interface {
	Execute(executionId uuid.UUID,
		command cacao.Command,
		authentication cacao.AuthenticationInformation,
		target cacao.AgentTarget,
		variable map[string]cacao.Variables,
		module cacao.AgentTarget) (uuid.UUID, map[string]cacao.Variables, error)
}

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(capabilities map[string]capability.ICapability) *Executer {
	var instance = Executer{}
	instance.capabilities = capabilities
	return &instance
}

type Executer struct {
	capabilities map[string]capability.ICapability
}

func (executer *Executer) Execute(executionId uuid.UUID,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variable map[string]cacao.Variables,
	agent cacao.AgentTarget) (uuid.UUID, map[string]cacao.Variables, error) {

	if capability, ok := executer.capabilities[agent.Name]; ok {
		returnVariables, err := capability.Execute(executionId, command, authentication, target, variable)
		return executionId, returnVariables, err
	} else {
		empty := map[string]cacao.Variables{}
		message := "executor is not available in soarca"
		err := errors.New(message)
		log.Error(message)
		return executionId, empty, err
	}

}
