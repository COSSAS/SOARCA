package thehive

import (
	thehive_models "soarca/pkg/integration/thehive/common/models"
	"soarca/pkg/models/cacao"
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
	log.Trace("TheHive reporter reporting workflow start")
	_, err := theHiveReporter.connector.PostNewExecutionCase(
		thehive_models.ExecutionMetadata{
			ExecutionId: executionId.String(),
			Playbook:    playbook,
		},
		at,
	)
	return err
}

// Marks case closure according to workflow execution. Also reports all variables, and data
func (theHiveReporter *TheHiveReporter) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowErr error, at time.Time) error {
	log.Trace("TheHive reporter reporting workflow end")
	_, err := theHiveReporter.connector.UpdateEndExecutionCase(
		thehive_models.ExecutionMetadata{
			ExecutionId:  executionId.String(),
			Variables:    playbook.PlaybookVariables,
			ExecutionErr: workflowErr,
		},
		at,
	)
	return err
}

// Adds *event* to case
func (theHiveReporter *TheHiveReporter) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	log.Trace("TheHive reporter reporting step start")
	_, err := theHiveReporter.connector.UpdateStartStepTaskInCase(
		thehive_models.ExecutionMetadata{
			ExecutionId: executionId.String(),
			Step:        step,
		},
		at,
	)
	return err
}

// Populates event with step execution information
func (theHiveReporter *TheHiveReporter) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepErr error, at time.Time) error {
	log.Trace("TheHive reporter reporting step end")
	_, err := theHiveReporter.connector.UpdateEndStepTaskInCase(
		thehive_models.ExecutionMetadata{
			ExecutionId:  executionId.String(),
			Step:         step,
			Variables:    stepResults,
			ExecutionErr: stepErr,
		},
		at,
	)
	return err
}
