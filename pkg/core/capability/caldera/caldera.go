package caldera

import (
	"reflect"

	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type calderaCapability struct{}

type Empty struct{}

const (
	calderaResult  = "__soarca_caldera_cmd_result__"
	calderaError   = "__soarca_caldera_cmd_error__"
	capabilityName = "soarca-caldera-cmd"
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

func (c *calderaCapability) Execute(
	metadata execution.Metadata,
	context capability.Context) (cacao.Variables, error) {

	log.Info("Successfully called execute on the caldera capability")
	return cacao.NewVariables(), nil
}
