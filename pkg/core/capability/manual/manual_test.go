package manual

import (
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
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
	var capturedComm manualModel.ManualCapabilityCommunication

	manual := New(&interactionMock)

	meta := execution.Metadata{}
	commandContext := capability.Context{}

	command := manualModel.CommandInfo{
		Metadata:         execution.Metadata{},
		Context:          capability.Context{},
		OutArgsVariables: cacao.NewVariables(),
	}

	// Capture the channel passed to Queue

	interactionMock.On("Queue", command, mock_interaction.AnyManualCapabilityCommunication()).Return(nil).Run(func(args mock.Arguments) {
		capturedComm = args.Get(1).(manualModel.ManualCapabilityCommunication)
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
	capturedComm.Channel <- manualModel.InteractionResponse{
		OutArgsVariables: cacao.NewVariables(),
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
