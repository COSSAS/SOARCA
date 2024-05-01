package report

import (
	"soarca/internal/guid"
	"soarca/models/cacao"
	"time"
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
)

type ExecutionEntry struct {
	ExecutionId    guid.IGuid
	PlaybookId     string
	Started        time.Time
	Ended          time.Time
	StepResults    map[string]StepResult
	PlaybookResult error
	Status         Status
}

type StepResult struct {
	ExecutionId guid.IGuid
	StepId      string
	Started     time.Time
	Ended       time.Time
	// Make sure we can have a playbookID for playbook actions, and also
	// the execution ID for the invoked playbook
	Variables cacao.Variables
	Status    Status
	Errors    []error
}
