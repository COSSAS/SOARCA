package downstream_reporter

import (
	"soarca/models/cacao"

	"github.com/google/uuid"
)

// TODO:
// We should understand better how to handle variables at execution level, not reporting, so that only relevant data is sent to reporting
type IDownStreamReporter interface {
	ReportWorkflow(executionId uuid.UUID, playbook cacao.Playbook) error
	ReportStep(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error
}
