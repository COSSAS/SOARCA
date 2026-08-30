package mock_manual_inbox

import (
	"context"
	"soarca/internal/manual/model"
	"soarca/internal/runs/model"

	"github.com/stretchr/testify/mock"
)

type MockInbox struct {
	mock.Mock
}

func (mock *MockInbox) Queue(command manual.CommandInfo,
	manualComms manual.Waiter) error {
	args := mock.Called(command, manualComms)
	return args.Error(0)
}

func (mock *MockInbox) Deregister(metadata run.Metadata) error {
	args := mock.Called(metadata)
	return args.Error(0)
}

// Custom matcher for context that always returns true
func AnyContext() interface{} {
	return mock.MatchedBy(func(ctx context.Context) bool {
		return true
	})
}

// Custom matcher to capture the channel
func AnyChannel() interface{} {
	return mock.MatchedBy(func(ch chan manual.Response) bool {
		return true
	})
}

// Custom matcher for any Waiter
func AnyWaiter() interface{} {
	return mock.MatchedBy(func(comm manual.Waiter) bool {
		return true
	})
}
