package cacao_test

import (
	"soarca/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestVariableMapMerging(t *testing.T) {
	base := cacao.VariableMap{
		"__var0__": {
			Value: "OLD",
		},
	}

	new := cacao.VariableMap{
		"__var1__": {
			Value: "NEW",
		},
	}

	merged := base.Merge(new)
	assert.Equal(t, merged["__var0__"].Value, "OLD")
	assert.Equal(t, merged["__var1__"].Value, "NEW")
}

func TestVariableMapMergingWithUpdate(t *testing.T) {
	base := cacao.VariableMap{
		"__var__": {
			Value: "OLD",
		},
	}

	// Parent variables are not replaced
	new := cacao.VariableMap{
		"__var__": {
			Value: "NEW",
		},
	}

	merged := base.Merge(new)
	assert.Equal(t, merged["__var__"].Value, "OLD")
}

func TestVariableMapMergingWithConstant(t *testing.T) {
	base := cacao.VariableMap{
		"__var__": {
			Value: "OLD",
		},
	}

	// Parent variables are replaced if the replacement is constant
	new := cacao.VariableMap{
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

	vars := cacao.VariableMap{
		"__var0__": {
			Value: "GO",
		},
	}

	replaced := vars.Replace(original)
	assert.Equal(t, replaced, "GO is GO")
}

func TestVariableMapStringReplaceMultiple(t *testing.T) {
	original := "__var0__:value is __var1__:value"

	vars := cacao.VariableMap{
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
