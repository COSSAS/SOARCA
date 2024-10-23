package thehive

import (
	"soarca/models/cacao"

	"soarca/internal/reporter/downstream_reporter/thehive/connector"

	"github.com/gofrs/uuid"
)

type TheHiveReporter struct {
	connector connector.ITheHiveConnector
}

func New(connector connector.ITheHiveConnector) *TheHiveReporter {
	return &TheHiveReporter{connector: connector}
}

func (theHiveReporter *TheHiveReporter) ConnectorTest() string {
	return theHiveReporter.connector.Hello()
}

func (theHiveReporter *TheHiveReporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	return nil
}

func (theHiveReporter *TheHiveReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error) error {
	return nil
}

func (theHiveReporter *TheHiveReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	return nil
}

func (theHiveReporter *TheHiveReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	return nil
}
