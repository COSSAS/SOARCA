package manual

import (
	"context"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

// ################################################################################
// Data structures for native SOARCA manual command handling
// ################################################################################

type ManualResponseStatus string

const (
	Success ManualResponseStatus = "success"
	Failure ManualResponseStatus = "failure"
)

type InteractionStorageEntry struct {
	CommandInfo CommandInfo
	Channel     chan InteractionResponse
}

// Object passed by the manual capability to the Interaction module
type CommandInfo struct {
	Metadata         execution.Metadata
	Context          capability.Context
	OutArgsVariables cacao.Variables
}

// Deep copy for the command that the Interaction module notifies to the integrations
type InteractionIntegrationCommand struct {
	Metadata         execution.Metadata
	Context          capability.Context
	OutArgsVariables cacao.Variables
}

// Object returned to the Interaction object in fulfilment of a manual command
type InteractionResponse struct {
	Metadata         execution.Metadata
	ResponseStatus   ManualResponseStatus
	ResponseError    error
	OutArgsVariables cacao.Variables
}

type ManualCapabilityCommunication struct {
	TimeoutContext context.Context
	Channel        chan InteractionResponse
}
