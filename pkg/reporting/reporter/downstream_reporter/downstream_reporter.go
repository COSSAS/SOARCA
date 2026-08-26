package downstream_reporter

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"time"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) error
	ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error, at time.Time) error

	// metadata.StepExecutionId identifies this specific invocation of
	// metadata.StepId - see execution.Metadata for why StepId alone is not
	// sufficient.
	ReportStepStart(metadata execution.Metadata, step cacao.Step, stepResults cacao.Variables, at time.Time) error
	ReportStepEnd(metadata execution.Metadata, step cacao.Step, stepResults cacao.Variables, err error, at time.Time) error
}
