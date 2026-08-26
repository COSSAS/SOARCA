package capability

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

// ResolvedTarget pairs a step target with the authentication information
// resolved for it. Each target in a step may legitimately use different
// authentication, so auth is nested per-target rather than a single
// top-level field.
type ResolvedTarget struct {
	Target         cacao.AgentTarget
	Authentication cacao.AuthenticationInformation
}

// Context carries everything a capability needs to execute a single step.
// Commands and Targets are both plain arrays (0, 1, or many) — a capability
// receives the full step in one call and is responsible for iterating over
// its own commands/targets (and any target-level parallelism) internally,
// rather than being invoked once per (command, target) pair.
type Context struct {
	Commands  []cacao.Command
	Targets   []ResolvedTarget
	Step      cacao.Step
	Variables cacao.Variables
}

type ICapability interface {
	Execute(metadata execution.Metadata,
		context Context) (cacao.Variables, error)
	GetType() string
}
