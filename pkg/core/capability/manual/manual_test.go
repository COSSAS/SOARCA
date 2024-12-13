package manual

import (
	"soarca/pkg/core/capability"
	"soarca/pkg/models/execution"
	manualModel "soarca/pkg/models/manual"
	"soarca/test/unittest/mocks/mock_interaction"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

func TestManualExecution(t *testing.T) {
	interactionMock := mock_interaction.MockInteraction{}
	var capturedChannel chan manualModel.InteractionResponse

	manual := New(&interactionMock)

	meta := execution.Metadata{}
	commandContext := capability.Context{}

	command := manualModel.InteractionCommand{}

	// Capture the channel passed to Queue
	interactionMock.On("Queue", command, mock_interaction.AnyChannel(), mock_interaction.AnyContext()).Return(nil).Run(func(args mock.Arguments) {
		capturedChannel = args.Get(1).(chan manualModel.InteractionResponse)
	})

	// Use a WaitGroup to wait for the Execute method to complete
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		vars, err := manual.Execute(meta, commandContext)
		assert.Equal(t, err, nil)
		assert.NotEqual(t, vars, nil)
	}()

	// Simulate the response after ensuring the channel is captured
	time.Sleep(100 * time.Millisecond)
	capturedChannel <- manualModel.InteractionResponse{
		Payload: manualModel.ManualOutArgUpdatePayload{
			ResponseOutArgs: manualModel.ManualOutArgs{
				"example": {Value: "example_value"},
			},
		},
	}

	// Wait for the Execute method to complete
	wg.Wait()

}

func TestTimetoutCalculationNotSet(t *testing.T) {
	interactionMock := mock_interaction.MockInteraction{}
	manual := New(&interactionMock)
	timeout := manual.getTimeoutValue(0)
	assert.Equal(t, timeout, time.Minute)
}

func TestTimetoutCalculation(t *testing.T) {
	interactionMock := mock_interaction.MockInteraction{}
	manual := New(&interactionMock)
	timeout := manual.getTimeoutValue(1)
	assert.Equal(t, timeout, time.Millisecond*1)
}
