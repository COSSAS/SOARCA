package cacao

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestNewVariables(t *testing.T) {
	variable := Variable{
		Type:  "string",
		Name:  "variable 1",
		Value: "value 1",
	}
	variables := NewVariables(variable)
	expected := Variables{"variable 1": variable}
	assert.Equal(t, variables, expected)
}

func TestVariablesFind(t *testing.T) {
	variables := make(Variables)
	inserted := Variable{
		Name:  "__var0__",
		Value: "value",
	}
	variables["__var0__"] = inserted
	variable, found := variables.Find("__var0__")
	assert.Equal(t, found, true)
	assert.Equal(t, variable, inserted)
}

func TestVariablesInsertNew(t *testing.T) {
	vars := NewVariables()
	inserted := vars.Insert(Variable{
		Name:  "__var0__",
		Value: "value",
	})

	assert.Equal(t, inserted, true)
	assert.Equal(t, vars["__var0__"].Value, "value")
}

func TestVariablesInsertExisting(t *testing.T) {
	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "old",
	})
	inserted := vars.Insert(Variable{
		Name:  "__var0__",
		Value: "new",
	})

	assert.Equal(t, inserted, false)
	assert.Equal(t, vars["__var0__"].Value, "old")
}

func TestVariablesInsertOrReplace(t *testing.T) {
	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "old",
	})
	replaced := vars.InsertOrReplace(Variable{
		Name:  "__var0__",
		Value: "new",
	})

	assert.Equal(t, replaced, true)
	assert.Equal(t, vars["__var0__"].Value, "new")
}

func TestVariablesInsertRange(t *testing.T) {
	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "old",
	})
	otherRange := NewVariables(Variable{
		Name:  "__var0__",
		Value: "new",
	}, Variable{
		Name:  "__var1__",
		Value: "new2",
	})
	vars.InsertRange(otherRange)

	assert.Equal(t, vars["__var0__"].Value, "old")
	assert.Equal(t, vars["__var1__"].Value, "new2")
}

func TestVariablesMerge(t *testing.T) {
	base := NewVariables(Variable{
		Name:  "__var0__",
		Value: "OLD",
	})

	new := NewVariables(Variable{
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
	base := NewVariables(Variable{
		Name:  "__var0__",
		Value: "OLD",
	})

	new := NewVariables(Variable{
		Name:  "__var0__",
		Value: "NEW",
	})

	base.Merge(new)

	assert.Equal(t, base["__var0__"].Value, "NEW")
}

func TestVariablesStringInterpolation(t *testing.T) {
	original := "__var0__:value is __var0__:value"

	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "GO is GO")
}

func TestVariablesStringInterpolationEmptyString(t *testing.T) {
	original := ""

	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "")
}

func TestVariablesStringInterpolateMultiple(t *testing.T) {
	original := "__var0__:value is __var1__:value"

	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	}, Variable{
		Name:  "__var1__",
		Value: "COOL",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "GO is COOL")
}

func TestVariablesStringInterpolateMultipleAndUnkown(t *testing.T) {
	original := "__var0__:value is __var1_unknown__:value"

	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	}, Variable{
		Name:  "__var1__",
		Value: "COOL",
	})

	replaced := vars.Interpolate(original)
	assert.Equal(t, replaced, "GO is __var1_unknown__:value")
}

func TestVariablesSelect(t *testing.T) {
	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	}, Variable{
		Name:  "__var1__",
		Value: "COOL",
	})
	filteredVars := vars.Select([]string{"__var0__", "__unknown__"})

	assert.Equal(t, filteredVars, NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	}))
}

func TestInsertIntoEmptyMap(t *testing.T) {
	vars := NewVariables(Variable{
		Name:  "__var0__",
		Value: "GO",
	}, Variable{
		Name:  "__var1__",
		Value: "COOL",
	})

	playbook := NewPlaybook()

	playbook.PlaybookVariables.Merge(vars)
}
