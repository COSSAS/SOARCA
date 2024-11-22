package downstream_reporter

import (
	"soarca/pkg/models/cacao"
	"time"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) error
	ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error, at time.Time) error

	ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, at time.Time) error
	ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error, at time.Time) error
}
