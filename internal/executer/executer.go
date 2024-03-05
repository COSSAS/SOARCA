package executer

import (
	"errors"
	"reflect"
	"soarca/internal/capability"
	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

type IExecuter interface {
	Execute(metadata execution.Metadata,
		command cacao.Command,
		authentication cacao.AuthenticationInformation,
		target cacao.AgentTarget,
		variable cacao.Variables,
		module cacao.AgentTarget) (uuid.UUID, cacao.Variables, error)
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

func (executer *Executer) Execute(metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variable cacao.Variables,
	agent cacao.AgentTarget) (uuid.UUID, cacao.Variables, error) {

	if capability, ok := executer.capabilities[agent.Name]; ok {
		returnVariables, err := capability.Execute(metadata, command, authentication, target, variable)
		return metadata.ExecutionId, returnVariables, err
	} else {
		empty := cacao.Variables{}
		message := "executor is not available in soarca"
		err := errors.New(message)
		log.Error(message)
		return metadata.ExecutionId, empty, err
	}

}
