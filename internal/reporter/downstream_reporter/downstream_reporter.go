package downstream_reporter

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	ReportWorkflow(workflowEntry WorkflowEntry) error
	ReportStep(stepEntry StepEntry) error
}

type WorkflowEntry struct {
	// TODO Change to context
	ExecutionId uuid.UUID
	Playbook    cacao.Playbook
}

type StepEntry struct {
	ExecutionContext execution.Metadata
	Variables        cacao.Variables
	Error            error
}
