package downstream_reporter

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) error
	ReportStep(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error
}
