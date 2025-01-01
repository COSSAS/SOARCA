package caldera_test

import (
	"github.com/go-playground/assert/v2"
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	calderaModels "soarca/pkg/core/capability/caldera/api/models"
	"soarca/pkg/core/capability/caldera"
	mock_caldera "soarca/test/unittest/mocks/mock_utils/caldera"
	"testing"
)

func TestCapabilityName(t *testing.T) {
	calderaCapability := caldera.New(&mock_caldera.MockCalderaConnectionFactory{})
	assert.Equal(t, calderaCapability.GetType(), "soarca-caldera-cmd")
}

func TestExecute(t *testing.T) {
	calderaCapability := caldera.New(&mock_caldera.MockCalderaConnectionFactory{})

	results, err := calderaCapability.Execute(
		execution.Metadata{},
		capability.Context{
			Command: cacao.Command{Type: "caldera-cmd", Command: "id: abilityID"},
		},
	)

	if err != nil {
		t.Fail()
	}

	assert.Equal(t, cacao.NewVariables(), results)
}

func TestExecuteB64(t *testing.T) {
	calderaCapability := caldera.New(&mock_caldera.MockCalderaConnectionFactory{})

	results, err := calderaCapability.Execute(
		execution.Metadata{},
		capability.Context{
			Command: cacao.Command{Type: "caldera-cmd", CommandB64: "e30="},
		},
	)

	if err != nil {
		t.Fail()
	}

	assert.Equal(t, cacao.NewVariables(), results)
}

func TestParseYamlAbility(t *testing.T) {
	resultingAbility := caldera.ParseYamlAbility([]byte("ability_id: 9a30740d-3aa8-4c23-8efa-d51215e8a5b9"))
	assert.Equal(t, resultingAbility.AbilityID, "9a30740d-3aa8-4c23-8efa-d51215e8a5b9")
}

func TestParseYamlAbilityWithException(t *testing.T) {
	// This should not crash, just produce an empty Ability
	resultingAbility := caldera.ParseYamlAbility([]byte("  / very %$#% invalid yaml"))
	assert.Equal(t, resultingAbility.AbilityID, "")
}

func TestExecuteErrorConnection(t *testing.T) {
	calderaCapability := caldera.New(&mock_caldera.MockBadCalderaConnectionFactory{})

	_, err := calderaCapability.Execute(
		execution.Metadata{},
		capability.Context{
			Command: cacao.Command{Type: "caldera-cmd", CommandB64: "e30="},
		},
	)

	if err == nil {
		t.Fail()
	}

	assert.Equal(t, "Error Creating Ability", err.Error())
}

func TestGetCalderaInstance(t *testing.T) {
	_, err := caldera.GetCalderaInstance()
	assert.Equal(t, err, nil)
}


func TestSomethingElse(t *testing.T) {
	capability := caldera.New(nil)
	connection, err := capability.Factory.Create()
	assert.Equal(t, err, nil)

	_, err1 := connection.CreateOperation("agentGroup-0001", "adversary-0001")
	assert.NotEqual(t, err1, nil)

	_, err2 := connection.CreateAdversary("ability-0001")
	assert.NotEqual(t, err2, nil)

	_, err3 := connection.IsOperationFinished("operation-0001")
	assert.NotEqual(t, err3, nil)

	_, err4 := connection.RequestFacts("operation-0001")
	assert.NotEqual(t, err4, nil)

	_, err5 := connection.CreateAbility(&calderaModels.Ability{})
	assert.NotEqual(t, err5, nil)

	connection.DeleteAbility("ability-0001")

}
