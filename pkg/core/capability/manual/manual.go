package manual

import (
	"errors"
	"reflect"
	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/core/capability/manual/interaction"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	manualModel "soarca/pkg/models/manual"
	"time"
)

var (
	component = reflect.TypeOf(ManualCapability{}).PkgPath()
	log       *logger.Log
)

const (
	manualResultVariableName = "__soarca_manual_result__"
	manualCapabilityName     = "soarca-manual-http"
	fallbackTimeout          = time.Minute * 1
)

func New(controller interaction.ICapabilityInteraction,
	channel chan manualModel.InteractionResponse) ManualCapability {
	// channel := make(chan interaction.InteractionResponse)
	return ManualCapability{interaction: controller, channel: channel}
}

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ManualCapability struct {
	interaction interaction.ICapabilityInteraction
	channel     chan manualModel.InteractionResponse
}

func (manual *ManualCapability) GetType() string {
	return manualCapabilityName
}

func (manual *ManualCapability) Execute(
	metadata execution.Metadata,
	commandContext capability.Context) (cacao.Variables, error) {

	command := manualModel.InteractionCommand{Metadata: metadata, Context: commandContext}

	err := manual.interaction.Queue(command, manual.channel)
	if err != nil {
		return cacao.NewVariables(), err
	}

	result, err := manual.awaitUserInput(manual.getTimeoutValue(commandContext.Step.Timeout))
	if err != nil {
		return cacao.NewVariables(), err
	}
	return result.Select(commandContext.Step.OutArgs), nil

}

func (manual *ManualCapability) awaitUserInput(timeout time.Duration) (cacao.Variables, error) {
	timer := time.NewTimer(time.Duration(timeout))
	for {
		select {
		case <-timer.C:
			err := errors.New("manual response timeout, user responded not in time")
			log.Error(err)
			return cacao.NewVariables(), err
		case response := <-manual.channel:
			log.Trace("received response from api")
			cacaoVars := manual.copyOutArgsToVars(response.OutArgs.ResponseOutArgs)
			return cacaoVars, response.ResponseError

		}
	}
}

func (manual *ManualCapability) copyOutArgsToVars(outArgs manualModel.ManualOutArgs) cacao.Variables {
	vars := cacao.NewVariables()
	for name, outVar := range outArgs {

		vars[name] = cacao.Variable{
			Type:  outVar.Type,
			Name:  outVar.Name,
			Value: outVar.Value,
		}
	}
	return vars
}

func (manual *ManualCapability) getTimeoutValue(userTimeout int) time.Duration {
	if userTimeout == 0 {
		log.Warning("timeout is not set or set to 0 fallback timeout of 1 minute is used to complete step")
		return fallbackTimeout
	}
	return time.Duration(userTimeout) * time.Millisecond
}
