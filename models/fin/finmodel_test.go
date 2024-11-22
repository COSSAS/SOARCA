package fin

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestFinCommandCreation(t *testing.T) {
	command := NewCommand()
	// Check if set
	assert.Equal(t, command.Type, MessageTypeCommand)
	assert.Equal(t, command.CommandSubstructure.Context.Timeout, 1)
	assert.Equal(t, time.Time.IsZero(command.CommandSubstructure.Context.GeneratedOn), true)

	// Check if not set
	assert.Equal(t, command.MessageId, "")
	assert.Equal(t, command.Meta.SenderId, "")
	assert.Equal(t, time.Time.IsZero(command.Meta.Timestamp), true)
	assert.Equal(t, time.Time.IsZero(command.CommandSubstructure.Context.CompletedOn), true)
	assert.Equal(t, command.CommandSubstructure.Context.Delay, 0)
	assert.Equal(t, command.CommandSubstructure.Context.StepId, "")
	assert.Equal(t, command.CommandSubstructure.Context.ExecutionId, "")
	assert.Equal(t, command.CommandSubstructure.Context.PlaybookId, "")
}
