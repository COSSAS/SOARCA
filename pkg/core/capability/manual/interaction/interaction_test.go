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

func TestQueue(t *testing.T) {
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

func TestRegisterRetrieveNewPendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Fail()
	}
	retrievedCommand, err := interaction.getPendingInteraction(testMetadata)
	if err != nil {
		t.Fail()
	}

	//Channel
	assert.Equal(t,
		retrievedCommand.Channel,
		testChan,
	)

	// Type
	assert.Equal(t,
		retrievedCommand.CommandData.Type,
		testInteractionCommand.Context.Command.Type,
	)
	// ExecutionId
	assert.Equal(t,
		retrievedCommand.CommandData.ExecutionId,
		testInteractionCommand.Metadata.ExecutionId.String(),
	)
	// PlaybookId
	assert.Equal(t,
		retrievedCommand.CommandData.PlaybookId,
		testInteractionCommand.Metadata.PlaybookId,
	)
	// StepId
	assert.Equal(t,
		retrievedCommand.CommandData.StepId,
		testInteractionCommand.Metadata.StepId,
	)
	// Description
	assert.Equal(t,
		retrievedCommand.CommandData.Description,
		testInteractionCommand.Context.Command.Description,
	)
	// Command
	assert.Equal(t,
		retrievedCommand.CommandData.Command,
		testInteractionCommand.Context.Command.Command,
	)
	// CommandB64
	assert.Equal(t,
		retrievedCommand.CommandData.CommandBase64,
		testInteractionCommand.Context.Command.CommandB64,
	)
	// Target
	assert.Equal(t,
		retrievedCommand.CommandData.Target,
		testInteractionCommand.Context.Target,
	)
	// OutArgs
	assert.Equal(t,
		retrievedCommand.CommandData.OutArgs,
		testInteractionCommand.Context.Variables,
	)
}

func TestRegisterRetrieveSameExecutionMultiplePendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Fail()
	}

	testNewInteractionCommandSecond := testInteractionCommand
	newStepId2 := "test_second_step_id"
	testNewInteractionCommandSecond.Metadata.StepId = newStepId2

	testNewInteractionCommandThird := testInteractionCommand
	newStepId3 := "test_third_step_id"
	testNewInteractionCommandThird.Metadata.StepId = newStepId3

	err = interaction.registerPendingInteraction(testNewInteractionCommandSecond, testChan)
	if err != nil {
		t.Fail()
	}
	err = interaction.registerPendingInteraction(testNewInteractionCommandThird, testChan)
	if err != nil {
		t.Fail()
	}
}

func TestRegisterRetrieveExistingExecutionNewPendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Fail()
	}

	testNewInteractionCommand := testInteractionCommand
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testNewInteractionCommand.Metadata.ExecutionId = uuid.MustParse(newExecId)

	err = interaction.registerPendingInteraction(testNewInteractionCommand, testChan)
	if err != nil {
		t.Fail()
	}
}

func TestFailOnRegisterSamePendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Fail()
	}

	err = interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err == nil {
		t.Fail()
	}

	expectedErr := errors.New(
		"a manual step is already pending for execution " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562, step test_step_id. " +
			"There can only be one pending manual command per action step.",
	)
	assert.Equal(t, err, expectedErr)
}

// ############################################################################
// Utils
// ############################################################################

var testUUIDStr string = "61a6c41e-6efc-4516-a242-dfbc5c89d562"
var testMetadata = execution.Metadata{
	ExecutionId: uuid.MustParse(testUUIDStr),
	PlaybookId:  "test_playbook_id",
	StepId:      "test_step_id",
}

var testInteractionCommand = manualModel.InteractionCommand{
	Metadata: testMetadata,
	Context: capability.Context{
		Command: cacao.Command{
			Type:             "test_type",
			Command:          "test_command",
			Description:      "test_description",
			CommandB64:       "test_command_b64",
			Version:          "1.0",
			PlaybookActivity: "test_activity",
			Headers:          cacao.Headers{},
			Content:          "test_content",
			ContentB64:       "test_content_b64",
		},
		Step: cacao.Step{
			Type:        "test_type",
			ID:          "test_id",
			Name:        "test_name",
			Description: "test_description",
			Timeout:     1,
			StepVariables: cacao.Variables{
				"var1": {
					Type:        "string",
					Name:        "var1",
					Description: "test variable",
					Value:       "test_value",
					Constant:    false,
					External:    false,
				},
			},
			Commands: []cacao.Command{
				{
					Type:    "test_type",
					Command: "test_command",
				},
			},
		},
		Authentication: cacao.AuthenticationInformation{},
		Target: cacao.AgentTarget{
			ID:          "test_id",
			Type:        "test_type",
			Name:        "test_name",
			Description: "test_description",
		},
		Variables: cacao.Variables{
			"var1": {
				Type:        "string",
				Name:        "var1",
				Description: "test variable",
				Value:       "test_value",
				Constant:    false,
				External:    false,
			},
		},
	},
}
