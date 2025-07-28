package incident

import (
	"fmt"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporter/downstream_reporter/correlation/observable"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestCreateCase(t *testing.T) {

	caseId, _ := uuid.NewV7()
	var1 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	variables := cacao.NewVariables(var1)
	meta := execution.Metadata{ExecutionId: uuid.MustParse("01983c5a-067f-7016-a29b-10801077e6d6"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}
	thisCase := New(caseId, meta, variables)

	fmt.Println(thisCase)

	assert.NotEqual(t, thisCase.GetId().String(), "")

}

func TestIfObservableIsInCase(t *testing.T) {

	ip1 := observable.Observable{Type: observable.SourceAddressIpv4,
		Name:  "__DESTINATION_IPV4__",
		Value: "10.0.0.1"}

	observables := make(map[string]observable.Observable)
	observables[ip1.Value] = ip1

	thisCase := Case{Observables: observables, FirstObserved: time.Now()}

	var1 := cacao.Variable{Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	result, err := thisCase.CheckObservableInCase(var1)

	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
}

func TestDouble(t *testing.T) {

	meta1 := execution.Metadata{ExecutionId: uuid.MustParse("01983c5a-067f-7016-a29b-10801077e6d6"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}

	ip1 := observable.Observable{Type: observable.SourceAddressIpv4,
		Name:       "__DESTINATION_IPV4__",
		Value:      "10.0.0.1",
		Executions: observable.Executions{meta1.ExecutionId: observable.Initial}}

	observables := make(map[string]observable.Observable)
	observables[ip1.Value] = ip1

	thisCase := Case{Observables: observables, FirstObserved: time.Now()}

	var1 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	result, err := thisCase.CheckObservableInCase(var1)

	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	meta2 := execution.Metadata{ExecutionId: uuid.MustParse("01983c5a-067f-7016-a29b-10801077e6d6"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}

	var2 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.2"}
	thisCase.AddIfNotInCase(meta2, var2)
	result, err = thisCase.CheckObservableInCase(var2)

	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
}
