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

type ExecutionEntry struct {
	ExecutionId    uuid.UUID
	PlaybookId     string
	Started        time.Time
	Ended          time.Time
	StepResults    map[string]StepResult
	PlaybookResult error
	Status         Status
}

type StepResult struct {
	ExecutionId uuid.UUID
	StepId      string
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
