package cacao_test

import (
	"soarca/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestVariableMapMerging(t *testing.T) {
	base := cacao.Variables{
		"__var0__": {
			Value: "OLD",
		},
	}

	new := cacao.Variables{
		"__var1__": {
			Value: "NEW",
		},
	}

	merged := base.Merge(new)
	assert.Equal(t, merged["__var0__"].Value, "OLD")
	assert.Equal(t, merged["__var1__"].Value, "NEW")
}

func TestVariableMapMergingWithUpdate(t *testing.T) {
	base := cacao.Variables{
		"__var__": {
			Value: "OLD",
		},
	}

	// Parent variables are not replaced
	new := cacao.Variables{
		"__var__": {
			Value: "NEW",
		},
	}

	merged := base.Merge(new)
	assert.Equal(t, merged["__var__"].Value, "OLD")
}

func TestVariableMapMergingWithConstant(t *testing.T) {
	base := cacao.Variables{
		"__var__": {
			Value: "OLD",
		},
	}

	// Parent variables are replaced if the replacement is constant
	new := cacao.Variables{
		"__var__": {
			Value:    "NEW",
			Constant: true,
		},
	}

	merged := base.Merge(new)
	assert.Equal(t, merged["__var__"].Value, "NEW")
}

func TestVariableMapStringReplace(t *testing.T) {
	original := "__var0__:value is __var0__:value"

	vars := cacao.Variables{
		"__var0__": {
			Value: "GO",
		},
	}

	replaced := vars.Replace(original)
	assert.Equal(t, replaced, "GO is GO")
}

func TestVariableMapStringReplaceMultiple(t *testing.T) {
	original := "__var0__:value is __var1__:value"

	vars := cacao.Variables{
		"__var0__": {
			Value: "GO",
		},
		"__var1__": {
			Value: "COOL",
		},
	}

	replaced := vars.Replace(original)
	assert.Equal(t, replaced, "GO is COOL")
}

func TestVariableMapSelect(t *testing.T) {
	vars := cacao.Variables{
		"__var1__": {Value: "val1"},
		"__var2__": {Value: "val2"},
	}

	filteredVars := vars.Select([]string{"__var1__", "__unknown__"})

	assert.Equal(t, filteredVars, cacao.Variables{"__var1__": {Value: "val1"}})
}
