package execution

import (
	"github.com/google/uuid"
)

type Metadata struct {
	ExecutionId uuid.UUID
	PlaybookId  string
	StepId      string
	// StepExecutionId uniquely identifies one *invocation* of StepId within
	// ExecutionId. StepId alone is not enough: while-loops and cyclic
	// on_completion graphs can visit the same StepId more than once within a
	// single execution, and (once StepTypeParallel is implemented) the same
	// StepId could even be dispatched concurrently. A fresh StepExecutionId
	// is minted for every such invocation, so per-invocation state (pending
	// manual interactions, reporting cache entries, ...) can be keyed
	// without colliding across re-executions of the same step.
	StepExecutionId uuid.UUID
}
