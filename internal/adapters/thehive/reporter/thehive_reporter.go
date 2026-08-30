package thehive

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/adapters/thehive/common/connector"
	thehive_models "soarca/internal/adapters/thehive/common/models"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"time"

	"github.com/google/uuid"
)

var (
	component = reflect.TypeOf(TheHiveReporter{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type TheHiveReporter struct {
	connector connector.ITheHiveConnector
}

func NewReporter(connector connector.ITheHiveConnector) *TheHiveReporter {
	return &TheHiveReporter{connector: connector}
}

func (theHiveReporter *TheHiveReporter) ConnectorTest() string {
	return theHiveReporter.connector.Hello()
}

// Creates a new *case* in The Hive with related triggering metadata
func (theHiveReporter *TheHiveReporter) ReportWorkflowStart(runId uuid.UUID, playbook cacao.Playbook, at time.Time) error {
	log.Trace("TheHive reporter reporting workflow start")
	_, err := theHiveReporter.connector.PostNewRunCase(
		thehive_models.RunMetadata{
			RunId: runId.String(),
			Playbook:    playbook,
		},
		at,
	)
	return err
}

// Marks case closure according to workflow run. Also reports all variables, and data
func (theHiveReporter *TheHiveReporter) ReportWorkflowEnd(runId uuid.UUID, playbook cacao.Playbook, workflowErr error, at time.Time) error {
	log.Trace("TheHive reporter reporting workflow end")
	_, err := theHiveReporter.connector.UpdateEndRunCase(
		thehive_models.RunMetadata{
			RunId:  runId.String(),
			Variables:    playbook.PlaybookVariables,
			RunErr: workflowErr,
		},
		at,
	)
	return err
}

// Adds *event* to case
func (theHiveReporter *TheHiveReporter) ReportStepStart(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	log.Trace("TheHive reporter reporting step start")
	_, err := theHiveReporter.connector.UpdateStartStepTaskInCase(
		thehive_models.RunMetadata{
			RunId: metadata.RunId.String(),
			Step:        step,
		},
		at,
	)
	return err
}

// Populates event with step run information
func (theHiveReporter *TheHiveReporter) ReportStepEnd(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, stepErr error, at time.Time) error {
	log.Trace("TheHive reporter reporting step end")
	_, err := theHiveReporter.connector.UpdateEndStepTaskInCase(
		thehive_models.RunMetadata{
			RunId:  metadata.RunId.String(),
			Step:         step,
			Variables:    stepResults,
			RunErr: stepErr,
		},
		at,
	)
	return err
}
