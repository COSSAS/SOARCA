package caldera

import (
	"reflect"

	"soarca/logger"
	"soarca/models/cacao"
	"soarca/models/execution"
)

type calderaCapability struct{}

type Empty struct{}

const (
	calderaResult  = "__soarca_caldera_result__"
	calderaError   = "__soarca_caldera_error__"
	capabilityName = "soarca-caldera"
)

var (
	component = reflect.TypeOf(Empty{}).PkgPath()
	log       *logger.Log
)

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New() *calderaCapability {
	return &calderaCapability{}
}

func (capability *calderaCapability) GetType() string {
	return capabilityName
}

func (capability *calderaCapability) Execute(
	metadata execution.Metadata,
	command cacao.Command,
	authentication cacao.AuthenticationInformation,
	target cacao.AgentTarget,
	variables cacao.Variables) (cacao.Variables, error) {

	log.Info("Successfully called execute on the caldera capability")
	return cacao.NewVariables(), nil
}
