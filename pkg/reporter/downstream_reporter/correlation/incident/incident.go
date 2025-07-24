package incident

import (
	"errors"
	"reflect"
	"slices"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"time"

	"github.com/google/uuid"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ObservableType int
type MatchReason string

const (
	SourceAddressIpv4 ObservableType = iota
	SourcePort
	DestinationAddressIpv4
	DestinationPort
	SourceMac
	DestinationMac
	Other
)

const (
	SameObservable  MatchReason = "Observable is seen again"
	LateralMovement MatchReason = "Lateral movement, source of this observation is destination of previous observation"
)

type Observable struct {
	Type  ObservableType
	Name  string
	Value string
}

type Observables map[string]Observable

type ICase interface {
	CheckObservableInCase(cacao.Variable) bool
	AddIfNotInCase(cacao.Variable)
}

var ip_v4_source_type_ids = []string{"__SOURCE_IPV4__"}
var ip_v4_destination_ids = []string{"__DESTINATION_IPV4__"}

func deductObservableType(name string) ObservableType {
	if slices.Contains(ip_v4_source_type_ids, name) {
		return SourceAddressIpv4
	} else if slices.Contains(ip_v4_destination_ids, name) {
		return DestinationAddressIpv4
	}
	return Other
}

func createObservable(variable cacao.Variable) (Observable, error) {

	if variable.Name == "" {
		err := errors.New("variable is empty")
		log.Error(err)
		return Observable{}, err
	}

	return Observable{Type: deductObservableType(variable.Name),
		Name:  variable.Name,
		Value: variable.Value}, nil
}

type Executions map[uuid.UUID]execution.Metadata

type Case struct {
	Id            uuid.UUID
	Executions    Executions
	Observables   map[string]Observable
	FirstObserved time.Time
	// IsClosed      bool
	// time          itime.ITime
}

func New(meta execution.Metadata, variables cacao.Variables) Case {
	exe := Executions{meta.ExecutionId: meta}
	observables := Observables{}
	for _, variable := range variables {
		if observable, err := createObservable(variable); err == nil {
			observables[variable.Value] = observable
		}
	}
	thisCase := Case{Id: uuid.New(), Executions: exe,
		Observables: observables}
	return thisCase
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

func (cas *Case) AddIfNotInCase(variable cacao.Variable) (bool, error) {
	if _, ok := cas.Observables[variable.Value]; !ok {
		observable, err := createObservable(variable)
		if err != nil {
			return false, err
		}
		cas.Observables[variable.Value] = observable
		return true, nil
	}
	return false, nil
}
