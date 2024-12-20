package interaction

import (
	"context"
	"errors"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	manualModel "soarca/pkg/models/manual"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestQueuSimple(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer testCancel()

	testCapComms := manualModel.ManualCapabilityCommunication{
		Channel:        make(chan manualModel.InteractionResponse),
		TimeoutContext: testCtx,
	}

	// Call queue
	err := interaction.Queue(testInteractionCommand, testCapComms)
	if err != nil {
		t.Fail()
	}

	// Fetch pending command
	retrievedCommand, err := interaction.getPendingInteraction(testMetadata)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t,
		retrievedCommand.CommandData.ExecutionId,
		testInteractionCommand.Metadata.ExecutionId.String(),
	)
	assert.Equal(t,
		retrievedCommand.CommandData.PlaybookId,
		testInteractionCommand.Metadata.PlaybookId,
	)
	assert.Equal(t,
		retrievedCommand.CommandData.StepId,
		testInteractionCommand.Metadata.StepId,
	)
	assert.Equal(t,
		retrievedCommand.CommandData.Command,
		testInteractionCommand.Context.Command.Command,
	)

}

func TestQueueFailWithoutTimeout(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})

	testCommand := manualModel.InteractionCommand{}

	testCapComms := manualModel.ManualCapabilityCommunication{
		Channel:        make(chan manualModel.InteractionResponse),
		TimeoutContext: context.WithoutCancel(context.Background()),
	}
	err := interaction.Queue(testCommand, testCapComms)
	assert.Equal(t, err, errors.New("manual command does not have a deadline"))

}

// ############################################################################
// Utils
// ############################################################################

var testUUIDStr string = "61a6c41e-6efc-4516-a242-dfbc5c89d562"
var testMetadata = execution.Metadata{
	ExecutionId: uuid.MustParse(testUUIDStr),
	PlaybookId:  "dummy_playbook_id",
	StepId:      "dummy_step_id",
}

var testInteractionCommand = manualModel.InteractionCommand{
	Metadata: testMetadata,
	Context: capability.Context{
		Command: cacao.Command{
			Type:             "dummy_type",
			Command:          "dummy_command",
			Description:      "dummy_description",
			CommandB64:       "dummy_command_b64",
			Version:          "1.0",
			PlaybookActivity: "dummy_activity",
			Headers:          cacao.Headers{},
			Content:          "dummy_content",
			ContentB64:       "dummy_content_b64",
		},
		Step: cacao.Step{
			Type:        "dummy_type",
			ID:          "dummy_id",
			Name:        "dummy_name",
			Description: "dummy_description",
			Timeout:     1,
			StepVariables: cacao.Variables{
				"var1": {
					Type:        "string",
					Name:        "var1",
					Description: "dummy variable",
					Value:       "dummy_value",
					Constant:    false,
					External:    false,
				},
			},
			Commands: []cacao.Command{
				{
					Type:    "dummy_type",
					Command: "dummy_command",
				},
			},
		},
		Authentication: cacao.AuthenticationInformation{},
		Target: cacao.AgentTarget{
			ID:          "dummy_id",
			Type:        "dummy_type",
			Name:        "dummy_name",
			Description: "dummy_description",
		},
		Variables: cacao.Variables{
			"var1": {
				Type:        "string",
				Name:        "var1",
				Description: "dummy variable",
				Value:       "dummy_value",
				Constant:    false,
				External:    false,
			},
		},
	},
}
