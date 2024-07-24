package mapper_test

import (
	"errors"
	"soarca/models/cacao"
	"soarca/utils/mapper"
	"testing"

	"github.com/go-playground/assert/v2"
)

const variableValue = "some value"

func TestVariables(t *testing.T) {
	var1 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var1__"}
	var2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var2__"}
	var3 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var3__"}

	stepVariables := cacao.NewVariables(var1, var2, var3)
	out_args := []string{"__var1__"}
	capArgs := []string{"__cap1__"}
	capabilityResult := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString, Name: "__cap1__", Value: variableValue})

	output, err := mapper.Variables(stepVariables, out_args, capabilityResult, capArgs)
	assert.Equal(t, err, nil)
	variable, found := output.Find("__var1__")
	assert.Equal(t, found, true)
	assert.Equal(t, variable.Value, variableValue)
}

func TestVariablesOutArgsNotInStepVariables(t *testing.T) {
	var1 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var1__"}
	var2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var2__"}
	var3 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var3__"}

	stepVariables := cacao.NewVariables(var1, var2, var3)
	out_args := []string{"__varX__"}
	capArgs := []string{"__cap1__"}
	capabilityResult := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString, Name: "__cap1__", Value: variableValue})

	output, err := mapper.Variables(stepVariables, out_args, capabilityResult, capArgs)
	assert.Equal(t, err, errors.New("key is not found in variables"))
	variable, found := output.Find("__var1__")
	assert.Equal(t, found, false)
	assert.Equal(t, variable.Value, "")

}

func TestVariablesKeyNotFound(t *testing.T) {
	var1 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var1__"}
	var2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var2__"}
	var3 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var3__"}

	stepVariables := cacao.NewVariables(var1, var2, var3)
	out_args := []string{"__var1__"}
	dataArgs := []string{"__cap2__"}
	capabilityResult := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString, Name: "__cap1__", Value: variableValue})

	output, err := mapper.Variables(stepVariables, out_args, capabilityResult, dataArgs)
	assert.Equal(t, err, errors.New("key is not found in result data"))
	variable, found := output.Find("__var1__")
	assert.Equal(t, found, false)
	assert.Equal(t, variable.Value, "")

}

func TestNoOutArgsResultInEmptyVariable(t *testing.T) {
	var1 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var1__"}
	var2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var2__"}
	var3 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var3__"}

	stepVariables := cacao.NewVariables(var1, var2, var3)
	out_args := []string{}
	capArgs := []string{"__cap2__"}
	capabilityResult := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString, Name: "__cap1__", Value: variableValue})

	output, err := mapper.Variables(stepVariables, out_args, capabilityResult, capArgs)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(output), 0)

}

func TestMismatchVariableType(t *testing.T) {
	var1 := cacao.Variable{Type: cacao.VariableTypeBool, Name: "__var1__"}
	var2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var2__"}
	var3 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var3__"}

	stepVariables := cacao.NewVariables(var1, var2, var3)
	out_args := []string{"__var1__"}
	capArgs := []string{"__cap1__"}
	capabilityResult := cacao.NewVariables(cacao.Variable{Type: cacao.VariableTypeString, Name: "__cap1__", Value: variableValue})

	output, err := mapper.Variables(stepVariables, out_args, capabilityResult, capArgs)
	assert.Equal(t, err, errors.New("variable __var1__ is not of expected type string"))
	assert.Equal(t, len(output), 0)

}

func TestMultipleArgs(t *testing.T) {
	var1 := cacao.Variable{Type: cacao.VariableTypeBool, Name: "__var1__"}
	var2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__var2__"}
	var3 := cacao.Variable{Type: cacao.VariableTypeUri, Name: "__var3__"}

	stepVariables := cacao.NewVariables(var1, var2, var3)
	out_args := []string{"__var1__", "__var2__", "__var3__"}
	capArgs := []string{"__cap1__", "__cap2__", "__cap3__"}
	cap1 := cacao.Variable{Type: cacao.VariableTypeBool, Name: "__cap1__", Value: "True"}
	cap2 := cacao.Variable{Type: cacao.VariableTypeString, Name: "__cap2__", Value: "some string"}
	cap3 := cacao.Variable{Type: cacao.VariableTypeUri, Name: "__cap3__", Value: "https://example.com"}
	capabilityResult := cacao.NewVariables(cap1, cap2, cap3)

	output, err := mapper.Variables(stepVariables, out_args, capabilityResult, capArgs)
	assert.Equal(t, err, nil)
	assert.Equal(t, len(output), 3)
	outVar1, found1 := output.Find("__var1__")
	assert.Equal(t, found1, true)
	assert.Equal(t, outVar1.Value, cap1.Value)

	outVar2, found2 := output.Find("__var2__")
	assert.Equal(t, found2, true)
	assert.Equal(t, outVar2.Value, cap2.Value)
	outVar3, found3 := output.Find("__var3__")
	assert.Equal(t, found3, true)
	assert.Equal(t, outVar3.Value, cap3.Value)

}
