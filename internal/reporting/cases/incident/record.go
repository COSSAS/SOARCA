package incident

import (
	"errors"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
	"soarca/internal/reporting/cases/observable"

	"time"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ICase interface {
	CheckObservableInCase(cacao.Variable) (bool, error)
	AddIfNotInCase(run.Metadata, cacao.Variable) (bool, error)
	GetId() uuid.UUID
	GetExternalId() string
	GetObservables() Observables
}

type StepResult struct {
	StepId      string
	Description string
	State       string
}

type Result struct {
	Meta        run.Metadata
	StepResults map[string]StepResult
}

type Runs map[uuid.UUID]run.Metadata
type Observables map[string]observable.Observable

type Case struct {
	Id            uuid.UUID
	ExternalId    string
	Observables   map[string]observable.Observable
	FirstObserved time.Time
	Runs    Runs
	// time          itime.ITime
	// IsClosed      bool
}

func New(guid uuid.UUID,
	meta run.Metadata,
	variables cacao.Variables) ICase {
	observables := Observables{}
	thisCase := Case{Id: guid,
		Observables: observables}
	for _, variable := range variables {
		_, err := thisCase.AddIfNotInCase(meta, variable)
		log.Error(err)
	}

	return &thisCase
}

func (cas *Case) GetId() uuid.UUID {
	return cas.Id
}

func (cas *Case) GetExternalId() string {
	return cas.ExternalId
}

func (cas *Case) CheckObservableInCase(variable cacao.Variable) (bool, error) {

	// delta := time.Since(cas.FirstObserved)
	// if delta > time.Duration(time.Minute*10) {
	// 	return false, errors.New("not in time range")
	// }
	if _, ok := cas.Observables[variable.Value]; ok {
		return true, nil
	}
	return false, nil
}

func (cas *Case) AddIfNotInCase(meta run.Metadata,
	variable cacao.Variable) (bool, error) {
	if _, ok := cas.Observables[variable.Value]; !ok {
		observed, err := observable.CreateObservable(variable)
		if err != nil {
			return false, err
		}
		observed.AddRunToObservable(meta.RunId, observable.Initial)
		cas.Observables[variable.Value] = observed

		return true, nil
	}
	return false, nil
}

func (cas *Case) AddRunAndStepData(variable cacao.Variable, step cacao.Step, meta run.Metadata) {
	newObserved, err := observable.CreateObservable(variable)
	if err != nil {
		log.Error(err)
		return
	}
	if observed, ok := cas.Observables[variable.Value]; ok {

		err := observed.Match(newObserved, meta.RunId)
		if err != nil {
			log.Error(err)
			return
		}
		cas.Observables[variable.Value] = observed
		cas.Runs[meta.RunId] = meta
	}
}

func (cas *Case) GetObservable(value string) (observable.Observable, error) {
	observable, ok := cas.Observables[value]
	if !ok {
		return observable, errors.New("observable is not in case")
	}
	return observable, nil
}

func (cas *Case) GetObservables() Observables {
	return cas.Observables
}
