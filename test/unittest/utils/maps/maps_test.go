package maps_test

import (
	"soarca/models/cacao"
	"soarca/utils/maps"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestNew(t *testing.T) {
	vars := maps.New[cacao.AgentTarget]()
	t.Log(vars)
}

func TestInsert(t *testing.T) {
	vars := maps.New[cacao.AgentTarget]()
	agent := cacao.AgentTarget{}
	res := maps.Insert(vars, "agent1", agent)
	t.Log(vars)
	assert.Equal(t, res, true)
}

func TestMerge(t *testing.T) {
	vars := maps.New[cacao.AgentTarget]()
	maps.Insert(vars, "agent1", cacao.AgentTarget{ID: "1"})
	vars2 := maps.New[cacao.AgentTarget]()
	maps.Insert(vars2, "agent2", cacao.AgentTarget{ID: "2"})

	maps.Merge(vars, vars2)
	t.Log(vars)
}
