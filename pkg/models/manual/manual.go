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

// Object stored in interaction storage and provided back from the API
type InteractionCommandData struct {
	Type          string            `bson:"type" json:"type" validate:"required" example:"execution-status"` // The type of this content
	ExecutionId   string            `bson:"execution_id" json:"execution_id" validate:"required"`            // The id of the execution
	PlaybookId    string            `bson:"playbook_id" json:"playbook_id" validate:"required"`              // The id of the CACAO playbook executed by the execution
	StepId        string            `bson:"step_id" json:"step_id" validate:"required"`                      // The id of the step executed by the execution
	Description   string            `bson:"description" json:"description" validate:"required"`              // The description from the workflow step
	Command       string            `bson:"command" json:"command" validate:"required"`                      // The command for the agent either command
	CommandBase64 string            `bson:"commandb64,omitempty" json:"commandb64,omitempty"`                // The command in b64 if present
	Target        cacao.AgentTarget `bson:"targets" json:"targets" validate:"required"`                      // Map of cacao agent-target with the target(s) of this command
	OutVariables  cacao.Variables   `bson:"out_args" json:"out_args" validate:"required"`                    // Map of cacao variables handled in the step out args with current values and definitions
}

type InteractionStorageEntry struct {
	CommandData InteractionCommandData
	Channel     chan InteractionResponse
}

// Object passed by the manual capability to the Interaction module
type InteractionCommand struct {
	Metadata execution.Metadata
	Context  capability.Context
}

// The variables returned to SOARCA from a manual interaction
// Alike to the cacao.Variable, but with only type name and value required
type ManualOutArg struct {
	Type        string `bson:"type,omitempty" json:"type,omitempty" example:"string"`                    // Type of the variable should be OASIS  variable-type-ov
	Name        string `bson:"name" json:"name" validate:"required" example:"__example_string__"`        // The name of the variable in the style __variable_name__
	Description string `bson:"description,omitempty" json:"description,omitempty" example:"some string"` // A description of the variable
	Value       string `bson:"value" json:"value" validate:"required" example:"this is a value"`         // The value of the that the variable will evaluate to
	Constant    bool   `bson:"constant,omitempty" json:"constant,omitempty" example:"false"`             // Indicate if it's a constant
	External    bool   `bson:"external,omitempty" json:"external,omitempty" example:"false"`             // Indicate if it's external
}

// The collection of out args mapped per variable name
type ManualOutArgs map[string]ManualOutArg

// The object posted on the manual API Continue() payload
type ManualOutArgUpdatePayload struct {
	Type            string        `bson:"type" json:"type" validate:"required" example:"string"`          // The type of this content
	ExecutionId     string        `bson:"execution_id" json:"execution_id" validate:"required"`           // The id of the execution
	PlaybookId      string        `bson:"playbook_id" json:"playbook_id" validate:"required"`             // The id of the CACAO playbook executed by the execution
	StepId          string        `bson:"step_id" json:"step_id" validate:"required"`                     // The id of the step executed by the execution
	ResponseStatus  bool          `bson:"response_status" json:"response_status" validate:"required"`     // Can be either success or failure
	ResponseOutArgs ManualOutArgs `bson:"response_out_args" json:"response_out_args" validate:"required"` // Map of cacao variables expressed as ManualOutArgs, handled in the step out args, with current values and definitions
}

// The object that the Interaction module presents back to the manual capability
type InteractionResponse struct {
	ResponseError error
	Payload       cacao.Variables
}

type ManualCapabilityCommunication struct {
	Channel        chan InteractionResponse
	TimeoutContext context.Context
}

// ################################################################################
// Data structures for integrations manual command handling
// ################################################################################

// As manual interaction integrations are called on go routines, this
// duplications prevents inconsistencies on the objects by forcing
// full deep copies of the objects.

// The command that the Interactin module notifies to the integrations
type InteractionIntegrationCommand struct {
	Metadata execution.Metadata
	Context  capability.Context
}

// The payload that an integration puts back on a channel for the Interaction module
// to receive
type InteractionIntegrationResponse struct {
	ResponseError error
	Payload       ManualOutArgUpdatePayload
}
