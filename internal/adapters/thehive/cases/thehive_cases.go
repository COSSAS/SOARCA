package cases

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/adapters/thehive/common/connector"
	thehive_models "soarca/internal/adapters/thehive/common/models"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"soarca/internal/reporting/cases"
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

func (manager *HiveCaseManager) AddToExistingOrCreateNew(meta run.Metadata,
	playbook cacao.Playbook) cacao.Variable {

	//convert variables to observables
	observables := CreateHiveObservables(playbook.PlaybookVariables)

	// check if observables exists (will only find the first one)
	caseId := ""
	for value := range observables {
		if cases, err := manager.connector.FindCaseOfObservable(value); err != nil {
			log.Error(err)
			continue // continue to the next value
		} else {
			if len(cases) > 0 {
				caseId = cases[0].ID
				exe := thehive_models.RunMetadata{RunId: meta.RunId.String(),
					Playbook: playbook}
				if err := manager.connector.SetMapping(exe, caseId); err != nil {
					log.Error(err)
				}
				break // break out of the loop
			}
		}
	}
	if caseId == "" {
		newCaseId, err := manager.connector.PostNewRunCase(
			thehive_models.RunMetadata{
				RunId: meta.RunId.String(),
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

	for id, observable := range observables {
		log.Info("addiing observable:", id, " to case: ", caseId)
		if err := manager.connector.CreateObservableInCase(caseId, observable); err != nil {
			log.Error(err)
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
func (manager *HiveCaseManager) ReportWorkflowStart(runId uuid.UUID, playbook cacao.Playbook, at time.Time) error {
	log.Trace("TheHive casesreporting workflow start")
	// _, err := manager.connector.PostNewRunCase(
	// 	thehive_models.RunMetadata{
	// 		RunId: runId.String(),
	// 		Playbook:    playbook,
	// 	},
	// 	at,
	// )
	return nil
}

// Marks case closure according to workflow run. Also reports all variables, and data
func (manager *HiveCaseManager) ReportWorkflowEnd(runId uuid.UUID, playbook cacao.Playbook, workflowErr error, at time.Time) error {
	log.Trace("TheHive casesreporting workflow end")
	_, err := manager.connector.UpdateEndRunCase(
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
func (manager *HiveCaseManager) ReportStepStart(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, at time.Time) error {
	log.Trace("TheHive casesreporting step start")
	_, err := manager.connector.UpdateStartStepTaskInCase(
		thehive_models.RunMetadata{
			RunId: metadata.RunId.String(),
			Step:        step,
		},
		at,
	)
	return err
}

// Populates event with step run information
func (manager *HiveCaseManager) ReportStepEnd(metadata run.Metadata, step cacao.Step, stepResults cacao.Variables, stepErr error, at time.Time) error {
	log.Trace("TheHive casesreporting step end")
	_, err := manager.connector.UpdateEndStepTaskInCase(
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
