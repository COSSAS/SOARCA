package manual

import (
	"errors"
	"reflect"
	"soarca/internal/workflow/capability"
	"soarca/internal/transport/http/schema"
	"soarca/pkg/cacao"
	"soarca/internal/manual/model"
	"soarca/internal/runs/model"
	"soarca/test/unittest/mocks/mock_manual_inbox_storage"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestParseManualOutArgsUpdate(t *testing.T) {
	manualHandler := NewManualHandler(&mock_manual_inbox_storage.MockInboxStorage{})

	jsonPayload := `{"type":"out-args-update","response_status":"success","response_out_args":{"__test__":{"type":"string","name":"__test__","value":"updated!"}}}`
	bytesPayload := []byte(jsonPayload)

	outVariable := cacao.Variable{Type: "string", Name: "__test__", Value: "updated!"}
	outVariables := map[string]cacao.Variable{"__test__": outVariable}

	expectedPayload := api.ManualOutArgsUpdatePayload{
		Type:            "out-args-update",
		ResponseStatus:  manual.ManualResponseSuccessStatus,
		ResponseOutArgs: outVariables,
	}

	receivedPayload, err := manualHandler.parseManualOutArgsUpdate(bytesPayload)
	if err != nil {
		t.Fatalf("failed to parse manual out args update: %v", err)
	}
	assert.Equal(t, receivedPayload, expectedPayload)
}

func TestParseManualOutArgsUpdateFailOnVariablesNames(t *testing.T) {
	manualHandler := NewManualHandler(&mock_manual_inbox_storage.MockInboxStorage{})

	jsonPayload := `{"type":"out-args-update","response_status":"success","response_out_args":{"__test__":{"type":"string","name":"__wrong_name__","value":"updated!"}}}`
	bytesPayload := []byte(jsonPayload)

	expecedErr := errors.New("variable name mismatch for variable __test__: has different name property: __wrong_name__")
	_, err := manualHandler.parseManualOutArgsUpdate(bytesPayload)
	if err == nil {
		t.Log("an error for non-matching variables names should have been raised")
		t.Fail()
	}

	assert.Equal(t, err, expecedErr)
}

func TestParseManualOutArgsUpdateFailOnInvalidModel(t *testing.T) {
	manualHandler := NewManualHandler(&mock_manual_inbox_storage.MockInboxStorage{})

	jsonPayload := `{"invalidProperty":"out-args-update","response_status":"success","response_out_args":{"__test__":{"type":"string","name":"__wrong_name__","value":"updated!"}}}`
	bytesPayload := []byte(jsonPayload)

	expectedErr := "failed to unmarshal JSON: json: unknown field \"invalidProperty\""
	_, err := manualHandler.parseManualOutArgsUpdate(bytesPayload)
	if err == nil {
		t.Log("an error for non-matching variables names should have been raised")
		t.Fail()
	}

	assert.Equal(t, err.Error(), expectedErr)
}

func TestParseCommandInfoToResponseIncludesAllCommandsAndTargets(t *testing.T) {

	manualHandler := NewManualHandler(&mock_manual_inbox_storage.MockInboxStorage{})

	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"
	testStepExecId := "71a4d52c-6efc-4516-a242-dfbc5c89d999"

	commandOne := cacao.Command{Type: "manual", Command: "please do a test thanks", Description: "testing!"}
	commandTwo := cacao.Command{Type: "manual", CommandB64: "cGxlYXNlIGRvIGFub3RoZXIgdGVzdA==", Description: "testing again!"}
	targetOne := cacao.AgentTarget{Type: "target", Name: "myself"}
	targetTwo := cacao.AgentTarget{Type: "target", Name: "someoneelse"}
	authOne := cacao.AuthenticationInformation{Type: "user-auth", Username: "operator", Password: "hunter2"}
	variable2 := cacao.Variable{Type: "string", Name: "__test__", Value: "some value"}
	inputVariable := map[string]cacao.Variable{"__test__": variable2}

	context := capability.Context{
		Commands: []cacao.Command{commandOne, commandTwo},
		Targets: []capability.ResolvedTarget{
			{Target: targetOne, Authentication: authOne},
			{Target: targetTwo},
		},
		Variables: inputVariable,
	}

	testVariables := cacao.NewVariables(cacao.Variable{Type: "string", Name: "__test__", Value: "test!"})

	commandInfo := manual.CommandInfo{
		Metadata: run.Metadata{
			PlaybookId: testPlaybookId,
			RunId:      uuid.MustParse(testExecId),
			StepId:     testStepId,
			StepRunId:  uuid.MustParse(testStepExecId)},
		Context:          context,
		OutArgsVariables: testVariables,
	}

	expectedInteractionCommand := api.PendingCommandData{
		Type:       "manual-command-info",
		RunId:      testExecId,
		PlaybookId: testPlaybookId,
		StepId:     testStepId,
		StepRunId:  testStepExecId,
		Commands: []api.ManualCommand{
			{Description: "testing!", Command: "please do a test thanks", CommandIsBase64: false},
			{Description: "testing again!", Command: "cGxlYXNlIGRvIGFub3RoZXIgdGVzdA==", CommandIsBase64: true},
		},
		Targets: []capability.ResolvedTarget{
			{Target: targetOne, Authentication: authOne},
			{Target: targetTwo},
		},
		OutVariables: testVariables,
	}

	returnPendingCommandData := manualHandler.parseCommandInfoToResponse(commandInfo)
	t.Log(returnPendingCommandData)
	t.Log(expectedInteractionCommand)

	assert.Equal(t, reflect.DeepEqual(returnPendingCommandData, expectedInteractionCommand), true)
}

func TestParseManualOutArgsToResponse(t *testing.T) {
	manualHandler := NewManualHandler(&mock_manual_inbox_storage.MockInboxStorage{})

	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"
	testStepExecId := "71a4d52c-6efc-4516-a242-dfbc5c89d999"

	metadata := run.Metadata{
		PlaybookId: testPlaybookId,
		RunId:      uuid.MustParse(testExecId),
		StepId:     testStepId,
		StepRunId:  uuid.MustParse(testStepExecId),
	}

	outVariable := cacao.Variable{Type: "string", Name: "__test__", Value: "updated!"}
	outVariables := map[string]cacao.Variable{"__test__": outVariable}

	payload := api.ManualOutArgsUpdatePayload{
		Type:            "out-args-update",
		ResponseStatus:  manual.ManualResponseFailureStatus,
		ResponseOutArgs: outVariables,
	}

	expetedResponse := manual.Response{
		Metadata:         metadata,
		ResponseStatus:   manual.ManualResponseFailureStatus,
		OutArgsVariables: outVariables,
		ResponseError:    nil,
	}

	response := manualHandler.parseManualOutArgsToResponse(metadata, payload)

	assert.Equal(t, expetedResponse, response)
}
