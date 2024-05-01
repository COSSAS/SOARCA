package stix_test

import (
	"errors"
	"soarca/models/cacao"
	"soarca/utils/stix"
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

func TestStringEquals(t *testing.T) {

	var1 := cacao.Variable{Type: cacao.VariableTypeString}
	var1.Value = "a"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = a", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
	result, err = stix.Evaluate("__var1__:value = b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value = 1", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value > b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value < b", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value <= b", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("a =  b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
	result, err = stix.Evaluate("a = b c", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
	time.Now()
}

func TestIntEquals(t *testing.T) {

	var1 := cacao.Variable{Type: cacao.VariableTypeLong}
	var1.Value = "1000"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = 1000", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
	result, err = stix.Evaluate("__var1__:value = 9999", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value = 10000", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value > 999", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value < 1001", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value <= 1000", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= 1000", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= a", vars)
	assert.Equal(t, result, false)
	assert.NotEqual(t, err, nil)

	result, err = stix.Evaluate("a =  b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
	result, err = stix.Evaluate("a = b c", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
}

// func TestNotEquals(t *testing.T) {
// 	result, err := stix.Evaluate("a != a")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("a != b")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// }

// func TestGreater(t *testing.T) {
// 	result, err := stix.Evaluate("2 > 1")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200000000 > 10000000")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 > 199")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("199 > 200")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// }

// func TestLess(t *testing.T) {
// 	result, err := stix.Evaluate("2 < 1")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200000000 < 10000000")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 < 199")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("199 < 200")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// }

// func TestLessEqual(t *testing.T) {
// 	result, err := stix.Evaluate("2 <= 1")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200000000 <= 10000000")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 <= 199")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("199 <= 200")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 <= 200")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// }

// func TestGreaterEqual(t *testing.T) {
// 	result, err := stix.Evaluate("2 >= 1")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("10000000 >= 200000000")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 >= 199")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("199 >= 200")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 >= 200")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// }

// func TestIn(t *testing.T) {
// 	result, err := stix.Evaluate("2 = (1,2,3,4)")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("10000000 >= 200000000")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 >= 199")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("199 >= 200")
// 	assert.Equal(t, result, false)
// 	assert.Equal(t, err, nil)
// 	result, err = stix.Evaluate("200 >= 200")
// 	assert.Equal(t, result, true)
// 	assert.Equal(t, err, nil)
// }
