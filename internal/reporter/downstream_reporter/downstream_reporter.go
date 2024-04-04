package downstream_reporter

import (
	"soarca/models/cacao"
	"soarca/models/execution"

	"github.com/google/uuid"
)

type IDownStreamReporter interface {
	// ReportResults should be called immediately upon creations to link results to execution
	// Then:
	//	- Substitute with new report entries, OR
	// 	- Edit report entries
	ReportWorkflow(workflowEntry WorkflowEntry) error
	ReportStep(stepEntry StepEntry) error
}

type WorkflowEntry struct {
	ExecutionId uuid.UUID
	Playbook    cacao.Playbook
}

type StepEntry struct {
	ExecutionContext execution.Metadata
	Variables        cacao.Variables
	Error            error
}
