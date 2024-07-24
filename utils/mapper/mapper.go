package mapper

import (
	"errors"
	"soarca/models/cacao"
)

func Variables(variables cacao.Variables,
	outArgs []string,
	data cacao.Variables,
	dataArgs []string) (cacao.Variables, error) {

	if len(outArgs) == 0 {
		return cacao.NewVariables(), nil
	}

	if len(outArgs) != len(dataArgs) {
		return cacao.NewVariables(), errors.New("number of outargs does not match data array length")
	}

	outputVariables := cacao.NewVariables()

	for i, arg := range outArgs {
		variable, found := variables.Find(arg)
		if !found {
			return cacao.NewVariables(), errors.New("key is not found in variables")
		}
		dataVariable, ok := data.Find(dataArgs[i])
		if !ok {
			return cacao.NewVariables(), errors.New("key is not found in result data")
		}
		if variable.Type != dataVariable.Type {
			return cacao.Variables{}, errors.New("variable " + variable.Name + " is not of expected type " + dataVariable.Type)
		}
		variable.Value = dataVariable.Value
		outputVariables.Insert(variable)
	}
	return outputVariables, nil

}
