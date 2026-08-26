// Package fin implements capability.ICapability for the new HTTP/JSON
// pull-based Fin protocol (see docs/adr/FIN-WEBHOOK-PROTOCOL-PROPOSAL.md):
// it turns a step invocation into one fin.Job, enqueues it on the shared
// job queue, and blocks until a claiming Fin submits a result or the step's
// own timeout elapses.
package fin

import (
	"context"
	"errors"
	"reflect"
	"time"

	"soarca/internal/logger"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/fin"
	"soarca/pkg/utils/guid"
)

type Empty struct{}

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// fallbackLeaseDuration is used when a step does not declare a positive
// Timeout, mirroring manual.ManualCapability's own fallback (see
// pkg/core/capability/manual/manual.go).
const fallbackLeaseDuration = time.Minute

// IJobQueue is the subset of queue.Queue this capability depends on.
type IJobQueue interface {
	Enqueue(ctx context.Context, job fin.Job) (fin.JobResult, error)
}

// Capability implements capability.ICapability by enqueuing one fin.Job per
// step invocation. It is wired as the action executor's fallback
// capability (see action.Executor.SetFinFallback) rather than under a
// single static type in the capabilities map: Fin capability types are
// declared dynamically at Fin registration time, not known up front.
type Capability struct {
	queue IJobQueue
	guid  guid.IGuid
}

func New(jobQueue IJobQueue, guidGenerator guid.IGuid) *Capability {
	return &Capability{queue: jobQueue, guid: guidGenerator}
}

// GetType always returns "" - Capability has no single static type of its
// own (see the doc comment above); it is never looked up by GetType() the
// way built-in capabilities are.
func (finCapability *Capability) GetType() string {
	return ""
}

func (finCapability *Capability) Execute(metadata execution.Metadata,
	commandContext capability.Context) (cacao.Variables, error) {

	log.Trace(metadata.ExecutionId)

	timeout := leaseDuration(commandContext.Step.Timeout)
	job := fin.Job{
		JobId:                 finCapability.guid.New(),
		ExecutionId:           metadata.ExecutionId,
		PlaybookId:            metadata.PlaybookId,
		StepId:                metadata.StepId,
		StepExecutionId:       metadata.StepExecutionId,
		CapabilityType:        commandContext.Agent.Type,
		LeaseExpiresInSeconds: int(timeout.Seconds()),
		Step: fin.StepInfo{
			Name:        commandContext.Step.Name,
			Description: commandContext.Step.Description,
			Timeout:     commandContext.Step.Timeout,
			Delay:       commandContext.Step.Delay,
		},
		Commands:  toFinCommands(commandContext.Commands),
		Targets:   commandContext.Targets,
		Variables: commandContext.Variables,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result, err := finCapability.queue.Enqueue(ctx, job)
	if err != nil {
		log.Error("fin job ", job.JobId.String(), " did not complete: ", err)
		return cacao.NewVariables(), err
	}

	if result.State == fin.JobStateFailure {
		resultErr := errors.New(result.Error)
		if result.Error == "" {
			resultErr = errors.New("fin job reported failure without an error message")
		}
		log.Error(resultErr)
		return result.Variables, resultErr
	}

	return result.Variables, nil
}

func leaseDuration(stepTimeoutMillis int) time.Duration {
	if stepTimeoutMillis <= 0 {
		log.Warning("timeout is not set or set to 0, fallback timeout of 1 minute is used to complete step")
		return fallbackLeaseDuration
	}
	return time.Duration(stepTimeoutMillis) * time.Millisecond
}

func toFinCommands(commands []cacao.Command) []fin.Command {
	converted := make([]fin.Command, 0, len(commands))
	for _, command := range commands {
		converted = append(converted, fin.Command{
			Type:       command.Type,
			Command:    command.Command,
			Content:    command.Content,
			ContentB64: command.ContentB64,
			Headers:    map[string][]string(command.Headers),
		})
	}
	return converted
}
