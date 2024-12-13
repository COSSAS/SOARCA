package manual

import (
	"soarca/pkg/core/capability"
	"soarca/pkg/models/execution"
	manualModel "soarca/pkg/models/manual"
	"soarca/test/unittest/mocks/mock_interaction"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func returnQueueCall(channel chan manualModel.InteractionResponse) {

	time.Sleep(time.Millisecond * 10)
	response := manualModel.InteractionResponse{}
	channel <- response
}

func TestManualExecution(t *testing.T) {
	interactionMock := mock_interaction.MockInteraction{}
	channel := make(chan manualModel.InteractionResponse)
	manual := New(&interactionMock, channel)

	meta := execution.Metadata{}
	context := capability.Context{}

	command := manualModel.InteractionCommand{}
	go returnQueueCall(channel)
	interactionMock.On("Queue", command, channel).Return(nil)
	vars, err := manual.Execute(meta, context)
	assert.Equal(t, err, nil)
	assert.NotEqual(t, vars, nil)

}

func TestTimetoutCalculationNotSet(t *testing.T) {
	interactionMock := mock_interaction.MockInteraction{}
	channel := make(chan manualModel.InteractionResponse)
	manual := New(&interactionMock, channel)
	timeout := manual.getTimeoutValue(0)
	assert.Equal(t, timeout, time.Minute)
}

func TestTimetoutCalculation(t *testing.T) {
	interactionMock := mock_interaction.MockInteraction{}
	channel := make(chan manualModel.InteractionResponse)
	manual := New(&interactionMock, channel)
	timeout := manual.getTimeoutValue(1)
	assert.Equal(t, timeout, time.Millisecond*1)
}
