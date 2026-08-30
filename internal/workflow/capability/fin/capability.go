package fin

import (
	"context"
	"errors"
	"reflect"
	"time"

	"soarca/internal/logger"
	"soarca/internal/store"
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	"soarca/pkg/fins/protocol"
	"soarca/internal/runs/model"
	"soarca/pkg/utils"
	timeUtil "soarca/pkg/utils/time"

	"soarca/pkg/utils/guid"
)

type Empty struct{}

var log *logger.Log

func init() {
	log = logger.Logger(reflect.TypeOf(Empty{}).PkgPath(), logger.Info, "", logger.Json)
}

type IJobQueue interface {
	Enqueue(ctx context.Context, job fin.Job) (fin.JobResult, error)
}

type Capability struct {
	queue      IJobQueue
	guid       guid.IGuid
	store      storage.FinStore
	time       timeUtil.ITime
	staleAfter time.Duration
}

// Dependencies groups the dependencies needed to construct a Fin capability.
type Dependencies struct {
	Queue      IJobQueue
	GUID       guid.IGuid
	Store      storage.FinStore
	Time       timeUtil.ITime
	StaleAfter time.Duration
}

func New(deps Dependencies) *Capability {
	return &Capability{
		queue:      deps.Queue,
		guid:       deps.GUID,
		store:      deps.Store,
		time:       deps.Time,
		staleAfter: deps.StaleAfter,
	}
}

func (finCapability *Capability) GetType() string { return "" }

func (finCapability *Capability) Execute(metadata run.Metadata, commandContext capability.Context) (cacao.Variables, error) {
	log.Trace(metadata.RunId)
	if err := finCapability.checkCapableFin(commandContext.Agent.Type); err != nil {
		log.Error(err)
		return cacao.NewVariables(), err
	}
	timeout := leaseDuration(commandContext.Step.Timeout)
	job := fin.Job{
		JobId:                 finCapability.guid.New(),
		RunId:                 metadata.RunId,
		PlaybookId:            metadata.PlaybookId,
		StepId:                metadata.StepId,
		StepRunId:             metadata.StepRunId,
		CapabilityType:        commandContext.Agent.Type,
		LeaseExpiresInSeconds: int(timeout.Seconds()),
		Step:                  fin.StepInfo{Name: commandContext.Step.Name, Description: commandContext.Step.Description, Timeout: commandContext.Step.Timeout, Delay: commandContext.Step.Delay},
		Commands:              toFinCommands(commandContext.Commands),
		Targets:               commandContext.Targets,
		Variables:             commandContext.Variables,
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

func (finCapability *Capability) checkCapableFin(capabilityType string) error {
	if finCapability.store == nil {
		return nil
	}
	records, err := finCapability.store.List(context.Background())
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
		return fin.ErrOnlyStaleCapableFins{CapabilityType: capabilityType, FinIds: staleFinIds, StaleAfter: finCapability.staleAfter}
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
		fallback := utils.DefaultStepTimeout()
		log.Warning("timeout is not set or set to 0, fallback timeout of ", fallback, " is used to complete step")
		return fallback
	}
	return time.Duration(stepTimeoutMillis) * time.Millisecond
}

func toFinCommands(commands []cacao.Command) []fin.Command {
	converted := make([]fin.Command, 0, len(commands))
	for _, command := range commands {
		converted = append(converted, fin.Command{Type: command.Type, Command: command.Command})
	}
	return converted
}
