package capability

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type Context struct {
	Command        cacao.Command
	Step           cacao.Step
	Authentication cacao.AuthenticationInformation
	Target         cacao.AgentTarget
	Variables      cacao.Variables
}

type ICapability interface {
	Execute(metadata execution.Metadata,
		context Context) (cacao.Variables, error)
	GetType() string
}
