package mock_interaction

import (
	"context"
	"soarca/pkg/models/manual"

	"github.com/stretchr/testify/mock"
)

type MockInteraction struct {
	mock.Mock
}

func (mock *MockInteraction) Queue(command manual.InteractionCommand,
	manualComms manual.ManualCapabilityCommunication) error {
	args := mock.Called(command, manualComms)
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
	return mock.MatchedBy(func(ch chan manual.InteractionResponse) bool {
		return true
	})
}

// Custom matcher for any ManualCapabilityCommunication
func AnyManualCapabilityCommunication() interface{} {
	return mock.MatchedBy(func(comm manual.ManualCapabilityCommunication) bool {
		return true
	})
}
