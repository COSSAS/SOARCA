package observable

import (
	"errors"
	"reflect"
	"slices"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"

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
type Executions map[uuid.UUID]MatchReason

var ip_v4_source_type_ids = []string{"__SOURCE_IPV4__", "__SRC_IPV4__"}
var ip_v4_destination_ids = []string{"__DESTINATION_IPV4__", "__DEST_IPV4__"}
var mac_source_ids = []string{"__SOURCE_MAC__", "__SRC_MAC__"}
var mac_destination_ids = []string{"__DESTINATION_MAC__", "__DEST_MAC__"}
var port_source_type_ids = []string{"__SOURCE_PORT__", "__SRC_PORT__"}
var port_destination_type_ids = []string{"__DESTINATION_PORT__", "__DEST_PORT__"}

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
	Initial         MatchReason = "Initial observable"
	SameObservable  MatchReason = "Observable is seen again"
	LateralMovement MatchReason = "Lateral movement, source of this observation is destination of previous observation"
	NoMatch         MatchReason = "Not a match"
)

type Observable struct {
	Type       ObservableType
	Name       string
	Value      string
	Executions Executions
}

func deductObservableType(name string) ObservableType {
	if slices.Contains(ip_v4_source_type_ids, name) {
		return SourceAddressIpv4
	} else if slices.Contains(ip_v4_destination_ids, name) {
		return DestinationAddressIpv4
	} else if slices.Contains(mac_source_ids, name) {
		return SourceMac
	} else if slices.Contains(mac_destination_ids, name) {
		return DestinationMac
	} else if slices.Contains(port_source_type_ids, name) {
		return SourcePort
	} else if slices.Contains(port_destination_type_ids, name) {
		return DestinationPort
	}
	return Other
}

func CreateObservable(variable cacao.Variable) (Observable, error) {

	if variable.Name == "" {
		err := errors.New("variable is empty")
		log.Error(err)
		return Observable{}, err
	}

	return Observable{Type: deductObservableType(variable.Name),
		Name:       variable.Name,
		Value:      variable.Value,
		Executions: Executions{}}, nil
}

func (observable *Observable) AddExecutionToObservable(id uuid.UUID, match MatchReason) {
	if _, ok := observable.Executions[id]; ok {
		return
	}
	observable.Executions[id] = match

}

func (observable *Observable) Match(other Observable, exectuion uuid.UUID) error {
	if len(other.Executions) > 0 {
		log.Warning("the other executions will be discarded when updating the base")
	}

	reason := determineMatchReason(*observable, other)
	if reason == NoMatch {
		return errors.New("no match made with observables")
	}
	observable.Executions[exectuion] = reason
	log.Debug("Matched observable: ", observable)
	return nil
}

func determineMatchReason(var1 Observable, var2 Observable) MatchReason {
	// update to STIX engine in the future
	if var1.Type == var2.Type {
		return SameObservable
	} else if var1.Type == SourceAddressIpv4 && var2.Type == DestinationAddressIpv4 {
		return LateralMovement
	} else if var1.Type == DestinationAddressIpv4 && var2.Type == SourceAddressIpv4 {
		return LateralMovement
	}

	return NoMatch
}
