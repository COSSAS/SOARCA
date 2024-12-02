package capability

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type ICapability interface {
	Execute(metadata execution.Metadata,
		command cacao.Command,
		authentication cacao.AuthenticationInformation,
		target cacao.AgentTarget,
		variables cacao.Variables) (cacao.Variables, error)
	GetType() string
}
