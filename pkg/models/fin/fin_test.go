package fin

import (
	"encoding/json"
	"testing"

	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

// These tests pin the Fin protocol JSON wire shape.

func TestRegisterRequestJSONShape(t *testing.T) {
	request := RegisterRequest{
		RegistrationToken: "shared secret, see below",
		DisplayName:       "Example Pong Fin",
		ProtocolVersion:   "1.0.0",
		Capabilities: []Capability{
			{
				Type:        "pong",
				Description: "Ping/Pong capability",
				Version:     "0.1.0",
				StepExamples: []cacao.Step{
					{
						Type: "action",
						Name: "pong",
						Commands: []cacao.Command{
							{Type: "manual", Command: "pong"},
						},
					},
				},
			},
			{Type: "ping", Description: "Ping capability (same Fin process, second capability)", Version: "0.1.0"},
		},
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(bytes, &roundTripped); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, roundTripped["registration_token"], "shared secret, see below")
	assert.Equal(t, roundTripped["display_name"], "Example Pong Fin")
	assert.Equal(t, roundTripped["protocol_version"], "1.0.0")
	capabilities, ok := roundTripped["capabilities"].([]any)
	if !ok {
		t.Fatal("expected capabilities to be a JSON array")
	}
	assert.Equal(t, len(capabilities), 2)
	firstCapability, ok := capabilities[0].(map[string]any)
	if !ok {
		t.Fatal("expected first capability to be a JSON object")
	}
	assert.Equal(t, firstCapability["type"], "pong")

	// Round-trip back into the Go type and confirm equality.
	var decoded RegisterRequest
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, decoded, request)
}

func TestRegisterResponseJSONShape(t *testing.T) {
	response := RegisterResponse{
		FinId:                  "fin-9f2c1e3a",
		FinToken:               "opaque-per-fin-secret",
		PollIntervalSeconds:    5,
		LongPollTimeoutSeconds: 25,
		JobLeaseSeconds:        60,
	}

	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(bytes, &roundTripped); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, roundTripped["fin_id"], "fin-9f2c1e3a")
	assert.Equal(t, roundTripped["fin_token"], "opaque-per-fin-secret")
	assert.Equal(t, roundTripped["poll_interval_seconds"], float64(5))
	assert.Equal(t, roundTripped["long_poll_timeout_seconds"], float64(25))
	assert.Equal(t, roundTripped["job_lease_seconds"], float64(60))
}

func TestPollResponseJobJSONShape(t *testing.T) {
	jobId := uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	executionId := uuid.MustParse("d09351a2-a075-40c8-8054-0b7c423db83f")
	stepExecutionId := uuid.MustParse("81eff59f-d084-4324-9e0a-59e353dbd28f")

	job := Job{
		JobId:                 jobId,
		RunId:                 executionId,
		PlaybookId:            "playbook--uuid",
		StepId:                "action--uuid",
		StepRunId:             stepExecutionId,
		CapabilityType:        "ssh-executor",
		LeaseExpiresInSeconds: 60,
		Step: StepInfo{
			Name:        "string",
			Description: "string",
			Timeout:     30,
			Delay:       0,
		},
		Commands: []Command{
			{Type: "bash", Command: "string"},
		},
		Targets: []capability.ResolvedTarget{
			{
				Target: cacao.AgentTarget{
					Type:    "net-address",
					Name:    "target",
					Address: cacao.Addresses{"ipv4": {"10.0.0.1"}},
					Port:    "22",
				},
				Authentication: cacao.AuthenticationInformation{
					Type:     "user-auth",
					Username: "operator",
				},
			},
		},
		Variables: cacao.NewVariables(),
	}

	response := PollResponse{Job: job}

	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(bytes, &roundTripped); err != nil {
		t.Fatal(err)
	}

	jobField, ok := roundTripped["job"].(map[string]any)
	if !ok {
		t.Fatal("expected top-level job field")
	}
	assert.Equal(t, jobField["job_id"], jobId.String())
	assert.Equal(t, jobField["capability_type"], "ssh-executor")
	assert.Equal(t, jobField["lease_expires_in_seconds"], float64(60))

	var decoded PollResponse
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, decoded.Job.JobId, jobId)
	assert.Equal(t, decoded.Job.CapabilityType, "ssh-executor")
	assert.Equal(t, len(decoded.Job.Commands), 1)
	assert.Equal(t, len(decoded.Job.Targets), 1)
}

func TestResultRequestJSONShape(t *testing.T) {
	targetIndex := 0
	failedCommandIndex := 1

	request := ResultRequest{
		JobResult: JobResult{
			State: JobStateFailure,
			Error: "something failed",
			TargetResults: []TargetResult{
				{
					TargetIndex:        &targetIndex,
					State:              JobStateFailure,
					FailedCommandIndex: &failedCommandIndex,
					Error:              "command 1 failed",
				},
			},
		},
	}

	bytes, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	var decoded ResultRequest
	if err := json.Unmarshal(bytes, &decoded); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, decoded.State, JobStateFailure)
	assert.Equal(t, decoded.Error, "something failed")
	assert.Equal(t, len(decoded.TargetResults), 1)
	assert.Equal(t, *decoded.TargetResults[0].TargetIndex, 0)
	assert.Equal(t, *decoded.TargetResults[0].FailedCommandIndex, 1)
}

func TestStatusPingResponseJSONShape(t *testing.T) {
	response := StatusPingResponse{Action: "cancel"}

	bytes, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	var roundTripped map[string]any
	if err := json.Unmarshal(bytes, &roundTripped); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, roundTripped["action"], "cancel")
}
