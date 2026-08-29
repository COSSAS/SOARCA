// Package fin defines data models for the Fin HTTP/JSON protocol.
package fin

import (
	"fmt"
	"strings"
	"time"

	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"

	"github.com/google/uuid"
)

// Capability describes one capability a Fin can execute.
type Capability struct {
	// Type is the routing key used to match jobs to Fins.
	Type string `bson:"type" json:"type" validate:"required"`
	// Description is free text for humans (dashboards, logs) — it plays no
	// role in routing.
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	// Version is the capability implementation's own version string, for
	// operator/debugging visibility only.
	Version string `bson:"version,omitempty" json:"version,omitempty"`
	// StepExamples are optional, illustrative CACAO action steps showing
	// how a playbook author would invoke this capability (agent, commands,
	// targets, ...), surfaced to help playbook authors — they are never
	// interpreted or validated by SOARCA. A capability may offer more than
	// one example (e.g. to illustrate different commands or target shapes).
	StepExamples []cacao.Step `bson:"step_examples,omitempty" json:"step_examples,omitempty"`
}

// Record is a persisted Fin registration with liveness metadata.
type Record struct {
	// FinId is server-assigned at registration (never client-chosen),
	// avoiding id collisions and any implicit trust that a Fin picks a
	// unique id for itself.
	FinId string `bson:"_id" json:"fin_id"`
	// FinTokenHash is a one-way hash of the per-Fin credential (fin_token)
	// returned once at registration. The plaintext token is never
	// persisted: SOARCA hashes an incoming Authorization: Bearer token the
	// same way and compares hashes, so a database read/leak alone cannot
	// recover a usable credential.
	FinTokenHash string `bson:"fin_token_hash" json:"-"`
	// DisplayName is free text for humans/logs only; it plays no role in
	// routing or identity.
	DisplayName     string       `bson:"display_name,omitempty" json:"display_name,omitempty"`
	ProtocolVersion string       `bson:"protocol_version,omitempty" json:"protocol_version,omitempty"`
	Capabilities    []Capability `bson:"capabilities" json:"capabilities"`
	RegisteredAt    time.Time    `bson:"registered_at" json:"registered_at"`
	// LastSeen is updated by poll, result submission, and status ping.
	LastSeen time.Time `bson:"last_seen" json:"last_seen"`
	// Stale is computed at read time and is not persisted.
	Stale bool `bson:"-" json:"stale"`
}

// JobState is the aggregated outcome of a job.
type JobState string

const (
	JobStateSuccess JobState = "success"
	JobStateFailure JobState = "failure"
)

// Job is one pollable unit of work for a Fin.
type Job struct {
	// JobId identifies this specific poll-able unit of work (== the lease
	// handle used for result submission and status pings).
	JobId uuid.UUID `bson:"_id" json:"job_id"`
	// StepExecutionId disambiguates repeated invocations of the same StepId.
	ExecutionId     uuid.UUID `bson:"execution_id" json:"execution_id"`
	PlaybookId      string    `bson:"playbook_id" json:"playbook_id"`
	StepId          string    `bson:"step_id" json:"step_id"`
	StepExecutionId uuid.UUID `bson:"step_execution_id" json:"step_execution_id"`
	// CapabilityType is the routing key this job was queued under — it
	// matches Capability.Type of whichever Fin ultimately claims it.
	CapabilityType string `bson:"capability_type" json:"capability_type"`
	// Lease timeout; jobs may be requeued after expiry.
	LeaseExpiresInSeconds int                         `bson:"lease_expires_in_seconds" json:"lease_expires_in_seconds"`
	Step                  StepInfo                    `bson:"step" json:"step"`
	Commands              []Command                   `bson:"commands" json:"commands"`
	Targets               []capability.ResolvedTarget `bson:"targets" json:"targets"`
	Variables             cacao.Variables             `bson:"variables" json:"variables"`
}

// StepInfo carries the metadata a Fin needs.
type StepInfo struct {
	Name        string `bson:"name,omitempty" json:"name,omitempty"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`
	Timeout     int    `bson:"timeout,omitempty" json:"timeout,omitempty"`
	Delay       int    `bson:"delay,omitempty" json:"delay,omitempty"`
}

