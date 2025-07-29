package correlation

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type ICorrelation interface {
	GetCaseId(execution.Metadata, cacao.Playbook) string
}
