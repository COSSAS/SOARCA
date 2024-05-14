package api

import (
	"errors"
	"fmt"
	"soarca/models/cacao"
	cache_model "soarca/models/cache"
)

type Status uint8

// Reporter model adapted from https://github.com/cyentific-rni/workflow-status/blob/main/README.md

const (
	ReportLevelPlaybook = "playbook"
	ReportLevelStep     = "step"

	SuccessfullyExecuted    = "successfully_executed"
	Failed                  = "failed"
	Ongoing                 = "ongoing"
	ServerSideError         = "server_side_error"
	ClientSideError         = "client_side_error"
	TimeoutError            = "timeout_error"
	ExceptionConditionError = "exception_condition_error"
	AwaitUserInput          = "await_user_input"

	SuccessfullyExecutedText    = "%s execution completed successfully"
	FailedText                  = "something went wrong in the execution of this %s"
	OngoingText                 = "this %s is currently being executed"
	ServerSideErrorText         = "there was a server-side problem with the execution of this %s"
	ClientSideErrorText         = "something in the data provided for this %s raised an issue"
	TimeoutErrorText            = "the execution of this %s timed out"
	ExceptionConditionErrorText = "the execution of this %s raised a playbook exception"
	AwaitUserInputText          = "waiting for users to provide input for the %s execution"
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
	ExecutionId        string
	StepId             string
	Started            string
	Ended              string
	Status             string
	StatusText         string
	ExecutedBy         string
	CommandsB64        []string
	Error              string
	Variables          map[string]cacao.Variable
	AutomatedExecution string
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
	case cache_model.AwaitUserInput:
		return AwaitUserInput, nil
	default:
		return "", errors.New("unable to read execution information status")
	}
}

func GetCacheStatusText(status string, level string) (string, error) {
	if level != ReportLevelPlaybook && level != ReportLevelStep {
		return "", errors.New("invalid reporting level provided. use either 'playbook' or 'step'")
	}
	switch status {
	case SuccessfullyExecuted:
		return fmt.Sprintf(SuccessfullyExecutedText, level), nil
	case Failed:
		return fmt.Sprintf(FailedText, level), nil
	case Ongoing:
		return fmt.Sprintf(OngoingText, level), nil
	case ServerSideError:
		return fmt.Sprintf(ServerSideErrorText, level), nil
	case ClientSideError:
		return fmt.Sprintf(ClientSideErrorText, level), nil
	case TimeoutError:
		return fmt.Sprintf(TimeoutErrorText, level), nil
	case ExceptionConditionError:
		return fmt.Sprintf(ExceptionConditionErrorText, level), nil
	case AwaitUserInput:
		return fmt.Sprintf(AwaitUserInputText, level), nil
	default:
		return "", errors.New("unable to read execution information status")
	}
}
