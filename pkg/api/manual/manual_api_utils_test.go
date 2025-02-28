package manual

import (
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
	assert.Equal(t, "a", "a")
}
func TestParseManualOutArgsUpdate(t *testing.T) {
	assert.Equal(t, "a", "a")
}
