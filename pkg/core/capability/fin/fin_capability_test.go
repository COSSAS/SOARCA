package fin

import (
	"context"
	"errors"
	"testing"
	"time"

	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/fin"
	"soarca/test/unittest/mocks/mock_guid"
	mock_time "soarca/test/unittest/mocks/mock_utils/time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockQueue struct {
	mock.Mock
}

func (m *mockQueue) Enqueue(ctx context.Context, job fin.Job) (fin.JobResult, error) {
	args := m.Called(ctx, job)
	return args.Get(0).(fin.JobResult), args.Error(1)
}

type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) List() ([]fin.Record, error) {
	args := m.Called()
	return args.Get(0).([]fin.Record), args.Error(1)
}

func newMetadataAndContext() (execution.Metadata, capability.Context) {
	metadata := execution.Metadata{
		ExecutionId:     uuid.New(),
		PlaybookId:      "playbook--test",
		StepId:          "action--test",
		StepExecutionId: uuid.New(),
	}
	commandContext := capability.Context{
		Commands: []cacao.Command{{Type: "http-api", Command: "GET /"}},
		Targets: []capability.ResolvedTarget{{
			Target:         cacao.AgentTarget{Type: "target", Name: "myself"},
			Authentication: cacao.AuthenticationInformation{Type: "user-auth", Username: "operator"},
		}},
		Step:      cacao.Step{Type: cacao.StepTypeAction, Name: "test step", Timeout: 5000},
		Variables: cacao.NewVariables(),
		Agent:     cacao.AgentTarget{Type: "http-executor", Name: "some fin"},
	}
	return metadata, commandContext
}

func TestExecuteEnqueuesJobRoutedByAgentTypeAndReturnsResultVariables(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	jobId := uuid.New()
	guidMock.On("New").Return(jobId)

	metadata, commandContext := newMetadataAndContext()

	expectedVariables := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString, Name: "__out__", Value: "ok"})

	queue.On("Enqueue", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			job := args.Get(1).(fin.Job)
			assert.Equal(t, job.JobId, jobId)
			assert.Equal(t, job.ExecutionId, metadata.ExecutionId)
			assert.Equal(t, job.PlaybookId, metadata.PlaybookId)
			assert.Equal(t, job.StepId, metadata.StepId)
			assert.Equal(t, job.StepExecutionId, metadata.StepExecutionId)
			// The job must be routed by the step's resolved agent.Type, not
			// by any built-in capability name - this is the whole point of
			// FinCapability being a dynamic-type fallback.
			assert.Equal(t, job.CapabilityType, "http-executor")
			assert.Equal(t, job.LeaseExpiresInSeconds, 5)
			assert.Equal(t, len(job.Commands), 1)
			assert.Equal(t, job.Commands[0].Command, "GET /")
			assert.Equal(t, len(job.Targets), 1)
			assert.Equal(t, job.Targets[0].Target.Name, "myself")
			assert.Equal(t, job.Targets[0].Authentication.Username, "operator")
		}).
		Return(fin.JobResult{State: fin.JobStateSuccess, Variables: expectedVariables}, nil)

	// A nil repository disables the liveness check entirely, so Execute
	// always proceeds straight to Enqueue - this is the pre-existing
	// behavior these tests are pinning.
	finCapability := New(queue, guidMock, nil, nil, 0)
	variables, err := finCapability.Execute(metadata, commandContext)

	assert.Equal(t, err, nil)
	assert.Equal(t, variables, expectedVariables)
	queue.AssertExpectations(t)
}

func TestExecuteReturnsErrorWhenJobResultIsFailure(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(uuid.New())

	metadata, commandContext := newMetadataAndContext()

	queue.On("Enqueue", mock.Anything, mock.Anything).
		Return(fin.JobResult{State: fin.JobStateFailure, Error: "command exited 1"}, nil)

	finCapability := New(queue, guidMock, nil, nil, 0)
	_, err := finCapability.Execute(metadata, commandContext)

	assert.NotEqual(t, err, nil)
	assert.Equal(t, err.Error(), "command exited 1")
}

func TestExecuteReturnsErrorWhenQueueFailsOrTimesOut(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(uuid.New())

	metadata, commandContext := newMetadataAndContext()

	queueErr := errors.New("context deadline exceeded")
	queue.On("Enqueue", mock.Anything, mock.Anything).
		Return(fin.JobResult{}, queueErr)

	finCapability := New(queue, guidMock, nil, nil, 0)
	_, err := finCapability.Execute(metadata, commandContext)

	assert.Equal(t, err, queueErr)
}

