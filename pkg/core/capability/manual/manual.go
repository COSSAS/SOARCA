package manual

import (
	"context"
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
	manualCapabilityName     = "soarca-manual"
	fallbackTimeout          = time.Minute * 1
)

func New(controller interaction.ICapabilityInteraction) ManualCapability {
	return ManualCapability{interaction: controller}
}

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ManualCapability struct {
	interaction interaction.ICapabilityInteraction
}

func (manual *ManualCapability) GetType() string {
	return manualCapabilityName
}

func (manual *ManualCapability) Execute(
	metadata execution.Metadata,
	commandContext capability.Context) (cacao.Variables, error) {

	command := manualModel.CommandInfo{
		Metadata:         metadata,
		Context:          commandContext,
		OutArgsVariables: commandContext.Variables.Select(commandContext.Step.OutArgs),
	}

	timeout := manual.getTimeoutValue(commandContext.Step.Timeout)
	log.Trace("timeout is set to: ", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// One channel per Execute() invocation. Async manual capability Execute() invocations can thus
	// use separate channels per each specific manual command, preventing manual returned args interfering
	channel := make(chan manualModel.InteractionResponse)
	defer close(channel)

	err := manual.interaction.Queue(command, manualModel.ManualCapabilityCommunication{
		Channel:        channel,
		TimeoutContext: ctx,
	})

	if err != nil {
		return cacao.NewVariables(), err
	}

	result, err := manual.awaitUserInput(channel, ctx)
	if err != nil {
		return cacao.NewVariables(), err
	}
	return result.Select(commandContext.Step.OutArgs), nil

}

func (manual *ManualCapability) awaitUserInput(channel chan manualModel.InteractionResponse, ctx context.Context) (cacao.Variables, error) {

	for {
		select {
		case <-ctx.Done():
			err := errors.New("manual response timed-out, no response received on time")
			log.Error(err)
			return cacao.NewVariables(), err
		case response := <-channel:
			log.Trace("received response from api")
			cacaoVars := response.OutArgsVariables
			return cacaoVars, response.ResponseError

		}
	}
}

func (manual *ManualCapability) getTimeoutValue(userTimeout int) time.Duration {
	if userTimeout == 0 {
		log.Warning("timeout is not set or set to 0 fallback timeout of 1 minute is used to complete step")
		return fallbackTimeout
	}
	return time.Duration(userTimeout) * time.Millisecond
}
