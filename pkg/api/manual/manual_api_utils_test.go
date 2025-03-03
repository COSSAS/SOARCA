package manual

import (
	"errors"
	"reflect"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/api"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/models/manual"
	"soarca/test/unittest/mocks/mock_interaction_storage"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestParseManualOutArgsUpdate(t *testing.T) {
	manualHandler := NewManualHandler(&mock_interaction_storage.MockInteractionStorage{})

	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"

	jsonPayload := `{"type":"out-args-update","execution_id":"50b6d52c-6efc-4516-a242-dfbc5c89d421","playbook_id":"21a4d52c-6efc-4516-a242-dfbc5c89d312","step_id":"61a4d52c-6efc-4516-a242-dfbc5c89d312","response_status":"success","response_out_args":{"__test__":{"type":"string","name":"__test__","value":"updated!"}}}`
	bytesPayload := []byte(jsonPayload)

	outVariable := cacao.Variable{Type: "string", Name: "__test__", Value: "updated!"}
	outVariables := map[string]cacao.Variable{"__test__": outVariable}

	expectedPayload := api.ManualOutArgsUpdatePayload{
		Type:            "out-args-update",
		ExecutionId:     testExecId,
		PlaybookId:      testPlaybookId,
		StepId:          testStepId,
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
	manualHandler := NewManualHandler(&mock_interaction_storage.MockInteractionStorage{})

	jsonPayload := `{"type":"out-args-update","execution_id":"50b6d52c-6efc-4516-a242-dfbc5c89d421","playbook_id":"21a4d52c-6efc-4516-a242-dfbc5c89d312","step_id":"61a4d52c-6efc-4516-a242-dfbc5c89d312","response_status":"success","response_out_args":{"__test__":{"type":"string","name":"__wrong_name__","value":"updated!"}}}`
	bytesPayload := []byte(jsonPayload)

	expecedErr := errors.New("variable name mismatch for variable __test__: has different name property: __wrong_name__")
	_, err := manualHandler.parseManualOutArgsUpdate(bytesPayload)
	if err == nil {
		t.Log("an error for non-matching variables names should have been raised")
		t.Fail()
	}

	assert.Equal(t, err, expecedErr)
}

func TestParseCommandInfoToResponse(t *testing.T) {

	manualHandler := NewManualHandler(&mock_interaction_storage.MockInteractionStorage{})

	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"

	command := cacao.Command{Type: "manual", Command: "please do a test thanks", Description: "testing!"}
	target := cacao.AgentTarget{Type: "target", Name: "myself"}
	variable2 := cacao.Variable{Type: "string", Name: "__test__", Value: "some value"}
	inputVariable := map[string]cacao.Variable{"__test__": variable2}

	context := capability.Context{
		Command:   command,
		Target:    target,
		Variables: inputVariable,
	}

	testVariables := cacao.NewVariables(cacao.Variable{Type: "string", Name: "__test__", Value: "test!"})

	commandInfo := manual.CommandInfo{
		Metadata: execution.Metadata{
			PlaybookId:  testPlaybookId,
			ExecutionId: uuid.MustParse(testExecId),
			StepId:      testStepId},
		Context:          context,
		OutArgsVariables: testVariables,
	}

	expectedInteractionCommand := api.InteractionCommandData{
		Type:            "manual-command-info",
		ExecutionId:     testExecId,
		PlaybookId:      testPlaybookId,
		StepId:          testStepId,
		Description:     "testing!",
		Command:         "please do a test thanks",
		CommandIsBase64: false,
		Target:          target,
		OutVariables:    testVariables,
	}

	returnInteractionCommandData := manualHandler.parseCommandInfoToResponse(commandInfo)
	t.Log(returnInteractionCommandData)
	t.Log(expectedInteractionCommand)

	assert.Equal(t, reflect.DeepEqual(returnInteractionCommandData, expectedInteractionCommand), true)
}

func TestParseManualOutArgsToInteractionResponse(t *testing.T) {
	manualHandler := NewManualHandler(&mock_interaction_storage.MockInteractionStorage{})

	testExecId := "50b6d52c-6efc-4516-a242-dfbc5c89d421"
	testStepId := "61a4d52c-6efc-4516-a242-dfbc5c89d312"
	testPlaybookId := "21a4d52c-6efc-4516-a242-dfbc5c89d312"

	outVariable := cacao.Variable{Type: "string", Name: "__test__", Value: "updated!"}
	outVariables := map[string]cacao.Variable{"__test__": outVariable}

	payload := api.ManualOutArgsUpdatePayload{
		Type:            "out-args-update",
		ExecutionId:     testExecId,
		PlaybookId:      testPlaybookId,
		StepId:          testStepId,
		ResponseStatus:  manual.ManualResponseFailureStatus,
		ResponseOutArgs: outVariables,
	}

	expetedInteractionResponse := manual.InteractionResponse{
		Metadata: execution.Metadata{
			PlaybookId:  testPlaybookId,
			ExecutionId: uuid.MustParse(testExecId),
			StepId:      testStepId,
		},
		ResponseStatus:   manual.ManualResponseFailureStatus,
		OutArgsVariables: outVariables,
		ResponseError:    nil,
	}

	interactionResponse, err := manualHandler.parseManualOutArgsToInteractionResponse(payload)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	assert.Equal(t, expetedInteractionResponse, interactionResponse)

}
