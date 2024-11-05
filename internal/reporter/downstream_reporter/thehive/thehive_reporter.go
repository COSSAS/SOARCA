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

func (theHiveReporter *TheHiveReporter) ConnectorTest() string {
	return theHiveReporter.connector.Hello()
}

// Creates a new *case* in The Hive with related triggering metadata
func (theHiveReporter *TheHiveReporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook) error {
	_, err := theHiveReporter.connector.PostNewExecutionCase(executionId.String(), playbook)
	return err
}

// Marks case closure according to workflow execution. Also reports all variables, and data
func (theHiveReporter *TheHiveReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowErr error) error {
	_, err := theHiveReporter.connector.UpdateEndExecutionCase(executionId.String(), playbook.PlaybookVariables, workflowErr)
	return err
}

// Adds *event* to case
func (theHiveReporter *TheHiveReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables) error {
	_, err := theHiveReporter.connector.UpdateStartStepTaskInCase(executionId.String(), step, stepResults)
	return err
}

// Populates event with step execution information
func (theHiveReporter *TheHiveReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepErr error) error {
	_, err := theHiveReporter.connector.UpdateEndStepTaskInCase(executionId.String(), step, stepResults, stepErr)
	return err
}
