package interaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	manualModel "soarca/pkg/models/manual"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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
		t.Log(err)
		t.Fail()
	}
}

func TestQueueFailWithoutTimeout(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})

	testCommand := manualModel.CommandInfo{}

	testCapComms := manualModel.ManualCapabilityCommunication{
		Channel:        make(chan manualModel.InteractionResponse),
		TimeoutContext: context.WithoutCancel(context.Background()),
	}
	err := interaction.Queue(testCommand, testCapComms)
	assert.Equal(t, err, errors.New("manual command does not have a deadline"))
}

func TestQueueExitOnTimeout(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	timeout := 30 * time.Millisecond
	testCtx, testCancel := context.WithTimeout(context.Background(), timeout)
	defer testCancel()

	hook := NewTestLogHook()
	log.Logger.AddHook(hook)

	testCapComms := manualModel.ManualCapabilityCommunication{
		Channel:        make(chan manualModel.InteractionResponse),
		TimeoutContext: testCtx,
	}

	err := interaction.Queue(testInteractionCommand, testCapComms)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	time.Sleep(50 * time.Millisecond)

	expectedLogEntry := "manual command timed out. deregistering associated pending command"
	assert.NotEqual(t, len(hook.Entries), 0)
	contains := false
	for _, entry := range hook.Entries {
		if strings.Contains(entry.Message, expectedLogEntry) {
			contains = true
			break
		}
	}
	assert.Equal(t, contains, true)

}

func TestRegisterRetrieveNewPendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	retrievedCommand, err := interaction.getPendingInteraction(testMetadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	//Channel
	assert.Equal(t,
		retrievedCommand.Channel,
		testChan,
	)

	// ExecutionId
	assert.Equal(t,
		retrievedCommand.CommandInfo.Metadata.ExecutionId.String(),
		testInteractionCommand.Metadata.ExecutionId.String(),
	)
	// PlaybookId
	assert.Equal(t,
		retrievedCommand.CommandInfo.Metadata.PlaybookId,
		testInteractionCommand.Metadata.PlaybookId,
	)
	// StepId
	assert.Equal(t,
		retrievedCommand.CommandInfo.Metadata.StepId,
		testInteractionCommand.Metadata.StepId,
	)
	// Description
	assert.Equal(t,
		retrievedCommand.CommandInfo.Context.Command.Description,
		testInteractionCommand.Context.Command.Description,
	)
	// Command
	assert.Equal(t,
		retrievedCommand.CommandInfo.Context.Command.Command,
		testInteractionCommand.Context.Command.Command,
	)
	// CommandB64
	assert.Equal(t,
		retrievedCommand.CommandInfo.Context.Command.CommandB64,
		testInteractionCommand.Context.Command.CommandB64,
	)
	// Target
	assert.Equal(t,
		retrievedCommand.CommandInfo.Context.Target,
		testInteractionCommand.Context.Target,
	)
	// OutArgs
	assert.Equal(t,
		retrievedCommand.CommandInfo.OutArgsVariables,
		testInteractionCommand.OutArgsVariables,
	)
}

