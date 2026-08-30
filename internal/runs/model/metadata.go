package run

import (
	"github.com/google/uuid"
)

type Metadata struct {
	RunId      uuid.UUID
	PlaybookId string
	StepId     string
	// StepRunId uniquely identifies one *invocation* of StepId within
	// RunId. StepId alone is not enough: while-loops and cyclic
	// on_completion graphs can visit the same StepId more than once within a
	// single run, and (once StepTypeParallel is implemented) the same
	// StepId could even be dispatched concurrently. A fresh StepRunId
	// is minted for every such invocation, so per-invocation state (pending
	// pending manual commands, run-state entries, ...) can be keyed
	// without colliding across re-runs of the same step.
	StepRunId uuid.UUID
}
