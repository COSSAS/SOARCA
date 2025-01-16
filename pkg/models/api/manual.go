package api

import "soarca/pkg/models/cacao"

// Object interfaced to users storing info about pending manual commands
// TODO: change to manualcommandinfo
type InteractionCommandData struct {
	Type          string            `bson:"type" json:"type" validate:"required" example:"execution-status"` // The type of this content
	ExecutionId   string            `bson:"execution_id" json:"execution_id" validate:"required"`            // The id of the execution
	PlaybookId    string            `bson:"playbook_id" json:"playbook_id" validate:"required"`              // The id of the CACAO playbook executed by the execution
	StepId        string            `bson:"step_id" json:"step_id" validate:"required"`                      // The id of the step executed by the execution
	Description   string            `bson:"description" json:"description" validate:"required"`              // The description from the workflow step
	Command       string            `bson:"command" json:"command" validate:"required"`                      // The command for the agent either command
	CommandBase64 string            `bson:"commandb64,omitempty" json:"commandb64,omitempty"`                // The command in b64 if present
	Target        cacao.AgentTarget `bson:"target" json:"target" validate:"required"`                        // Map of cacao agent-target with the target(s) of this command
	OutVariables  cacao.Variables   `bson:"out_args" json:"out_args" validate:"required"`                    // Map of cacao variables handled in the step out args with current values and definitions
}

// The object posted on the manual API Continue() payload
type ManualOutArgsUpdatePayload struct {
	Type            string          `bson:"type" json:"type" validate:"required" example:"string"`          // The type of this content
	ExecutionId     string          `bson:"execution_id" json:"execution_id" validate:"required"`           // The id of the execution
	PlaybookId      string          `bson:"playbook_id" json:"playbook_id" validate:"required"`             // The id of the CACAO playbook executed by the execution
	StepId          string          `bson:"step_id" json:"step_id" validate:"required"`                     // The id of the step executed by the execution
	ResponseStatus  string          `bson:"response_status" json:"response_status" validate:"required"`     // true indicates success, all good. false indicates request not met.
	ResponseOutArgs cacao.Variables `bson:"response_out_args" json:"response_out_args" validate:"required"` // Map of cacao variables storing the out args value, handled in the step out args, with current values and definitions
}
