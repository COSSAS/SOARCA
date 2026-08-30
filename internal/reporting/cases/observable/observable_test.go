package observable

import (
	"soarca/pkg/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/google/uuid"
)

func TestRuns(t *testing.T) {

	id, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs := Observable{Type: SourceAddressIpv4, Name: "__SRC_IP__", Value: "10.0.0.1", Runs: Runs{id: Initial}}

	id2, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs.AddRunToObservable(id2, SameObservable)

	assert.Equal(t, obs.Runs[id], Initial)
	assert.Equal(t, obs.Runs[id2], SameObservable)

}

func TestCreate(t *testing.T) {

	var1 := cacao.Variable{Type: cacao.VariableTypeIpv4Address, Name: "__SRC_IP__", Description: "source ip", Value: "10.0.0.1"}

	obs, err := CreateObservable(var1)
	assert.Equal(t, err, nil)

	id2, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs.AddRunToObservable(id2, SameObservable)
	assert.Equal(t, obs.Runs[id2], SameObservable)

}

func TestMatchSame(t *testing.T) {
	id, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs := Observable{Type: SourceAddressIpv4, Name: "__SRC_IP__", Value: "10.0.0.1", Runs: Runs{id: Initial}}

	id2, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs2 := Observable{Type: SourceAddressIpv4, Name: "__SRC_IP__", Value: "10.0.0.1", Runs: Runs{id2: Initial}}

	err = obs.Match(obs2, id2)
	assert.Equal(t, err, nil)

}

func TestMatchLateral(t *testing.T) {
	id, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs := Observable{Type: DestinationAddressIpv4, Name: "__DEST_IP__", Value: "10.0.0.1", Runs: Runs{id: Initial}}

	id2, err := uuid.NewV7()
	assert.Equal(t, err, nil)
	obs2 := Observable{Type: SourceAddressIpv4, Name: "__SRC_IP__", Value: "10.0.0.1", Runs: Runs{id2: Initial}}
	err = obs.Match(obs2, id2)
	assert.Equal(t, err, nil)
}
