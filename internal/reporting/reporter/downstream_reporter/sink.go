package downstream_reporter

import (
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"time"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	ReportWorkflowStart(runId uuid.UUID, playbook cacao.Playbook, at time.Time) error
	ReportWorkflowEnd(runId uuid.UUID, playbook cacao.Playbook, err error, at time.Time) error

	// metadata.StepRunId identifies this specific invocation of
	// metadata.StepId - see run.Metadata for why StepId alone is not
	// sufficient.
	ReportStepStart(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, at time.Time) error
	ReportStepEnd(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, err error, at time.Time) error
}
