package fin

// RegisterRequest is the body of POST /fin/register.
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

// PollRequest is the body of POST /fin/poll. The calling Fin is
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

// ResultRequest is the body of PUT /fin/jobs/{job_id}.
type ResultRequest struct {
	JobResult
}

// StatusPingRequest extends an active lease for long-running jobs.
type StatusPingRequest struct {
	Progress string `json:"progress,omitempty"`
}

// StatusPingResponse carries optional instructions (for example, cancel).
type StatusPingResponse struct {
	Action string `json:"action,omitempty"`
}

// ListResponse is returned by GET /fin/.
type ListResponse struct {
	Fins []Record `json:"fins"`
}
