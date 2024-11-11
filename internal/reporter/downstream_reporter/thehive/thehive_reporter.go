package thehive

import (
	"soarca/models/cacao"
	"time"

	"github.com/google/uuid"
)

type TheHiveReporter struct {
	connector ITheHiveConnector
}

func NewReporter(connector ITheHiveConnector) *TheHiveReporter {
	return &TheHiveReporter{connector: connector}
}

func (theHiveReporter *TheHiveReporter) ConnectorTest() string {
	return theHiveReporter.connector.Hello()
}

// Creates a new *case* in The Hive with related triggering metadata
func (theHiveReporter *TheHiveReporter) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) error {
	_, err := theHiveReporter.connector.PostNewExecutionCase(executionId.String(), playbook, at)
	return err
}

// Marks case closure according to workflow execution. Also reports all variables, and data
func (theHiveReporter *TheHiveReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowErr error, at time.Time) error {
	_, err := theHiveReporter.connector.UpdateEndExecutionCase(executionId.String(), playbook.PlaybookVariables, workflowErr, at)
	return err
}

// Adds *event* to case
func (theHiveReporter *TheHiveReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	_, err := theHiveReporter.connector.UpdateStartStepTaskInCase(executionId.String(), step, stepResults, at)
	return err
}

// Populates event with step execution information
func (theHiveReporter *TheHiveReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepErr error, at time.Time) error {
	_, err := theHiveReporter.connector.UpdateEndStepTaskInCase(executionId.String(), step, stepResults, stepErr, at)
	return err
}
