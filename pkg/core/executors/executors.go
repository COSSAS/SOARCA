package executors

import (
	"soarca/models/cacao"
	"soarca/models/execution"
)

type IPlaybookExecuter interface {
	Execute(execution.Metadata,
		cacao.Step,
		cacao.Variables) (cacao.Variables, error)
}