func TestExecuteFallsBackToDefaultLeaseWhenStepTimeoutIsUnset(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(uuid.New())

	metadata, commandContext := newMetadataAndContext()
	commandContext.Step.Timeout = 0

	queue.On("Enqueue", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			job := args.Get(1).(fin.Job)
			assert.Equal(t, job.LeaseExpiresInSeconds, 60)
		}).
		Return(fin.JobResult{State: fin.JobStateSuccess, Variables: cacao.NewVariables()}, nil)

	finCapability := New(queue, guidMock, nil, nil, 0)
	_, err := finCapability.Execute(metadata, commandContext)
	assert.Equal(t, err, nil)
}

// ############################################################################
// Fail-fast liveness check (checkCapableFin)
// ############################################################################

func TestExecuteFailsFastWhenNoFinIsRegisteredForCapabilityType(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	repository := new(mockRepository)
	clock := new(mock_time.MockTime)
	clock.On("Now").Return(time.Unix(1000, 0))

	metadata, commandContext := newMetadataAndContext()

	repository.On("List").Return([]fin.Record{
		{FinId: "other-fin", LastSeen: time.Unix(1000, 0), Capabilities: []fin.Capability{{Type: "some-other-type"}}},
	}, nil)

	finCapability := New(queue, guidMock, repository, clock, time.Minute)
	_, err := finCapability.Execute(metadata, commandContext)

	var noCapableFin fin.ErrNoCapableFin
	assert.Equal(t, errors.As(err, &noCapableFin), true)
	assert.Equal(t, noCapableFin.CapabilityType, "http-executor")
	// Must never even touch the queue - that is the whole point of failing fast.
	queue.AssertNotCalled(t, "Enqueue", mock.Anything, mock.Anything)
}

func TestExecuteFailsFastWhenEveryCapableFinIsStale(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	repository := new(mockRepository)
	clock := new(mock_time.MockTime)
	now := time.Unix(10000, 0)
	clock.On("Now").Return(now)

	metadata, commandContext := newMetadataAndContext()

	staleAfter := time.Minute
	repository.On("List").Return([]fin.Record{
		{
			FinId:        "stale-fin",
			LastSeen:     now.Add(-2 * staleAfter),
			Capabilities: []fin.Capability{{Type: "http-executor"}},
		},
	}, nil)

	finCapability := New(queue, guidMock, repository, clock, staleAfter)
	_, err := finCapability.Execute(metadata, commandContext)

	var onlyStale fin.ErrOnlyStaleCapableFins
	assert.Equal(t, errors.As(err, &onlyStale), true)
	assert.Equal(t, onlyStale.CapabilityType, "http-executor")
	assert.Equal(t, onlyStale.FinIds, []string{"stale-fin"})
	queue.AssertNotCalled(t, "Enqueue", mock.Anything, mock.Anything)
}

func TestExecuteProceedsWhenAtLeastOneCapableFinIsLive(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(uuid.New())
	repository := new(mockRepository)
	clock := new(mock_time.MockTime)
	now := time.Unix(10000, 0)
	clock.On("Now").Return(now)

	metadata, commandContext := newMetadataAndContext()

	staleAfter := time.Minute
	repository.On("List").Return([]fin.Record{
		{
			FinId:        "stale-fin",
			LastSeen:     now.Add(-2 * staleAfter),
			Capabilities: []fin.Capability{{Type: "http-executor"}},
		},
		{
			FinId:        "live-fin",
			LastSeen:     now.Add(-1 * time.Second),
			Capabilities: []fin.Capability{{Type: "http-executor"}},
		},
	}, nil)

	queue.On("Enqueue", mock.Anything, mock.Anything).
		Return(fin.JobResult{State: fin.JobStateSuccess, Variables: cacao.NewVariables()}, nil)

	finCapability := New(queue, guidMock, repository, clock, staleAfter)
	_, err := finCapability.Execute(metadata, commandContext)

	assert.Equal(t, err, nil)
	queue.AssertExpectations(t)
}

func TestExecuteProceedsWhenRepositoryListFails(t *testing.T) {
	queue := new(mockQueue)
	guidMock := new(mock_guid.Mock_Guid)
	guidMock.On("New").Return(uuid.New())
	repository := new(mockRepository)
	clock := new(mock_time.MockTime)

	metadata, commandContext := newMetadataAndContext()

	repository.On("List").Return([]fin.Record{}, errors.New("database unavailable"))
	queue.On("Enqueue", mock.Anything, mock.Anything).
		Return(fin.JobResult{State: fin.JobStateSuccess, Variables: cacao.NewVariables()}, nil)

	// A repository error must fail open - fall back to the pre-existing
	// enqueue-and-wait behavior rather than blocking a step over an
	// inability to check liveness.
	finCapability := New(queue, guidMock, repository, clock, time.Minute)
	_, err := finCapability.Execute(metadata, commandContext)

	assert.Equal(t, err, nil)
	queue.AssertExpectations(t)
}
