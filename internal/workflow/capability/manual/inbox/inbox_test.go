package inbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"soarca/internal/workflow/capability"
	"soarca/pkg/cacao"
	manualModel "soarca/internal/manual/model"
	"soarca/internal/runs/model"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func TestQueue(t *testing.T) {
	interaction := New([]Notifier{})
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer testCancel()

	testCapComms := manualModel.Waiter{
		Channel:        make(chan manualModel.Response),
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
	interaction := New([]Notifier{})

	testCommand := manualModel.CommandInfo{}

	testCapComms := manualModel.Waiter{
		Channel:        make(chan manualModel.Response),
		TimeoutContext: context.WithoutCancel(context.Background()),
	}
	err := interaction.Queue(testCommand, testCapComms)
	assert.Equal(t, err, errors.New("manual command does not have a deadline"))
}

func TestQueueExitOnTimeout(t *testing.T) {
	interaction := New([]Notifier{})
	timeout := 30 * time.Millisecond
	testCtx, testCancel := context.WithTimeout(context.Background(), timeout)
	defer testCancel()

	hook := NewTestLogHook()
	log.Logger.AddHook(hook)

	testCapComms := manualModel.Waiter{
		Channel:        make(chan manualModel.Response),
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

func TestRegisterRetrieveNewPendingCommand(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	err := interaction.registerPendingCommand(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	retrievedCommand, err := interaction.getPendingCommand(testMetadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	//Channel
	assert.Equal(t,
		retrievedCommand.Channel,
		testChan,
	)

	// Run metadata
	assert.Equal(t,
		retrievedCommand.CommandInfo.Metadata,
		testInteractionCommand.Metadata,
	)
	// Contexts
	assert.Equal(t,
		retrievedCommand.CommandInfo.Context,
		testInteractionCommand.Context,
	)
	// OutArgs
	assert.Equal(t,
		retrievedCommand.CommandInfo.OutArgsVariables,
		testInteractionCommand.OutArgsVariables,
	)
}

func TestGetAllPendingCommands(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	localTestInteractionCommand := testInteractionCommand
	err := interaction.registerPendingCommand(localTestInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testNewInteractionCommand := localTestInteractionCommand
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testNewInteractionCommand.Metadata.RunId = uuid.MustParse(newExecId)

	err = interaction.registerPendingCommand(testNewInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	expectedInteractions := []manualModel.CommandInfo{localTestInteractionCommand, testNewInteractionCommand}
	receivedInteractions := interaction.getAllPendingCommandsInfo()

	// Sort both slices by RunId
	sort.Slice(expectedInteractions, func(i, j int) bool {
		return expectedInteractions[i].Metadata.RunId.String() < expectedInteractions[j].Metadata.RunId.String()
	})
	sort.Slice(receivedInteractions, func(i, j int) bool {
		return receivedInteractions[i].Metadata.RunId.String() < receivedInteractions[j].Metadata.RunId.String()
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

func TestRegisterRetrieveSameRunMultiplePendingCommand(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	localTestInteractionCommand := testInteractionCommand

	err := interaction.registerPendingCommand(localTestInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testNewInteractionCommandSecond := localTestInteractionCommand
	newStepId2 := "test_second_step_id"
	testNewInteractionCommandSecond.Metadata.StepId = newStepId2
	testNewInteractionCommandSecond.Metadata.StepRunId = uuid.MustParse("22a6c41e-6efc-4516-a242-dfbc5c89d562")

	testNewInteractionCommandThird := localTestInteractionCommand
	newStepId3 := "test_third_step_id"
	testNewInteractionCommandThird.Metadata.StepId = newStepId3
	testNewInteractionCommandThird.Metadata.StepRunId = uuid.MustParse("33a6c41e-6efc-4516-a242-dfbc5c89d562")

	err = interaction.registerPendingCommand(testNewInteractionCommandSecond, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = interaction.registerPendingCommand(testNewInteractionCommandThird, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

// Two invocations of the *same* StepId (e.g. two overlapping while-loop
// iterations, or - once implemented - two parallel branches converging on
// the same step) must be able to have their own independently pending
// manual command, each tracked and resolved by its own StepRunId.
func TestRegisterRetrieveSameStepIdDifferentStepRunIdPendingCommands(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	firstInvocation := testInteractionCommand
	secondInvocation := testInteractionCommand
	secondInvocation.Metadata.StepRunId = uuid.MustParse("44a6c41e-6efc-4516-a242-dfbc5c89d562")

	// Same RunId and StepId, different StepRunId.
	err := interaction.registerPendingCommand(firstInvocation, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	err = interaction.registerPendingCommand(secondInvocation, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	firstRetrieved, err := interaction.getPendingCommand(firstInvocation.Metadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	secondRetrieved, err := interaction.getPendingCommand(secondInvocation.Metadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	assert.Equal(t, firstRetrieved.CommandInfo.Metadata, firstInvocation.Metadata)
	assert.Equal(t, secondRetrieved.CommandInfo.Metadata, secondInvocation.Metadata)

	// Resolving/removing one must not affect the other.
	err = interaction.removeCommandFromPending(firstInvocation.Metadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	_, err = interaction.getPendingCommand(secondInvocation.Metadata)
	if err != nil {
		t.Log("second invocation should still be pending after removing the first")
		t.Fail()
	}
}

func TestCopyOutArgsToVars(t *testing.T) {

	interaction := New([]Notifier{})
	testCtx, testCancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer testCancel()

	testCapComms := manualModel.Waiter{
		Channel:        make(chan manualModel.Response),
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

func TestValidateMatchingOutArgs(t *testing.T) {

	interaction := New([]Notifier{})

	respOutArg := cacao.Variable{
		Type:        "pears",
		Name:        "__var2__",
		Description: "this description is a description",
		Value:       "now the value is bananas",
		Constant:    true, // changed but won't be ported
		External:    true, // changed but won't be ported
	}

	respOutArgs := cacao.Variables{"__var2__": respOutArg}

	storedOutArg := cacao.Variable{
		Type:        "apples",
		Name:        "__var2__",
		Description: "this description is a description is a description",
		Value:       "now the value is bananas",
		Constant:    false, // different than respOutArg
		External:    false, // different than respOutArg
	}
	storedOutArgs := cacao.Variables{"__var2__": storedOutArg}

	interactionStorageEntry := manualModel.PendingCommand{
		CommandInfo: manualModel.CommandInfo{
			Metadata:         testMetadata,
			Context:          capability.Context{},
			OutArgsVariables: storedOutArgs,
		},
		Channel: make(chan manualModel.Response),
	}

	expectedLogEntry1 := "provided out arg __var2__ has different value for 'Constant' property of intended out arg. This different value is ignored."
	expectedLogEntry2 := "provided out arg __var2__ has different value for 'Description' property of intended out arg. This different value is ignored."
	expectedLogEntry3 := "provided out arg __var2__ has different value for 'External' property of intended out arg. This different value is ignored."
	expectedLogEntry4 := "provided out arg __var2__ has different value for 'Type' property of intended out arg. This different value is ignored."
	expectedLogs := []string{expectedLogEntry1, expectedLogEntry2, expectedLogEntry3, expectedLogEntry4}

	warns, err := interaction.validateMatchingOutArgs(interactionStorageEntry, respOutArgs)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	sort.Slice(warns, func(i, j int) bool {
		return warns[i] < warns[j]
	})
	sort.Slice(expectedLogs, func(i, j int) bool {
		return warns[i] < warns[j]
	})

	assert.Equal(t, expectedLogs, warns)
}

func TestPostContinueFailOnNonexistingVariable(t *testing.T) {
	interaction := New([]Notifier{})
	timeout := 500 * time.Millisecond
	testCtx, testCancel := context.WithTimeout(context.Background(), timeout)

	defer testCancel()

	hook := NewTestLogHook()
	log.Logger.AddHook(hook)

	testCapComms := manualModel.Waiter{
		Channel:        make(chan manualModel.Response),
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

	outArgsUpdate := manualModel.Response{
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

func TestRegisterRetrieveNewRunNewPendingCommand(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	err := interaction.registerPendingCommand(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testNewInteractionCommand := testInteractionCommand
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testNewInteractionCommand.Metadata.RunId = uuid.MustParse(newExecId)

	err = interaction.registerPendingCommand(testNewInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestGetEmptyPendingCommand(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	emptyCommandInfo, err := interaction.GetPendingCommand(testMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for run " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562",
	}

	assert.Equal(t, emptyCommandInfo, manualModel.CommandInfo{})
	assert.Equal(t, err, expectedErr)
}

func TestFailOnRegisterSamePendingCommand(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	err := interaction.registerPendingCommand(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	err = interaction.registerPendingCommand(testInteractionCommand, testChan)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := fmt.Errorf(
		"a manual command is already pending for run " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562, step run " +
			"11a6c41e-6efc-4516-a242-dfbc5c89d562 (step test_step_id)",
	)
	assert.Equal(t, err, expectedErr)
}

func TestFailOnRetrieveUnexistingRunInteraction(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	testDifferentMetadata := testMetadata
	newExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testDifferentMetadata.RunId = uuid.MustParse(newExecId)

	_, err := interaction.getPendingCommand(testDifferentMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for run 50b6d52c-6efc-4516-a242-dfbc5c89d421",
	}
	assert.Equal(t, err, expectedErr)
}

func TestFailOnRetrieveNonExistingCommandInteraction(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	err := interaction.registerPendingCommand(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	testDifferentMetadata := testMetadata
	newStepRunId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testDifferentMetadata.StepRunId = uuid.MustParse(newStepRunId)

	_, err = interaction.getPendingCommand(testDifferentMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending command found for run " +
			"61a6c41e-6efc-4516-a242-dfbc5c89d562 -> " +
			"step run 50b6d52c-6efc-4516-a242-dfbc5c89d421",
	}
	assert.Equal(t, err, expectedErr)
}

func TestRemovePendingInteraciton(t *testing.T) {
	interaction := New([]Notifier{})
	testChan := make(chan manualModel.Response)
	defer close(testChan)

	err := interaction.registerPendingCommand(testInteractionCommand, testChan)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	pendingCommand, err := interaction.getPendingCommand(testMetadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t,
		pendingCommand.CommandInfo.Metadata.RunId.String(),
		testInteractionCommand.Metadata.RunId.String(),
	)
	assert.Equal(t,
		pendingCommand.CommandInfo.Metadata.StepId,
		testInteractionCommand.Metadata.StepId,
	)

	err = interaction.removeCommandFromPending(testMetadata)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	err = interaction.removeCommandFromPending(testMetadata)
	if err == nil {
		t.Log(err)
		t.Fail()
	}

	expectedErr := manualModel.ErrorPendingCommandNotFound{
		Err: "no pending commands found for run " +
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
var testStepRunIdStr string = "11a6c41e-6efc-4516-a242-dfbc5c89d562"
var testMetadata = run.Metadata{
	RunId:      uuid.MustParse(testUUIDStr),
	PlaybookId: "test_playbook_id",
	StepId:     "test_step_id",
	StepRunId:  uuid.MustParse(testStepRunIdStr),
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
		Commands: []cacao.Command{
			{
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
			OutArgs: []string{"var2"},
			Commands: []cacao.Command{
				{
					Type:    "test_type",
					Command: "test_command",
				},
			},
		},
		Targets: []capability.ResolvedTarget{
			{
				Authentication: cacao.AuthenticationInformation{},
				Target: cacao.AgentTarget{
					ID:          "test_id",
					Type:        "test_type",
					Name:        "test_name",
					Description: "test_description",
				},
			},
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
