package fin

import (
	"context"
	"errors"
	"testing"

	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/fin"
	"soarca/test/unittest/mocks/mock_guid"

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

	finCapability := New(queue, guidMock)
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

	finCapability := New(queue, guidMock)
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

	finCapability := New(queue, guidMock)
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

	finCapability := New(queue, guidMock)
	_, err := finCapability.Execute(metadata, commandContext)
	assert.Equal(t, err, nil)
}
