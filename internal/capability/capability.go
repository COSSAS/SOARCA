package capability

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type ICapability interface {
	Execute(executionId uuid.UUID,
		command cacao.Command,
		authentication cacao.AuthenticationInformation,
		target cacao.AgentTarget,
		variables map[string]cacao.Variable) (map[string]cacao.Variable, error)
	GetType() string
}