func TestGetAllPendingInteractions(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	localTestInteractionCommand := testInteractionCommand
	err := interaction.registerPendingInteraction(localTestInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testNewInteractionCommand := localTestInteractionCommand
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testNewInteractionCommand.Metadata.ExecutionId = uuid.MustParse(newExecId)

	err = interaction.registerPendingInteraction(testNewInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedInteractions := []manualModel.CommandInfo{localTestInteractionCommand, testNewInteractionCommand}
	receivedInteractions := interaction.getAllPendingCommandsInfo()

	// Sort both slices by ExecutionId
	sort.Slice(expectedInteractions, func(i, j int) bool {
		return expectedInteractions[i].Metadata.ExecutionId.String() < expectedInteractions[j].Metadata.ExecutionId.String()
	})
	sort.Slice(receivedInteractions, func(i, j int) bool {
		return receivedInteractions[i].Metadata.ExecutionId.String() < receivedInteractions[j].Metadata.ExecutionId.String()
	})

	receivedInteractionsJson, err := json.MarshalIndent(receivedInteractions, "", "  ")
	if err != nil {
		t.Log("failed to marshal received interactions")
		t.Log(err)
		t.Fail()
	}
	t.Log("received interactions")
	t.Log(string(receivedInteractionsJson))

	for i, receivedInteraction := range receivedInteractions {
		if expectedInteractions[i].Metadata != receivedInteraction.Metadata {
			err = fmt.Errorf("expected %v, but got %v", expectedInteractions, receivedInteractions)
			t.Log(err)
			t.Fail()
		}
		if !reflect.DeepEqual(expectedInteractions[i].OutArgsVariables, receivedInteraction.OutArgsVariables) {
			err = fmt.Errorf("expected %v, but got %v", expectedInteractions, receivedInteractions)
			t.Log(err)
			t.Fail()
		}
		if !reflect.DeepEqual(expectedInteractions[i].Context, receivedInteraction.Context) {
			err = fmt.Errorf("expected %v, but got %v", expectedInteractions[i].Context, receivedInteraction.Context)
			t.Log(err)
			t.Fail()
		}
	}
}

func TestRegisterRetrieveSameExecutionMultiplePendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	localTestInteractionCommand := testInteractionCommand

	err := interaction.registerPendingInteraction(localTestInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testNewInteractionCommandSecond := localTestInteractionCommand
	newStepId2 := "test_second_step_id"
	testNewInteractionCommandSecond.Metadata.StepId = newStepId2

	testNewInteractionCommandThird := localTestInteractionCommand
	newStepId3 := "test_third_step_id"
	testNewInteractionCommandThird.Metadata.StepId = newStepId3

	err = interaction.registerPendingInteraction(testNewInteractionCommandSecond, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = interaction.registerPendingInteraction(testNewInteractionCommandThird, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestCopyOutArgsToVars(t *testing.T) {

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
		t.Log(err)
		t.Fail()
	}

	outArg := cacao.Variable{
		Type:        "string",
		Name:        "var2",
		Description: "this description will not make it to the returned var",
		Value:       "now the value is bananas",
		Constant:    true, // changed but won't be ported
		External:    true, // changed but won't be ported
	}

	expectedVariable := cacao.Variable{
		Type:  "string",
		Name:  "var2",
		Value: "now the value is bananas",
	}

	responseOutArgs := cacao.Variables{"var2": outArg}

	vars := responseOutArgs
	assert.Equal(t, expectedVariable.Type, vars["var2"].Type)
	assert.Equal(t, expectedVariable.Name, vars["var2"].Name)
	assert.Equal(t, expectedVariable.Value, vars["var2"].Value)
}

func TestPostContinueWarningsRaised(t *testing.T) {

	interaction := New([]IInteractionIntegrationNotifier{})
	timeout := 5 * time.Second
	testCtx, testCancel := context.WithTimeout(context.Background(), timeout)

	hook := NewTestLogHook()
	log.Logger.AddHook(hook)

	testCapComms := manualModel.ManualCapabilityCommunication{
		Channel:        make(chan manualModel.InteractionResponse),
		TimeoutContext: testCtx,
	}

	err := interaction.Queue(testInteractionCommand, testCapComms)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	outArg := cacao.Variable{
		Type:        "banana",
		Name:        "var2",
		Description: "this description will not make it to the returned var",
		Value:       "now the value is bananas",
		Constant:    true, // changed but won't be ported
		External:    true, // changed but won't be ported
	}

	outArgsUpdate := manualModel.InteractionResponse{
		Metadata:         testMetadata,
		ResponseStatus:   "success",
		ResponseError:    nil,
		OutArgsVariables: cacao.Variables{"var2": outArg},
	}

	// Start a goroutine to read from the channel to avoid blocking
	go func() {
		for response := range testCapComms.Channel {
			log.Trace("Received response:", response)
		}
	}()

	err = interaction.PostContinue(outArgsUpdate)
	var expectedErr error = nil
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, err, expectedErr)

	// Simulating Manual Capability closing the channel and the context
	close(testCapComms.Channel)
	testCancel()
	time.Sleep(300 * time.Millisecond)

	expectedLogEntry1 := "provided out arg var2 has different value for 'Constant' property of intended out arg. This different value is ignored."
	expectedLogEntry2 := "provided out arg var2 has different value for 'Description' property of intended out arg. This different value is ignored."
	expectedLogEntry3 := "provided out arg var2 has different value for 'External' property of intended out arg. This different value is ignored."
	expectedLogEntry4 := "provided out arg var2 has different value for 'Type' property of intended out arg. This different value is ignored."
	expectedLogs := []string{expectedLogEntry1, expectedLogEntry2, expectedLogEntry3, expectedLogEntry4}

	all := true
	for _, expectedMessage := range expectedLogs {
		containsAll := true
		for _, entry := range hook.Entries {
			if strings.Contains(entry.Message, expectedMessage) {
				containsAll = true
				break
			}
			if !strings.Contains(entry.Message, expectedMessage) {
				containsAll = false
			}
		}
		if !containsAll {
			t.Logf("log message: '%s' not found in logged messages", expectedMessage)
			all = false
			break
		}
	}

	assert.NotEqual(t, len(hook.Entries), 0)
	assert.Equal(t, all, true)

}

func TestPostContinueFailOnNonexistingVariable(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	timeout := 500 * time.Millisecond
	testCtx, testCancel := context.WithTimeout(context.Background(), timeout)

	defer testCancel()

	hook := NewTestLogHook()
	log.Logger.AddHook(hook)

	testCapComms := manualModel.ManualCapabilityCommunication{
		Channel:        make(chan manualModel.InteractionResponse),
		TimeoutContext: testCtx,
	}
	defer close(testCapComms.Channel)

	err := interaction.Queue(testInteractionCommand, testCapComms)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	outArg := cacao.Variable{
		Type:  "string",
		Name:  "testNotExisting",
		Value: "now the value is bananas",
	}

	outArgsUpdate := manualModel.InteractionResponse{
		Metadata:         testMetadata,
		ResponseStatus:   "success",
		ResponseError:    nil,
		OutArgsVariables: cacao.Variables{"testNotExisting": outArg},
	}

	err = interaction.PostContinue(outArgsUpdate)

	expectedErr := manualModel.ErrorNonMatchingOutArgs{
		Err: fmt.Sprintf("provided out arg %s does not match any intended out arg", outArg.Name)}

	assert.Equal(t, err, expectedErr)
}

func TestRegisterRetrieveNewExecutionNewPendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testNewInteractionCommand := testInteractionCommand
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testNewInteractionCommand.Metadata.ExecutionId = uuid.MustParse(newExecId)

	err = interaction.registerPendingInteraction(testNewInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetEmptyPendingCommand(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	emptyCommandInfo, err := interaction.GetPendingCommand(testMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for execution " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562",
	}

	assert.Equal(t, emptyCommandInfo, manualModel.CommandInfo{})
	assert.Equal(t, err, expectedErr)
}

func TestFailOnRegisterSamePendingInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	err = interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := errors.New(
		"a manual step is already pending for execution " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562, step test_step_id. " +
			"There can only be one pending manual command per action step.",
	)
	assert.Equal(t, err, expectedErr)
}

func TestFailOnRetrieveUnexistingExecutionInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	testDifferentMetadata := testMetadata
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testDifferentMetadata.ExecutionId = uuid.MustParse(newExecId)

	_, err := interaction.getPendingInteraction(testDifferentMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for execution 50b6d52c-6efc-4516-a242-dfbc5c89d421",
	}
	assert.Equal(t, err, expectedErr)
}

func TestFailOnRetrieveNonExistingCommandInteraction(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testDifferentMetadata := testMetadata
	newStepId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testDifferentMetadata.StepId = newStepId

	_, err = interaction.getPendingInteraction(testDifferentMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for execution " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562 -> " +
			"step 50b6d52c-6efc-4516-a242-dfbc5c89d421",
	}
	assert.Equal(t, err, expectedErr)
}

func TestRemovePendingInteraciton(t *testing.T) {
	interaction := New([]IInteractionIntegrationNotifier{})
	testChan := make(chan manualModel.InteractionResponse)
	defer close(testChan)

	err := interaction.registerPendingInteraction(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	pendingCommand, err := interaction.getPendingInteraction(testMetadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t,
		pendingCommand.CommandInfo.Metadata.ExecutionId.String(),
		testInteractionCommand.Metadata.ExecutionId.String(),
	)
	assert.Equal(t,
		pendingCommand.CommandInfo.Metadata.StepId,
		testInteractionCommand.Metadata.StepId,
	)

	err = interaction.removeInteractionFromPending(testMetadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	err = interaction.removeInteractionFromPending(testMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for execution " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562",
	}
	assert.Equal(t, err, expectedErr)
}

// ############################################################################
// Utils
// ############################################################################

type TestHook struct {
	Entries []*logrus.Entry
}

func (hook *TestHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (hook *TestHook) Fire(entry *logrus.Entry) error {
	hook.Entries = append(hook.Entries, entry)
	return nil
}

func NewTestLogHook() *TestHook {
	return &TestHook{}
}

var testUUIDStr string = "61a6c41e-6efc-4516-a242-dfbc5c89d562"
var testMetadata = execution.Metadata{
	ExecutionId: uuid.MustParse(testUUIDStr),
	PlaybookId:  "test_playbook_id",
	StepId:      "test_step_id",
}

var testInteractionCommand = manualModel.CommandInfo{
	Metadata: testMetadata,
	OutArgsVariables: cacao.Variables{
		"var2": {
			Type:        "string",
			Name:        "var2",
			Description: "test variable",
			Value:       "",
			Constant:    false,
			External:    false,
		},
	},
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
					Value:       "test_value_1",
					Constant:    false,
					External:    false,
				},
			},
			OutArgs: cacao.OutArgs{"var2"},
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
			"var2": {
				Type:        "string",
				Name:        "var2",
				Description: "test variable",
				Value:       "test_value_2",
				Constant:    false,
				External:    false,
			},
		},
	},
}
