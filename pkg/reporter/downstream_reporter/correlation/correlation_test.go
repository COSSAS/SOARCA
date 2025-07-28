package correlation

import (
	"fmt"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
	"soarca/pkg/reporter/downstream_reporter/correlation/incident"
	"soarca/pkg/reporter/downstream_reporter/correlation/observable"
	"soarca/test/unittest/mocks/mock_guid"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestCorrelation(t *testing.T) {
	guid := mock_guid.Mock_Guid{}
	correlation := New(&guid)
	var1 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	variables := cacao.NewVariables(var1)
	meta := execution.Metadata{ExecutionId: uuid.MustParse("01983c7b-2983-7b7d-b3bb-548d2b5d7a2a"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}

	id, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	guid.On("NewV7").Return(id).Times(1)
	caseId, err := correlation.Correlate(meta, variables)
	assert.Equal(t, caseId, id.String())
	assert.Equal(t, err, nil)
	assert.Equal(t, correlation.cases[id].GetId(), id)
	assert.Equal(t, len(correlation.cases[id].GetObservables()), 1)

	fmt.Println(correlation.cases)
}

func TestAddSecondCorrelation(t *testing.T) {
	guid := mock_guid.Mock_Guid{}
	correlation := New(&guid)
	var1 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	variables := cacao.NewVariables(var1)
	meta := execution.Metadata{ExecutionId: uuid.MustParse("01983c7b-2983-7b7d-b3bb-548d2b5d7a2a"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}

	id, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	guid.On("NewV7").Return(id).Times(1)
	caseId, err := correlation.Correlate(meta, variables)
	assert.Equal(t, caseId, id.String())
	assert.Equal(t, err, nil)
	assert.Equal(t, correlation.cases[id].GetId(), id)
	for id, cas := range correlation.cases {
		fmt.Println(id, cas.GetId())
	}

	var2 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.2"}
	meta2 := execution.Metadata{ExecutionId: uuid.MustParse("01983caa-3cd6-704b-9598-ee9424166ead"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}
	variables2 := cacao.NewVariables(var2)
	id2, err := uuid.NewV7()

	assert.Equal(t, err, nil)
	guid.On("NewV7").Return(id2).Times(1)
	caseId2, err := correlation.Correlate(meta2, variables2)
	assert.Equal(t, caseId2, id2.String())
	assert.Equal(t, err, nil)
	for id, cas := range correlation.cases {
		fmt.Println(id, cas.GetId())
	}

	ob1 := observable.Observable{Type: observable.SourceAddressIpv4,
		Name:       var1.Name,
		Value:      var1.Value,
		Executions: observable.Executions{meta.ExecutionId: observable.Initial}}

	obs1 := incident.Observables{}
	obs1[ob1.Value] = ob1
	assert.Equal(t, correlation.cases[id].GetObservables(), obs1)
	assert.Equal(t, len(correlation.cases[id].GetObservables()), 1)

	ob2 := observable.Observable{Type: observable.SourceAddressIpv4,
		Name:       var2.Name,
		Value:      var2.Value,
		Executions: observable.Executions{meta2.ExecutionId: observable.Initial}}

	obs2 := incident.Observables{}
	obs2[ob2.Value] = ob2
	assert.Equal(t, correlation.cases[id2].GetObservables(), obs2)
	assert.Equal(t, len(correlation.cases[id2].GetObservables()), 1)
}

func TestMultipleAdditionsToMultipleCases(t *testing.T) {
	guid := mock_guid.Mock_Guid{}
	correlation := New(&guid)
	var1 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.1"}
	variables := cacao.NewVariables(var1)
	meta := execution.Metadata{ExecutionId: uuid.MustParse("01983c7b-2983-7b7d-b3bb-548d2b5d7a2a"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}

	id, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	guid.On("NewV7").Return(id).Times(1)
	caseId, err := correlation.Correlate(meta, variables)
	assert.Equal(t, caseId, id.String())
	assert.Equal(t, err, nil)
	assert.Equal(t, correlation.cases[id].GetId(), id)
	for id, cas := range correlation.cases {
		fmt.Println(id, cas.GetId())
	}

	var2 := cacao.Variable{Name: "__SOURCE_IPV4__", Type: cacao.VariableTypeIpv4Address, Value: "10.0.0.2"}
	meta2 := execution.Metadata{ExecutionId: uuid.MustParse("01983caa-3cd6-704b-9598-ee9424166ead"),
		PlaybookId: "some-playbook",
		StepId:     "SomeStep"}
	variables2 := cacao.NewVariables(var2)
	id2, err := uuid.NewV7()

	assert.Equal(t, err, nil)
	guid.On("NewV7").Return(id2).Times(1)
	caseId2, err := correlation.Correlate(meta2, variables2)
	assert.Equal(t, caseId2, id2.String())
	assert.Equal(t, err, nil)
	for id, cas := range correlation.cases {
		fmt.Println(id, cas.GetId())
	}

	// assert.Equal(t, correlation.cases[id], nil)
	assert.Equal(t, len(correlation.cases[id].GetObservables()), 1)
	assert.Equal(t, len(correlation.cases[id2].GetObservables()), 1)
}
