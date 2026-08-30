package runstate

import (
	"soarca/pkg/cacao"
	"time"

	"github.com/google/uuid"
)

type Status uint8

const (
	SuccessfullyExecuted Status = iota
	Failed
	Ongoing
	ServerSideError
	ClientSideError
	TimeoutError
	ExceptionConditionError
	AwaitUserInput
)

func (status Status) String() string {
	return [...]string{
		"successfully_executed",
		"failed",
		"ongoing",
		"server_side_error",
		"client_side_error",
		"timeout_error",
		"exception_condition_error",
		"await_user_input",
	}[status]
}

type RunEntry struct {
	RunId       uuid.UUID
	Name        string
	Description string
	PlaybookId  string
	Started     time.Time
	Ended       time.Time
	// StepResults is keyed by StepRunId (not StepId): the same StepId
	// can be invoked more than once within one run (while-loops,
	// cyclic on_completion graphs), and each such invocation gets its own
	// entry here instead of overwriting/being rejected.
	StepResults map[string]StepResult
	Error       error
	Status      Status
}

type StepResult struct {
	RunId       uuid.UUID
	StepId      string
	StepRunId   uuid.UUID
	Name        string
	Description string
	Started     time.Time
	Ended       time.Time
	// Make sure we can have a playbookID for playbook actions, and also
	// the run ID for the invoked playbook
	CommandsB64 []string
	Variables   cacao.Variables
	Status      Status
	Error       error
	IsAutomated bool
}
