package capability

import (
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

// ResolvedTarget pairs a step target with the authentication information
// resolved for it. Each target in a step may legitimately use different
// authentication, so auth is nested per-target rather than a single
// top-level field.
//
// This is also SOARCA's one canonical wire shape for "a target plus its
// resolved auth" wherever that needs to cross an API boundary (the Manual
// API's pending-command payload, the Fin protocol's job payload) — reused
// directly rather than each protocol defining its own equivalent-but-
// differently-shaped DTO.
type ResolvedTarget struct {
	Target         cacao.AgentTarget               `bson:"target" json:"target" validate:"required"`
	Authentication cacao.AuthenticationInformation `bson:"authentication,omitempty" json:"authentication,omitempty"`
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
	// Agent is the step's resolved agent_definitions entry. Built-in
	// capabilities don't need it (they're already selected by
	// agent.Type), but the Fin fallback capability does: it has no single
	// static type of its own, so it reads Agent.Type here to route the
	// job to the right pool of registered Fins.
	Agent cacao.AgentTarget
}

type ICapability interface {
	Execute(metadata execution.Metadata,
		context Context) (cacao.Variables, error)
	GetType() string
}
