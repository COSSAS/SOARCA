package services

import (
	"context"

	"github.com/google/uuid"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/cache"
	"soarca/pkg/models/fin"
	"soarca/pkg/models/manual"
)

// ============================================================================
// CORE EXECUTION RUNTIME
// ============================================================================

// ExecutionRuntime is the core playbook orchestration service.
// It owns playbook execution, capability dispatch, and state management.
//
// Dependencies: PlaybookStore, Decomposer, Cache, Interaction
// Does NOT depend on: HTTP handlers, FIN operations, agent management
//
// This is the kernel that enables SOARCA to run playbooks in any context
// (HTTP, gRPC, CLI, embedded SDK, etc.).
type ExecutionRuntime interface {
	// StartExecution begins a playbook execution with the given variables.
	// Initiates async execution and returns the execution ID immediately.
	StartExecution(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (executionID uuid.UUID, err error)

	// ResumeManualStep resumes a paused manual step with the given response.
	// Used by manual interaction handlers to continue blocked executions.
	ResumeManualStep(ctx context.Context, execID uuid.UUID, stepExecID uuid.UUID, response manual.InteractionResponse) error

	// GetExecutionStatus retrieves the current status of an execution.
	GetExecutionStatus(ctx context.Context, execID uuid.UUID) (status cache.ExecutionEntry, err error)
}

// ============================================================================
// OPERATIONAL SERVICES (do not depend on ExecutionRuntime)
// ============================================================================

// FinRegistry manages FIN registration and admin lifecycle.
//
// Dependencies: FinStore (only)
// Does NOT depend on: ExecutionRuntime, playbook concepts, job leasing.
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
// Does NOT depend on: ExecutionRuntime, manual commands, HTTP.
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
// Depends on: ExecutionRuntime, Interaction controller
// Does NOT depend on: FIN leasing or claim semantics.
//
// Responsibility: tracking pending manual commands, allowing operators to
// view pending steps, and providing responses that resume execution.
type ManualInbox interface {
	// ListPendingCommands returns all pending manual steps across all executions.
	ListPendingCommands(ctx context.Context) (commands []manual.CommandInfo, err error)

	// GetPendingCommand retrieves a specific pending manual step.
	GetPendingCommand(ctx context.Context, execID uuid.UUID, stepExecID uuid.UUID) (command manual.CommandInfo, err error)

	// ContinuePendingCommand resolves a pending manual step with the operator's response.
	// Resumes the paused execution.
	ContinuePendingCommand(ctx context.Context, execID uuid.UUID, stepExecID uuid.UUID, response manual.InteractionResponse) error
}

// ============================================================================
// APPLICATION SERVICES (thin orchestrators)
// ============================================================================

// TriggerService orchestrates playbook execution from HTTP requests.
// Depends on: ExecutionRuntime, PlaybookStore
//
// Responsibility: variable validation, variable merging, decomposer coordination,
// and mapping HTTP trigger requests to ExecutionRuntime.StartExecution.
type TriggerService interface {
	// ExecutePlaybook triggers execution of a stored playbook with variables.
	ExecutePlaybook(ctx context.Context, playbookID string, variables cacao.Variables) (executionID uuid.UUID, err error)

	// ExecuteUploadedPlaybook triggers execution of a newly-uploaded playbook.
	ExecuteUploadedPlaybook(ctx context.Context, playbook *cacao.Playbook, variables cacao.Variables) (executionID uuid.UUID, err error)
}

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
	ListPlaybookMetas(ctx context.Context) (metas []cacao.Playbook, err error)
}

// ReporterService provides execution reporting and status views.
// Depends on: ExecutionRuntime (via Cache/Informer)
//
// Responsibility: querying execution history, formatting execution reports,
// and providing operational visibility into past/current playbook runs.
type ReporterService interface {
	// ListExecutions returns all recorded execution summaries.
	ListExecutions(ctx context.Context) (executions []cache.ExecutionEntry, err error)

	// GetExecutionReport retrieves a detailed report for a single execution.
	GetExecutionReport(ctx context.Context, executionID uuid.UUID) (report cache.ExecutionEntry, err error)
}

// ============================================================================
// DEPENDENCY NOTES
// ============================================================================

/*
Dependency Flow (what depends on what):

  ExecutionRuntime (core kernel)
    └─ owns: PlaybookStore, Decomposer, Cache, Interaction

  FinRegistry (independent)
    └─ owns: FinStore, FinQueue

  FinWorkService (leased work)
    └─ owns: FinQueue, FinStore

  ManualInbox (execution-aware)
    └─ depends on: ExecutionRuntime

  TriggerService (thin orchestrator)
    └─ depends on: ExecutionRuntime, PlaybookStore

  PlaybookService (CRUD adapter)
    └─ depends on: PlaybookStore

  ReporterService (view adapter)
    └─ depends on: ExecutionRuntime (for Cache)

HTTP Handlers (thin transport)
  └─ depend on: these services (FinRegistry, FinWorkService, ManualInbox, TriggerService, etc.)
  └─ do NOT depend on: stores, queues, runtime directly

Transport Layer (server.go)
  └─ constructs all services
  └─ injects them into handlers
  └─ does NOT manage execution, FIN, or manual logic
*/
