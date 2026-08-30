package manual

import (
	"context"
	"errors"
	"reflect"
	"soarca/internal/logger"
	"soarca/internal/workflow/capability"
	"soarca/internal/workflow/capability/manual/inbox"
	"soarca/pkg/cacao"
	manualModel "soarca/internal/manual/model"
	"soarca/internal/runs/model"
	"soarca/pkg/utils"
	"time"
)

var (
	component = reflect.TypeOf(ManualCapability{}).PkgPath()
	log       *logger.Log
)

const (
	manualResultVariableName = "__soarca_manual_result__"
	manualCapabilityName     = "soarca-manual"
)

func New(controller inbox.Dispatcher) ManualCapability {
	return ManualCapability{inbox: controller}
}

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type ManualCapability struct {
	inbox inbox.Dispatcher
}

func (manual *ManualCapability) GetType() string {
	return manualCapabilityName
}

func (manual *ManualCapability) Execute(
	metadata run.Metadata,
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
	channel := make(chan manualModel.Response)
	defer close(channel)

	err := manual.inbox.Queue(command, manualModel.Waiter{
		Channel:        channel,
		TimeoutContext: ctx,
	})

	if err != nil {
		return cacao.NewVariables(), err
	}

	result, err := manual.awaitUserInput(channel, ctx)

	// Deregister synchronously, before returning, so a subsequent
	// re-run of this step (e.g. the next iteration of a while-loop
	// body) can never race the async cleanup goroutine
	// Inbox.Queue also starts as a backstop. Keyed on this
	// invocation's StepRunId, so it never touches a different
	// invocation's still-pending entry, even one sharing the same StepId.
	if deregErr := manual.inbox.Deregister(metadata); deregErr != nil {
		log.Trace("manual command already deregistered: ", deregErr)
	}

	if err != nil {
		return cacao.NewVariables(), err
	}
	return result.Select(commandContext.Step.OutArgs), nil

}

func (manual *ManualCapability) awaitUserInput(channel chan manualModel.Response, ctx context.Context) (cacao.Variables, error) {

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
		fallback := utils.DefaultStepTimeout()
		log.Warning("timeout is not set or set to 0, fallback timeout of ", fallback, " is used to complete step")
		return fallback
	}
	return time.Duration(userTimeout) * time.Millisecond
}
