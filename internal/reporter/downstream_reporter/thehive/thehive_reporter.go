package thehive

import (
	"soarca/models/cacao"

	"soarca/internal/reporter/downstream_reporter/thehive/connector"

	"github.com/google/uuid"
)

type TheHiveReporter struct {
	connector connector.ITheHiveConnector
}

func New(connector connector.ITheHiveConnector) *TheHiveReporter {
	return &TheHiveReporter{connector: connector}
}

// TODO: add structures to handle Execution ID to TheHive IDs mapping

func (theHiveReporter *TheHiveReporter) ConnectorTest() string {
	return theHiveReporter.connector.Hello()
}

// Creates a new *case* in The Hive with related triggering metadata
func (theHiveReporter *TheHiveReporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	_, err := theHiveReporter.connector.PostNewCase(executionId.String(), playbook)
	return err
}

// Marks case closure according to workflow execution. Also reports all variables, and data
func (theHiveReporter *TheHiveReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, err error) error {
	return nil
}

// Adds *event* to case
func (theHiveReporter *TheHiveReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	return nil
}

// Populates event with step execution information
func (theHiveReporter *TheHiveReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, err error) error {
	return nil
}
