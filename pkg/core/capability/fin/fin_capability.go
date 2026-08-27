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
	timeUtil "soarca/pkg/utils/time"

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

// defaultStaleAfter is used when New is given a non-positive staleAfter. A
// Fin that hasn't been seen (via /poll) in this long is assumed to no
// longer be running - see the doc comment on checkCapableFin.
const defaultStaleAfter = 2 * time.Minute

// IJobQueue is the subset of queue.Queue this capability depends on.
type IJobQueue interface {
	Enqueue(ctx context.Context, job fin.Job) (fin.JobResult, error)
}

// IFinRepository is the subset of finrepository.IFinRepository this
// capability depends on (avoids an import of internal/database/fin from
// pkg/..., matching the same pattern used in pkg/api/fin).
type IFinRepository interface {
	List() ([]fin.Record, error)
}

// Capability implements capability.ICapability by enqueuing one fin.Job per
// step invocation. It is wired as the action executor's fallback
// capability (see action.Executor.SetFinFallback) rather than under a
// single static type in the capabilities map: Fin capability types are
// declared dynamically at Fin registration time, not known up front.
type Capability struct {
	queue      IJobQueue
	guid       guid.IGuid
	repository IFinRepository
	time       timeUtil.ITime
	// staleAfter is how long a registered Fin can go without a /poll
	// before checkCapableFin stops counting it as "live" for the purpose
	// of the fail-fast check. Not related to the job queue's own lease
	// expiry - a Fin can be well within this window and still take a
	// while to actually claim a given job.
	staleAfter time.Duration
}

// New builds a Fin fallback capability. repository is used to fail a step
// immediately (without enqueuing/waiting) when no live Fin could possibly
// claim it - see checkCapableFin. Passing a nil repository disables that
// check entirely (jobs are always enqueued and left to wait out the step's
// own timeout, matching this capability's original behavior). A
// non-positive staleAfter falls back to defaultStaleAfter.
func New(jobQueue IJobQueue, guidGenerator guid.IGuid, repository IFinRepository, clock timeUtil.ITime, staleAfter time.Duration) *Capability {
	if staleAfter <= 0 {
		staleAfter = defaultStaleAfter
	}
	return &Capability{
		queue:      jobQueue,
		guid:       guidGenerator,
		repository: repository,
		time:       clock,
		staleAfter: staleAfter,
	}
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

	if err := finCapability.checkCapableFin(commandContext.Agent.Type); err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}

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

// checkCapableFin fails fast (without ever touching the job queue) when no
// currently-registered Fin could possibly claim a job of capabilityType:
//   - zero registered Fins declare capabilityType at all -> fin.ErrNoCapableFin
//   - at least one does, but every one of them was last seen (via /poll)
//     more than staleAfter ago -> fin.ErrOnlyStaleCapableFins
//
// Otherwise (at least one live, matching Fin is registered) it returns nil
// and Execute proceeds to enqueue exactly as before - a live Fin may still
// legitimately take a while to get around to claiming the job, which
// remains bounded only by the step's own timeout.
//
// A nil repository (liveness checking disabled) or a repository error
// always returns nil - "fail open" to the pre-existing enqueue-and-wait
// behavior rather than blocking a step over an inability to check.
func (finCapability *Capability) checkCapableFin(capabilityType string) error {
	if finCapability.repository == nil {
		return nil
	}
	records, err := finCapability.repository.List()
	if err != nil {
		log.Warning("failed to list registered fins for capability type ", capabilityType, ": ", err)
		return nil
	}

	now := finCapability.time.Now()
	matched := false
	live := false
	staleFinIds := []string{}
	for _, record := range records {
		if !hasCapability(record, capabilityType) {
			continue
		}
		matched = true
		if now.Sub(record.LastSeen) <= finCapability.staleAfter {
			live = true
			continue
		}
		staleFinIds = append(staleFinIds, record.FinId)
	}

	if !matched {
		return fin.ErrNoCapableFin{CapabilityType: capabilityType}
	}
	if !live {
		return fin.ErrOnlyStaleCapableFins{
			CapabilityType: capabilityType,
			FinIds:         staleFinIds,
			StaleAfter:     finCapability.staleAfter,
		}
	}
	return nil
}

func hasCapability(record fin.Record, capabilityType string) bool {
	for _, cap := range record.Capabilities {
		if cap.Type == capabilityType {
			return true
		}
	}
	return false
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
