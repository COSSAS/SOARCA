package manual

import "soarca/pkg/models/cacao"

type ManualCommandData struct {
	Type            string                       `bson:"type" json:"type" validate:"required" example:"execution-status"` // The type of this content
	ExecutionId     string                       `bson:"execution_id" json:"execution_id" validate:"required"`            // The id of the execution
	PlaybookId      string                       `bson:"playbook_id" json:"playbook_id" validate:"required"`              // The id of the CACAO playbook executed by the execution
	StepId          string                       `bson:"step_id" json:"step_id" validate:"required"`                      // The id of the step executed by the execution
	Description     string                       `bson:"description" json:"description" validate:"required"`              // The description from the workflow step
	Command         string                       `bson:"command" json:"command" validate:"required"`                      // The command for the agent either command
	CommandIsBase64 bool                         `bson:"command_is_base64" json:"command_is_base64" validate:"required"`  // Indicate the command is in base 64
	Targets         map[string]cacao.AgentTarget `bson:"targets" json:"targets" validate:"required"`                      // Map of cacao agent-target with the target(s) of this command
	OutArgs         cacao.Variables              `bson:"out_args" json:"out_args" validate:"required"`                    // Map of cacao variables handled in the step out args with current values and definitions
}

type ManualOutArg struct {
	Name        string `bson:"name" json:"name" validate:"required" example:"__example_string__"`        // The name of the variable in the style __variable_name__
	Value       string `bson:"value" json:"value" validate:"required" example:"this is a value"`         // The value of the that the variable will evaluate to
	Type        string `bson:"type,omitempty" json:"type,omitempty" example:"string"`                    // Type of the variable should be OASIS  variable-type-ov
	Description string `bson:"description,omitempty" json:"description,omitempty" example:"some string"` // A description of the variable
	Constant    bool   `bson:"constant,omitempty" json:"constant,omitempty" example:"false"`             // Indicate if it's a constant
	External    bool   `bson:"external,omitempty" json:"external,omitempty" example:"false"`             // Indicate if it's external
}

type ManualRequestPayload struct {
	Type            string                  `bson:"type" json:"type" validate:"required" example:"string"`          // The type of this content
	ExecutionId     string                  `bson:"execution_id" json:"execution_id" validate:"required"`           // The id of the execution
	PlaybookId      string                  `bson:"playbook_id" json:"playbook_id" validate:"required"`             // The id of the CACAO playbook executed by the execution
	StepId          string                  `bson:"step_id" json:"step_id" validate:"required"`                     // The id of the step executed by the execution
	ResponseStatus  bool                    `bson:"response_status" json:"response_status" validate:"required"`     // Can be either success or failure
	ResponseOutArgs map[string]ManualOutArg `bson:"response_out_args" json:"response_out_args" validate:"required"` // Map of cacao variables expressed as ManualOutArgs, handled in the step out args, with current values and definitions
}
