package cacao_test

import (
	"soarca/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestVariablesInsertNew(t *testing.T) {
	vars := cacao.NewVariables()
	inserted := vars.Insert(cacao.Variable{
		Name:  "__var0__",
		Value: "value",
	})
	variable, found := vars.Find("__var0__")

	assert.Equal(t, inserted, true)
	assert.Equal(t, variable.Value, "value")
	assert.Equal(t, found, true)
}

func TestVariablesInsertExisting(t *testing.T) {
	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "old",
	})
	inserted := vars.Insert(cacao.Variable{
		Name:  "__var0__",
		Value: "new",
	})
	variable, found := vars.Find("__var0__")

	assert.Equal(t, inserted, false)
	assert.Equal(t, variable.Value, "old")
	assert.Equal(t, found, true)
}

func TestVariablesInsertOrReplace(t *testing.T) {
	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "old",
	})
	replaced := vars.InsertOrReplace(cacao.Variable{
		Name:  "__var0__",
		Value: "new",
	})
	variable, found := vars.Find("__var0__")

	assert.Equal(t, replaced, true)
	assert.Equal(t, variable.Value, "new")
	assert.Equal(t, found, true)
}

func TestVariablesInsertRange(t *testing.T) {
	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "old",
	})
	vars.InsertRange(cacao.Variable{
		Name:  "__var0__",
		Value: "new",
	}, cacao.Variable{
		Name:  "__var1__",
		Value: "new",
	})
	var0, found0 := vars.Find("__var0__")
	var1, found1 := vars.Find("__var1__")

	assert.Equal(t, var0.Value, "old")
	assert.Equal(t, found0, true)
	assert.Equal(t, var1.Value, "new")
	assert.Equal(t, found1, true)
}

func TestVariablesMerge(t *testing.T) {
	base := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "OLD",
	})

	new := cacao.NewVariables(cacao.Variable{
		Name:  "__var1__",
		Value: "NEW",
	})

	merged := base.Merge(new)
	var0, ok0 := merged.Find("__var0__")
	var1, ok1 := merged.Find("__var1__")

	assert.Equal(t, var0.Value, "OLD")
	assert.Equal(t, ok0, true)
	assert.Equal(t, var1.Value, "NEW")
	assert.Equal(t, ok1, true)
}

func TestVariablesMergeWithUpdate(t *testing.T) {
	base := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "OLD",
	})

	new := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "NEW",
	})

	merged := base.Merge(new)
	var0, ok0 := merged.Find("__var0__")

	assert.Equal(t, var0.Value, "NEW")
	assert.Equal(t, ok0, true)
}

func TestVariablesStringInterpolation(t *testing.T) {
	original := "__var0__:value is __var0__:value"

	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "GO is GO")
}

func TestVariablesStringInterpolateMultiple(t *testing.T) {
	original := "__var0__:value is __var1__:value"

	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	}, cacao.Variable{
		Name:  "__var1__",
		Value: "COOL",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "GO is COOL")
}

func TestVariablesSelect(t *testing.T) {
	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	}, cacao.Variable{
		Name:  "__var1__",
		Value: "COOL",
	})
	filteredVars := vars.Select([]string{"__var0__", "__unknown__"})

	assert.Equal(t, filteredVars, cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	}))
}
