package api

import (
	"soarca/models/cacao"
)

type Status uint8

const (
	SuccessfullyExecuted    = "successfully_executed"
	Failed                  = "failed"
	Ongoing                 = "ongoing"
	ServerSideError         = "server_side_error"
	ClientSideError         = "client_side_error"
	TimeoutError            = "timeout_error"
	ExceptionConditionError = "exception_condition_error"
)

type PlaybookExecutionReport struct {
	ExecutionId     string
	PlaybookId      string
	Started         string
	Ended           string
	Status          string
	StatusText      string
	StepResults     map[string]StepExecutionReport
	Errors          []string
	requestInterval int
}

type StepExecutionReport struct {
	ExecutionId string
	PlaybookId  string
	Started     string
	Ended       string
	Status      string
	StatusText  string
	Errors      []string
	Variables   cacao.Variables
	// Make sure we can have a playbookID for playbook actions, and also
	// the execution ID for the invoked playbook
}