// Command is one command in a job.
type Command struct {
	Type       string              `bson:"type" json:"type"`
	Command    string              `bson:"command,omitempty" json:"command,omitempty"`
	CommandB64 string              `bson:"command_b64,omitempty" json:"command_b64,omitempty"`
	Content    string              `bson:"content,omitempty" json:"content,omitempty"`
	ContentB64 string              `bson:"content_b64,omitempty" json:"content_b64,omitempty"`
	Headers    map[string][]string `bson:"headers,omitempty" json:"headers,omitempty"`
}

// TargetResult is optional, additive per-target diagnostic detail on a job
// result — useful for the reporter/GUI/audit log, but never consulted by
// playbook control flow, which only ever sees the single aggregated
// (State, Variables) pair on JobResult.
type TargetResult struct {
	// TargetIndex identifies which entry in the job's Targets this result
	// is for. Nil when the job had no targets at all.
	TargetIndex *int     `bson:"target_index,omitempty" json:"target_index,omitempty"`
	State       JobState `bson:"state" json:"state"`
	// FailedCommandIndex identifies which command in Commands aborted this
	// target's sequence, if any.
	FailedCommandIndex *int            `bson:"failed_command_index,omitempty" json:"failed_command_index,omitempty"`
	Variables          cacao.Variables `bson:"variables,omitempty" json:"variables,omitempty"`
	Error              string          `bson:"error,omitempty" json:"error,omitempty"`
}

// JobResult is submitted by a Fin after job completion.
type JobResult struct {
	State         JobState        `bson:"state" json:"state" validate:"required"`
	Variables     cacao.Variables `bson:"variables,omitempty" json:"variables,omitempty"`
	Error         string          `bson:"error,omitempty" json:"error,omitempty"`
	TargetResults []TargetResult  `bson:"target_results,omitempty" json:"target_results,omitempty"`
}

// Errors ######################################################################

// ErrRegistrationTokenInvalid is returned for invalid registration tokens.
type ErrRegistrationTokenInvalid struct{}

func (e ErrRegistrationTokenInvalid) Error() string {
	return "invalid or missing registration token"
}

// ErrFinTokenInvalid is returned for unknown Fin tokens.
type ErrFinTokenInvalid struct{}

func (e ErrFinTokenInvalid) Error() string {
	return "invalid or unknown fin token"
}

// ErrFinNotFound indicates no registration exists for the FinId.
type ErrFinNotFound struct {
	FinId string
}

func (e ErrFinNotFound) Error() string {
	return "no fin registered with id " + e.FinId
}

// ErrAlreadyRegistered indicates a duplicate Fin registration ID.
type ErrAlreadyRegistered struct {
	FinId string
}

func (e ErrAlreadyRegistered) Error() string {
	return "a fin is already registered with id " + e.FinId
}

// ErrJobNotFound indicates the job does not exist.
type ErrJobNotFound struct {
	JobId string
}

func (e ErrJobNotFound) Error() string {
	return "no job found with id " + e.JobId
}

// ErrJobNotLeasedToFin indicates a lease mismatch for the acting Fin.
type ErrJobNotLeasedToFin struct {
	JobId string
	FinId string
}

func (e ErrJobNotLeasedToFin) Error() string {
	return "job " + e.JobId + " is not leased to fin " + e.FinId
}

// ErrNoCapableFin indicates no Fin is registered for the capability type.
type ErrNoCapableFin struct {
	CapabilityType string
}

func (e ErrNoCapableFin) Error() string {
	return "no fin is registered for capability type " + e.CapabilityType
}

// ErrOnlyStaleCapableFins indicates all capable Fins are stale.
type ErrOnlyStaleCapableFins struct {
	CapabilityType string
	FinIds         []string
	StaleAfter     time.Duration
}

func (e ErrOnlyStaleCapableFins) Error() string {
	return fmt.Sprintf(
		"every fin registered for capability type %s has not been seen in over %s (fin ids: %s); assuming none are still running",
		e.CapabilityType, e.StaleAfter, strings.Join(e.FinIds, ", "),
	)
}
