package cases

import (
	"soarca/pkg/integration/thehive/common/models"
	"soarca/pkg/models/cacao"
)

func CreateHiveObservables(variables cacao.Variables) models.Observables {
	observables := models.Observables{}
	for _, variable := range variables {
		varType := ""

		switch variable.Type {
		case cacao.VariableTypeIpv4Address:
			varType = "ip"
		case cacao.VariableTypeMacAddress:
			varType = "other"
		default:
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
