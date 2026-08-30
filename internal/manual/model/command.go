package manual

import (
	"context"
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	"soarca/internal/runs/model"
)

// ################################################################################
// Data structures for native SOARCA manual command handling
// ################################################################################

type ManualResponseStatus string

const (
	ManualResponseSuccessStatus ManualResponseStatus = "success"
	ManualResponseFailureStatus ManualResponseStatus = "failure"
)

type PendingCommand struct {
	CommandInfo CommandInfo
	Channel     chan Response
}

// Object passed by the manual capability to the manual inbox
type CommandInfo struct {
	Metadata         run.Metadata
	Context          capability.Context
	OutArgsVariables cacao.Variables
}

// Deep copy for the command that the manual inbox notifies to the integrations
type Notification struct {
	Metadata         run.Metadata
	Context          capability.Context
	OutArgsVariables cacao.Variables
}

// Object returned to the manual inbox in fulfilment of a manual command
type Response struct {
	Metadata         run.Metadata
	ResponseStatus   ManualResponseStatus
	ResponseError    error
	OutArgsVariables cacao.Variables
}

type Waiter struct {
	TimeoutContext context.Context
	Channel        chan Response
}

// Errors #####################################################################
type ErrorPendingCommandNotFound struct {
	Err string
}

type ErrorNonMatchingOutArgs struct {
	Err string
}

func (e ErrorPendingCommandNotFound) Error() string {
	return e.Err
}

func (e ErrorNonMatchingOutArgs) Error() string {
	return e.Err
}
