package capability

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type ICapability interface {
	Execute(executionId uuid.UUID, command cacao.Command, variables cacao.Variables, target cacao.Target, OnCompletionCallback func(vars cacao.Variables))
}
