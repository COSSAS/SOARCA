package manual

import (
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	manualModel "soarca/internal/manual/model"
	"soarca/internal/runs/model"
	"soarca/pkg/utils"
	"soarca/test/unittest/mocks/mock_manual_inbox"
	"sync"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
)

func TestManualRun(t *testing.T) {
	interactionMock := mock_manual_inbox.MockInbox{}
	var capturedComm manualModel.Waiter

	manual := New(&interactionMock)

	meta := run.Metadata{}
	commandContext := capability.Context{}

	command := manualModel.CommandInfo{
		Metadata:         run.Metadata{},
		Context:          capability.Context{},
		OutArgsVariables: cacao.NewVariables(),
	}

	// Capture the channel passed to Queue

	interactionMock.On("Queue", command, mock_manual_inbox.AnyWaiter()).Return(nil).Run(func(args mock.Arguments) {
		capturedComm = args.Get(1).(manualModel.Waiter)
	})
	interactionMock.On("Deregister", meta).Return(nil)

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
	capturedComm.Channel <- manualModel.Response{
		OutArgsVariables: cacao.NewVariables(),
	}

	// Wait for the Execute method to complete
	wg.Wait()

}

func TestTimetoutCalculationNotSet(t *testing.T) {
	interactionMock := mock_manual_inbox.MockInbox{}
	manual := New(&interactionMock)
	timeout := manual.getTimeoutValue(0)
	assert.Equal(t, timeout, utils.DefaultStepTimeout())
}

func TestTimetoutCalculation(t *testing.T) {
	interactionMock := mock_manual_inbox.MockInbox{}
	manual := New(&interactionMock)
	timeout := manual.getTimeoutValue(1)
	assert.Equal(t, timeout, time.Millisecond*1)
}
