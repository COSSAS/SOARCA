package capability

import (
	"soarca/models/cacao"
	"soarca/models/execution"
)

type ICapability interface {
	Execute(metadata execution.Metadata,
		command cacao.Command,
		authentication cacao.AuthenticationInformation,
		target cacao.AgentTarget,
		variables cacao.Variables,
		inputVariableKeys []string,
		outputVariablesKeys []string) (cacao.Variables, error)
	GetType() string
}
