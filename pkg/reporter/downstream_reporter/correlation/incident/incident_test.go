package incident

import (
	"fmt"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestCreateCase(t *testing.T) {

	var1 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	variables := cacao.NewVariables(var1)
	meta := execution.Metadata{ExecutionId: uuid.MustParse("01983c5a-067f-7016-a29b-10801077e6d6"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}
	thisCase := New(meta, variables)

	fmt.Println(thisCase)

	assert.NotEqual(t, thisCase.Id.String(), "")
	assert.Equal(t, thisCase.Observables["10.0.0.1"].Value, var1.Value)

}

func TestIfObservableIsInCase(t *testing.T) {

	ip1 := Observable{Type: SourceAddressIpv4,
		Name:  "__DESTINATION_IPV4__",
		Value: "10.0.0.1"}

	observables := make(map[string]Observable)
	observables[ip1.Value] = ip1

	thisCase := Case{Observables: observables, FirstObserved: time.Now()}

	var1 := cacao.Variable{Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	result, err := thisCase.CheckObservableInCase(var1)

	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
}
