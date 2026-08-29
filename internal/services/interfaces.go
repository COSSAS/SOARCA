package services

import (
	"context"

	"github.com/google/uuid"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/fin"
	"soarca/pkg/models/manual"
)

// ============================================================================
// OPERATIONAL SERVICES
// ============================================================================

// FinRegistry manages FIN registration and admin lifecycle.
//
// Dependencies: FinStore (only)
// Does NOT depend on: playbook execution, job leasing.
type FinRegistry interface {
	// RegisterFin registers a new FIN with the given capabilities.
	RegisterFin(ctx context.Context, req fin.RegisterRequest) (finID string, token string, err error)

	// UnregisterFin unregisters a FIN and cleans up its record.
	UnregisterFin(ctx context.Context, finToken string) error

	// ListFins returns all registered FINs (admin use).
	ListFins(ctx context.Context) (fins []fin.Record, err error)

	// GetFin retrieves a specific FIN record by ID (admin use).
	GetFin(ctx context.Context, finID string) (record fin.Record, err error)

	// DeleteFin removes a FIN record (admin use).
	DeleteFin(ctx context.Context, finID string) error

	// ValidateToken checks if a token is valid and returns the FIN ID.
	// Used by auth middleware to verify credentials.
	ValidateToken(ctx context.Context, finToken string) (finID string, err error)
}

// FinWorkService manages leased FIN work items.
//
// Dependencies: FinQueue, FinStore (only)
// Does NOT depend on: playbook execution, manual commands, HTTP.
//
// Lease semantics belong here, not in the execution runtime.
type FinWorkService interface {
	// PollJob polls for available work matching FIN's capabilities.
	// Long-polls until a job is available or context timeout/cancellation.
	PollJob(ctx context.Context, finToken string, pollReq fin.PollRequest) (job *fin.Job, err error)

	// SubmitJobResult submits the result of a claimed job.
	SubmitJobResult(ctx context.Context, finToken string, jobID uuid.UUID, result fin.JobResult) error

	// HeartbeatJob extends the lease on an in-flight job (status-ping).
	HeartbeatJob(ctx context.Context, finToken string, jobID uuid.UUID) error
}

// ManualInbox manages manual step resolution during playbook execution.
// Depends on: Interaction controller
// Does NOT depend on: FIN leasing or claim semantics.
//
// Responsibility: tracking pending manual commands, allowing operators to
// view pending steps, and providing responses for outstanding manual work.
type ManualInbox interface {
	// ListPendingCommands returns all pending manual steps across all executions.
	ListPendingCommands() (commands []manual.CommandInfo, err error)

	// GetPendingCommand retrieves a specific pending manual step.
	GetPendingCommand(metadata execution.Metadata) (command manual.CommandInfo, err error)

	// ContinuePendingCommand resolves a pending manual step with the operator's response.
	ContinuePendingCommand(response manual.InteractionResponse) error
}

// ============================================================================
// APPLICATION SERVICES (thin orchestrators)
// ============================================================================

// PlaybookService provides CRUD operations over the playbook repository.
// Depends on: PlaybookStore
//
// Responsibility: playbook lifecycle (create, read, update, delete, list).
// Future: could add versioning, validation, domain rules.
type PlaybookService interface {
	// ListPlaybooks returns all stored playbooks.
	ListPlaybooks(ctx context.Context) (playbooks []cacao.Playbook, err error)

	// GetPlaybook retrieves a playbook by ID.
	GetPlaybook(ctx context.Context, playbookID string) (playbook *cacao.Playbook, err error)

	// CreatePlaybook stores a new playbook.
	CreatePlaybook(ctx context.Context, playbook *cacao.Playbook) error

	// UpdatePlaybook replaces an existing playbook.
	UpdatePlaybook(ctx context.Context, playbookID string, playbook *cacao.Playbook) error

	// DeletePlaybook removes a playbook.
	DeletePlaybook(ctx context.Context, playbookID string) error

	// ListPlaybookMetas returns metadata for all playbooks (efficient list).
	ListPlaybookMetas(ctx context.Context) (metas []api.PlaybookMeta, err error)
}

// ============================================================================
// DEPENDENCY NOTES
// ============================================================================

/*
Dependency Flow (what depends on what):

  executions.Runner (core kernel)
    └─ owns: engine (decomposer factory), PlaybookStore, Cache

  FinRegistry (independent)
    └─ owns: FinStore

  FinWorkService (leased work)
    └─ owns: FinQueue, FinStore

  ManualInbox
    └─ depends on: Interaction controller

  PlaybookService (CRUD adapter)
    └─ depends on: PlaybookStore

HTTP Handlers (thin transport)
  └─ depend on: these services
  └─ do NOT depend on: stores, queues, runtime directly
*/
