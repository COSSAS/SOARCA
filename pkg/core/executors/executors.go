package executors

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type IPlaybookExecutor interface {
	Execute(execution.Metadata,
		cacao.Step,
		cacao.Variables) (cacao.Variables, error)
}

type Context struct {
	Step      cacao.Step
	Variables cacao.Variables
}

type IConditionExecutor interface {
	Execute(metadata execution.Metadata,
		stepContext Context) (string, bool, error)
}

type PlaybookStepMetadata struct {
	Step      cacao.Step
	Targets   map[string]cacao.AgentTarget
	Auth      map[string]cacao.AuthenticationInformation
	Agent     cacao.AgentTarget
	Variables cacao.Variables
}

type IActionExecutor interface {
	Execute(metadata execution.Metadata,
		step PlaybookStepMetadata) (cacao.Variables, error)
}
