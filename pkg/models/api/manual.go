package api

import (
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/manual"
)

// Object interfaced to users storing info about pending manual commands
// TODO: change to manualcommandinfo
type InteractionCommandData struct {
	Type            string                      `bson:"type" json:"type" validate:"required" example:"run-status"`  // The type of this content
	RunId           string                      `bson:"run_id" json:"run_id" validate:"required"`                   // The id of the run
	PlaybookId      string                      `bson:"playbook_id" json:"playbook_id" validate:"required"`         // The id of the CACAO playbook executed by the run
	StepId          string                      `bson:"step_id" json:"step_id" validate:"required"`                 // The id of the step executed by the run
	StepRunId       string                      `bson:"step_run_id" json:"step_run_id" validate:"required"`         // The id of this specific step invocation. Distinguishes concurrent/repeated pending commands that share the same StepId (e.g. overlapping loop iterations)
	Commands        []ManualCommand             `bson:"commands" json:"commands" validate:"required"`                    // All commands of the step, in order. A manual step is a single unit of work resolved by one response, but may list multiple commands/instructions
	Targets         []capability.ResolvedTarget `bson:"targets" json:"targets" validate:"required"`                      // All targets of the step, in order, together with their resolved authentication information (needed by a human operator to perform the step manually)
	OutVariables    cacao.Variables             `bson:"out_args" json:"out_args" validate:"required"`                    // Map of cacao variables handled in the step out args with current values and definitions
}

// One command of a (possibly multi-command) manual step
type ManualCommand struct {
	Description     string `bson:"description" json:"description" validate:"required"` // The description from the workflow step
	Command         string `bson:"command" json:"command" validate:"required"`         // The command for the agent, either plain or base64
	CommandIsBase64 bool   `bson:"commandb64,omitempty" json:"commandb64,omitempty"`   // Indicates if the command is in b64
}

// The object posted on the manual API Continue() payload
type ManualOutArgsUpdatePayload struct {
	Type           string                      `bson:"type" json:"type" validate:"required" example:"string"`      // The type of this content
	ResponseStatus manual.ManualResponseStatus `bson:"response_status" json:"response_status" validate:"required"` // Indicates status of command

	ResponseOutArgs cacao.Variables `bson:"response_out_args" json:"response_out_args" validate:"required"` // Map of cacao variables storing the out args value, handled in the step out args, with current values and definitions
}
