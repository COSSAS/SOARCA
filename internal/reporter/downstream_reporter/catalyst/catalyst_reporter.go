package catalyst

import (
	"soarca/models/cacao"

	"github.com/gofrs/uuid"
)

type CatalystReporter struct {
	connector ICatalystConnector
}

func New(connector ICatalystConnector) *CatalystReporter {
	return &CatalystReporter{connector: connector}
}

func (catalystReporter *CatalystReporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	return nil
}

func (catalystReporter *CatalystReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error) error {
	return nil
}

func (catalystReporter *CatalystReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	return nil
}

func (catalystReporter *CatalystReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	return nil
}
