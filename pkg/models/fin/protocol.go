package fin

// RegisterRequest is the body of POST /api/v1/fins/register — a one-time,
// admin-gated call (see RegistrationToken) that creates a new Fin identity
// and returns its credential (RegisterResponse). Capabilities is an array
// from day one: a single Fin registration can declare more than one
// capability type.
type RegisterRequest struct {
	// RegistrationToken must match FIN_REGISTRATION_TOKEN.
	RegistrationToken string       `json:"registration_token" validate:"required"`
	DisplayName       string       `json:"display_name,omitempty"`
	ProtocolVersion   string       `json:"protocol_version,omitempty"`
	Capabilities      []Capability `json:"capabilities" validate:"required"`
}

// RegisterResponse is returned on successful registration.
type RegisterResponse struct {
	FinId    string `json:"fin_id"`
	FinToken string `json:"fin_token"`
	// Server-provided polling and lease defaults.
	PollIntervalSeconds    int `json:"poll_interval_seconds"`
	LongPollTimeoutSeconds int `json:"long_poll_timeout_seconds"`
	JobLeaseSeconds        int `json:"job_lease_seconds"`
}

// PollRequest is the body of POST /api/v1/fins/poll. The calling Fin is
// identified and authenticated by its Authorization: Bearer fin_token
// header — fin_id and capability types are never resent here; SOARCA
// already knows both server-side, keyed off the token.
type PollRequest struct {
	// Optional hint for available worker slots on the Fin side.
	ConcurrencyAvailable int `json:"concurrency_available,omitempty"`
}

// PollResponse is returned when a job is available.
type PollResponse struct {
	Job Job `json:"job"`
}

// ResultRequest is the body of PUT /api/v1/fins/jobs/{job_id}. job_id is a
// path parameter (identifying which job resource this updates), not
// repeated in the body.
type ResultRequest struct {
	JobResult
}

// StatusPingRequest is the body of PATCH /api/v1/fins/jobs/{job_id}/status
// — a Fin-initiated, optional call used only for jobs that run longer than
// a few seconds. It both extends the job's lease (so a legitimately
// long-running job isn't requeued out from under the Fin still working on
// it) and gives SOARCA a place to piggyback a pending instruction (see
// StatusPingResponse) without any inbound-facing channel on the Fin side.
type StatusPingRequest struct {
	Progress string `json:"progress,omitempty"`
}

// StatusPingResponse carries optional instructions (for example, cancel).
type StatusPingResponse struct {
	Action string `json:"action,omitempty"`
}

// ListResponse is returned by GET /api/v1/fins — read-only
// discovery/observability for admins and dashboards, so playbook authors
// and operators can see what capability types are actually live instead of
// needing out-of-band knowledge of a running Fin's id.
type ListResponse struct {
	Fins []Record `json:"fins"`
}
