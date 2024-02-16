package stix_test

import (
	"soarca/internal/stix"
	"soarca/models/cacao"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestPatternBasicMatch(t *testing.T) {
	p := stix.StixPattern{Pattern: "'a' = 'a'"}

	isTrue, err := p.IsTrue(cacao.Variables{})
	assert.Equal(t, isTrue, true)
	assert.Equal(t, err, nil)
}

func TestPatternBasicMismatch(t *testing.T) {
	p := stix.StixPattern{Pattern: "'a' = 'b'"}

	isTrue, err := p.IsTrue(cacao.Variables{})
	assert.Equal(t, isTrue, false)
	assert.Equal(t, err, nil)
}

func TestPatternBasicNeq(t *testing.T) {
	p := stix.StixPattern{Pattern: "'a' != 'b'"}

	isTrue, err := p.IsTrue(cacao.Variables{})
	assert.Equal(t, isTrue, true)
	assert.Equal(t, err, nil)
}

func TestPatternBasicNeqMismatch(t *testing.T) {
	p := stix.StixPattern{Pattern: "'a' != 'a'"}

	isTrue, err := p.IsTrue(cacao.Variables{})
	assert.Equal(t, isTrue, false)
	assert.Equal(t, err, nil)
}

func TestPatternVariableSubstitution(t *testing.T) {
	p := stix.StixPattern{Pattern: "__var0__:value = 'a'"}

	isTrue, err := p.IsTrue(cacao.Variables{"__var0__": cacao.Variable{Value: "a"}})
	assert.Equal(t, isTrue, true)
	assert.Equal(t, err, nil)
}

func TestPatternMultipleVariableSubstitution(t *testing.T) {
	p := stix.StixPattern{Pattern: "__var0__:value = __var1__:value"}

	isTrue, err := p.IsTrue(cacao.Variables{"__var0__": cacao.Variable{Value: "a"}, "__var1__": cacao.Variable{Value: "a"}})
	assert.Equal(t, isTrue, true)
	assert.Equal(t, err, nil)
}

func TestInvalidPatternTerms(t *testing.T) {
	p := stix.StixPattern{Pattern: "'a' = 'a' = 'a'"}

	_, err := p.IsTrue(cacao.Variables{})
	assert.NotEqual(t, err, nil)
}

func TestInvalidPatternOperator(t *testing.T) {
	p := stix.StixPattern{Pattern: "'a' ~= 'a'"}

	_, err := p.IsTrue(cacao.Variables{})
	assert.NotEqual(t, err, nil)
}
