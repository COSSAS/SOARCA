package fin

import (
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/fin/protocol"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	finModel "soarca/pkg/models/fin"
)

type FinCapability struct {
	finProtocol protocol.IFinProtocol
}

var component = reflect.TypeOf(FinCapability{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

func New(finProtocol protocol.IFinProtocol) *FinCapability {
	return &FinCapability{finProtocol: finProtocol}
}

func (FinCapability *FinCapability) GetType() string {
	return "soarca-fin"
}

func (finCapability *FinCapability) Execute(
	metadata execution.Metadata,
	context capability.Context) (cacao.Variables, error) {

	finCommand := finModel.NewCommand()
	finCommand.CommandSubstructure.Command = context.Command.Command
	finCommand.CommandSubstructure.Authentication = context.Authentication
	finCommand.CommandSubstructure.Variables = context.Variables
	finCommand.CommandSubstructure.Context.ExecutionId = metadata.ExecutionId.String()
	finCommand.CommandSubstructure.Context.PlaybookId = metadata.PlaybookId
	finCommand.CommandSubstructure.Context.StepId = metadata.StepId

	log.Trace("created command ", finCommand)
	return finCapability.finProtocol.SendCommand(finCommand)
}
