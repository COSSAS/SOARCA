package api

import (
	"errors"
	"soarca/models/cacao"
	cache_model "soarca/models/cache"
)

type Status uint8

// Reporter model adapted from https://github.com/cyentific-rni/workflow-status/blob/main/README.md

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
	Type            string
	ExecutionId     string
	PlaybookId      string
	Started         string
	Ended           string
	Status          string
	StatusText      string
	StepResults     map[string]StepExecutionReport
	Error           string
	RequestInterval int
}

type StepExecutionReport struct {
	ExecutionId string
	StepId      string
	Started     string
	Ended       string
	Status      string
	StatusText  string
	Error       string
	Variables   map[string]cacao.Variable
	// Make sure we can have a playbookID for playbook actions, and also
	// the execution ID for the invoked playbook
}

func CacheStatusEnum2String(status cache_model.Status) (string, error) {
	switch status {
	case cache_model.SuccessfullyExecuted:
		return SuccessfullyExecuted, nil
	case cache_model.Failed:
		return Failed, nil
	case cache_model.Ongoing:
		return Ongoing, nil
	case cache_model.ServerSideError:
		return ServerSideError, nil
	case cache_model.ClientSideError:
		return ClientSideError, nil
	case cache_model.TimeoutError:
		return TimeoutError, nil
	case cache_model.ExceptionConditionError:
		return ExceptionConditionError, nil
	default:
		return "", errors.New("unable to read execution information status")
	}
}
