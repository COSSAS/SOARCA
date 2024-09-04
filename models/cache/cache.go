package cache

import (
	"soarca/models/cacao"
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

type ExecutionEntry struct {
	ExecutionId uuid.UUID
	Name        string
	Description string
	PlaybookId  string
	Started     time.Time
	Ended       time.Time
	StepResults map[string]StepResult
	Error       error
	Status      Status
}

type StepResult struct {
	ExecutionId uuid.UUID
	StepId      string
	Name        string
	Description string
	Started     time.Time
	Ended       time.Time
	// Make sure we can have a playbookID for playbook actions, and also
	// the execution ID for the invoked playbook
	CommandsB64 []string
	Variables   cacao.Variables
	Status      Status
	Error       error
	IsAutomated bool
}
