package cases

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/integration/thehive/common/connector"
	thehive_models "soarca/pkg/integration/thehive/common/models"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporting/cases"
	"time"

	"github.com/google/uuid"
)

var (
	component = reflect.TypeOf(HiveCaseManager{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type HiveCaseManager struct {
	connector connector.ITheHiveConnector
}

func NewCaseManager(connector connector.ITheHiveConnector) *HiveCaseManager {
	return &HiveCaseManager{connector: connector}
}

func (manager *HiveCaseManager) AddToExistingOrCreateNew(meta execution.Metadata,
	playbook cacao.Playbook) cacao.Variable {

	//convert variables to observables
	observables := CreateHiveObservables(playbook.PlaybookVariables)

	// check if observables exists (will only find the first one)
	caseId := ""
	for value := range observables {
		if cases, err := manager.connector.FindCaseOfObservable(value); err != nil {
			continue // continue to the next value
		} else {
			if len(cases) > 0 {
				caseId = cases[0].ID
				exe := thehive_models.ExecutionMetadata{ExecutionId: meta.ExecutionId.String(),
					Playbook: playbook}
				if err := manager.connector.SetMapping(exe, meta.ExecutionId.String()); err != nil {
					log.Error(err)
				}
				break // break out of the loop
			}
		}
	}
	if caseId == "" {
		newCaseId, err := manager.connector.PostNewExecutionCase(
			thehive_models.ExecutionMetadata{
				ExecutionId: meta.ExecutionId.String(),
				Playbook:    playbook,
			},
			time.Now(),
		)
		if err != nil {
			log.Error(err)
		} else {
			caseId = newCaseId
		}

	}

	caseIdVar := cacao.Variable{Type: cacao.VariableTypeString,
		Name:        cases.SOARCA_PLAYBOOK_CASE_ID,
		Description: "SOARCA case id variable to be used in playbooks.",
		Value:       caseId,
		Constant:    true,
		External:    false}
	return caseIdVar
}

func (manager *HiveCaseManager) ConnectorTest() string {
	return manager.connector.Hello()
}

// Creates a new *case* in The Hive with related triggering metadata
func (manager *HiveCaseManager) ReportWorkflowStart(executionId uuid.UUID, playbook cacao.Playbook, at time.Time) error {
	log.Trace("TheHive casesreporting workflow start")
	// _, err := manager.connector.PostNewExecutionCase(
	// 	thehive_models.ExecutionMetadata{
	// 		ExecutionId: executionId.String(),
	// 		Playbook:    playbook,
	// 	},
	// 	at,
	// )
	return nil
}

// Marks case closure according to workflow execution. Also reports all variables, and data
func (manager *HiveCaseManager) ReportWorkflowEnd(executionId uuid.UUID, playbook cacao.Playbook, workflowErr error, at time.Time) error {
	log.Trace("TheHive casesreporting workflow end")
	_, err := manager.connector.UpdateEndExecutionCase(
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
func (manager *HiveCaseManager) ReportStepStart(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	log.Trace("TheHive casesreporting step start")
	_, err := manager.connector.UpdateStartStepTaskInCase(
		thehive_models.ExecutionMetadata{
			ExecutionId: executionId.String(),
			Step:        step,
		},
		at,
	)
	return err
}

// Populates event with step execution information
func (manager *HiveCaseManager) ReportStepEnd(executionId uuid.UUID, step cacao.Step, stepResults cacao.Variables, stepErr error, at time.Time) error {
	log.Trace("TheHive casesreporting step end")
	_, err := manager.connector.UpdateEndStepTaskInCase(
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
