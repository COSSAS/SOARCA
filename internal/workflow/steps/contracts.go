package executors

import (
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
)

type IPlaybookExecuter interface {
	Execute(run.Metadata,
		cacao.Step,
		cacao.Variables) (cacao.Variables, error)
}

type Context struct {
	Step      cacao.Step
	Variables cacao.Variables
}

type IConditionExecuter interface {
	Execute(metadata run.Metadata,
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
	Execute(metadata run.Metadata,
		step PlaybookStepMetadata) (cacao.Variables, error)
}
