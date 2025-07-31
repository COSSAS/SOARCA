package cases

import (
	"soarca/pkg/integration/thehive/common/models"
	"soarca/pkg/models/cacao"
)

func CreateHiveObservables(variables cacao.Variables) models.Observables {
	observables := models.Observables{}
	for _, variable := range variables {
		varType := "other"
		if variable.Type == cacao.VariableTypeIpv4Address {
			varType = "ip"
		} else if variable.Type == cacao.VariableTypeMacAddress {
			varType = "other"
		} else {
			// don't add other variables other then ip and mac to observables
			continue
		}
		obs := models.Observable{DataType: varType,
			Data:    []string{variable.Value},
			Message: variable.Description,
		}
		observables[variable.Value] = obs
	}
	return observables
}
