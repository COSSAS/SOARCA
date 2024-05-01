package downstream_reporter

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error
	ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error) error

	ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error
	ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error
}
