package cacao_test

import (
	"soarca/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestNewVariables(t *testing.T) {
	variable := cacao.Variable{
		Type:  "string",
		Name:  "variable 1",
		Value: "value 1",
	}
	variables := cacao.NewVariables(variable)
	expected := cacao.Variables{"variable 1": variable}
	assert.Equal(t, variables, expected)
}

func TestVariablesFind(t *testing.T) {
	variables := make(cacao.Variables)
	inserted := cacao.Variable{
		Name:  "__var0__",
		Value: "value",
	}
	variables["__var0__"] = inserted
	variable, found := variables.Find("__var0__")
	assert.Equal(t, found, true)
	assert.Equal(t, variable, inserted)
}

func TestVariablesInsertNew(t *testing.T) {
	vars := cacao.NewVariables()
	inserted := vars.Insert(cacao.Variable{
		Name:  "__var0__",
		Value: "value",
	})

	assert.Equal(t, inserted, true)
	assert.Equal(t, vars["__var0__"].Value, "value")
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

	assert.Equal(t, inserted, false)
	assert.Equal(t, vars["__var0__"].Value, "old")
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

	assert.Equal(t, replaced, true)
	assert.Equal(t, vars["__var0__"].Value, "new")
}

func TestVariablesInsertRange(t *testing.T) {
	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "old",
	})
	otherRange := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "new",
	}, cacao.Variable{
		Name:  "__var1__",
		Value: "new2",
	})
	vars.InsertRange(otherRange)

	assert.Equal(t, vars["__var0__"].Value, "old")
	assert.Equal(t, vars["__var1__"].Value, "new2")
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

	base.Merge(new)

	assert.Equal(t, base["__var0__"].Value, "OLD")
	assert.Equal(t, new["__var0__"].Value, "")
	assert.Equal(t, base["__var1__"].Value, "NEW")
	assert.Equal(t, new["__var1__"].Value, "NEW")
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

	base.Merge(new)

	assert.Equal(t, base["__var0__"].Value, "NEW")
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

func TestVariablesStringInterpolationEmptyString(t *testing.T) {
	original := ""

	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "")
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

func TestVariablesStringInterpolateMultipleAndUnkown(t *testing.T) {
	original := "__var0__:value is __var1_unknown__:value"

	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	}, cacao.Variable{
		Name:  "__var1__",
		Value: "COOL",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "GO is __var1_unknown__:value")
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

func TestInsertIntoEmptyMap(t *testing.T) {
	vars := cacao.NewVariables(cacao.Variable{
		Name:  "__var0__",
		Value: "GO",
	}, cacao.Variable{
		Name:  "__var1__",
		Value: "COOL",
	})

	playbook := cacao.NewPlaybook()

	playbook.PlaybookVariables.Merge(vars)
}
