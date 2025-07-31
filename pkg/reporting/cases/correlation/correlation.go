package correlation

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporting/cases/incident"

	"soarca/pkg/utils/guid"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ICorrelation interface {
	Correlate(execution.Metadata, cacao.Variables) (string, error)
}

type ICorrelationIntegration interface {
	GetExternalCaseId(string) string
}

type Cases map[uuid.UUID]incident.ICase

type Correlation struct {
	cases        Cases
	guid         guid.IGuid
	integrations ICorrelationIntegration
}

func New(guid guid.IGuid) Correlation {
	return Correlation{cases: Cases{}, guid: guid}
}

func (correlation *Correlation) GetCaseId(meta execution.Metadata, playbook cacao.Playbook) string {
	internalId, err := correlation.Correlate(meta, playbook.PlaybookVariables)
	if err != nil {
		return ""
	}

	externalId := correlation.integrations.GetExternalCaseId(internalId)
	return externalId

}

func (correlation *Correlation) Correlate(meta execution.Metadata, variables cacao.Variables) (string, error) {
	for _, item := range correlation.cases {
		for _, variable := range variables {
			if part, err := item.CheckObservableInCase(variable); err != nil {
				return "", err
			} else if part {
				log.Debug(part)
				correlation.addExecutionToCase(item, meta, variable)
				return item.GetId().String(), nil
			}
		}
	}
	log.Info("not in case creating a new case")
	id := correlation.createCase(meta, variables)
	return id.String(), nil
}

// func (correlation *Correlation) addEddxecutionToCase(meta execution.Metadata, variables cacao.Variables) error {
// 	return nil
// }

// func (Correlation *Correlation) searchForCase() {

// }

func (correlation *Correlation) addExecutionToCase(item incident.ICase, meta execution.Metadata, variable cacao.Variable) {
	added, err := correlation.cases[item.GetId()].AddIfNotInCase(meta, variable)
	if err != nil {
		log.Error(err)
	}
	log.Info(added)
}

func (correlation *Correlation) createCase(meta execution.Metadata, variables cacao.Variables) uuid.UUID {
	thisCase := incident.New(correlation.guid.NewV7(), meta, variables)
	correlation.cases[thisCase.GetId()] = thisCase
	return thisCase.GetId()
}
