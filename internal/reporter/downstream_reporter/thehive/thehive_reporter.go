package thehive

import (
	"soarca/models/cacao"

	"github.com/gofrs/uuid"
)

type TheHiveReporter struct {
	connector ITheHiveConnector
}

func New(connector ITheHiveConnector) *TheHiveReporter {
	return &TheHiveReporter{connector: connector}
}

func (thehiveReporter *TheHiveReporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	return nil
}

func (thehiveReporter *TheHiveReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error) error {
	return nil
}

func (thehiveReporter *TheHiveReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	return nil
}

func (thehiveReporter *TheHiveReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	return nil
}
