package cases

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporting/cases/observable"

	"github.com/gofrs/uuid"
)

const SOARCA_PLAYBOOK_CASE_ID = "__SOARCA_CASE_ID__"

type Cases map[uuid.UUID]ICase

type ICasesManager interface {
	AddToExistingOrCreateNew(execution.Metadata, cacao.Playbook) cacao.Variable
}

type CaseIds struct {
	Ids map[string]string
}

func (caseIds *CaseIds) Add(backendName string, value string) {
	if caseIds.Ids == nil {
		caseIds.Ids = map[string]string{}
	}
	caseIds.Ids[backendName] = value
}

func (caseIds *CaseIds) Remove(backendName string) {
	delete(caseIds.Ids, backendName)
}

type ICase interface {
	CheckObservableInCase(cacao.Variable) (bool, error)
	AddIfNotInCase(execution.Metadata, cacao.Variable) (bool, error)
	GetId() uuid.UUID
	GetExternalId() string
}

type IExternalCaseManager interface {
	CheckIfForCase(observable.Observable)
}
