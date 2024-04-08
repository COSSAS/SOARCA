package downstream_reporter

import (
	"soarca/models/cacao"
	"soarca/models/execution"
)

// Change custom structs to explicit arguments
type IDownStreamReporter interface {
	ReportWorkflow(workflowEntry WorkflowEntry) error
	ReportStep(stepEntry StepEntry) error
}

type WorkflowEntry struct {
	// TODO Change to context
	// Only execution ID and playbook
	ExecutionContext execution.Metadata
	Playbook         cacao.Playbook
}

type StepEntry struct {
	// Only execution ID, Step (contains ID, stepvars, in args, out args), results (of step execution: returnvariables), error
	// We should understand better how to handle variables at execution level, not reporting, so that only relevant data is sent to reporting
	ExecutionContext execution.Metadata
	Variables        cacao.Variables
	Error            error
}
