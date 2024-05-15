package api

import (
	"errors"
	"fmt"
	"soarca/models/cacao"
	cache_model "soarca/models/cache"
	"time"
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
	Type            string                         `bson:"type" json:"type"`
	ExecutionId     string                         `bson:"execution_id" json:"execution_id"`
	PlaybookId      string                         `bson:"playbook_id" json:"playbook_id"`
	Started         time.Time                      `bson:"started" json:"started"`
	Ended           time.Time                      `bson:"ended" json:"ended"`
	Status          string                         `bson:"status" json:"status"`
	StatusText      string                         `bson:"status_text" json:"status_text"`
	StepResults     map[string]StepExecutionReport `bson:"step_results" json:"step_results"`
	RequestInterval int                            `bson:"request_interval" json:"request_interval"`
}

type StepExecutionReport struct {
	ExecutionId        string                    `bson:"execution_id" json:"execution_id"`
	StepId             string                    `bson:"step_id" json:"step_id"`
	Started            time.Time                 `bson:"started" json:"started"`
	Ended              time.Time                 `bson:"ended" json:"ended"`
	Status             string                    `bson:"status" json:"status"`
	StatusText         string                    `bson:"status_text" json:"status_text"`
	ExecutedBy         string                    `bson:"executed_by" json:"executed_by"`
	CommandsB64        []string                  `bson:"commands_b64" json:"commands_b64"`
	Variables          map[string]cacao.Variable `bson:"variables" json:"variables"`
	AutomatedExecution bool                      `bson:"automated_execution" json:"automated_execution"`
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
