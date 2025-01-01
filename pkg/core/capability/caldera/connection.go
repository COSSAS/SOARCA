package caldera

import (
	"soarca/pkg/core/capability/caldera/api/client/abilities"
	"soarca/pkg/core/capability/caldera/api/client/adversaries"
	"soarca/pkg/core/capability/caldera/api/client/operationsops"
	"soarca/pkg/core/capability/caldera/api/models"
)

type ICalderaConnectionFactory interface {
	Create() (ICalderaConnection, error)
}
type calderaConnectionFactory struct{}

func (c calderaConnectionFactory) Create() (ICalderaConnection, error) {
	var instance, err = GetCalderaInstance()
	if err != nil {
		return nil, err
	}
	return &calderaConnection{instance}, nil
}

type ICalderaConnection interface {
	CreateAbility(ability *models.Ability) (string, error)
	DeleteAbility(abilityId string) error
	CreateOperation(agentGroupId string, adversaryId string) (string, error)
	IsOperationFinished(operationId string) (bool, error)
	RequestFacts(operationId string) ([]*models.PartialLink, error)
	CreateAdversary(abilityId string) (string, error)
}

type calderaConnection struct {
	instance *calderaInstance
}

type CalderaFacts map[string]string

func (cc calderaConnection) CreateAbility(ability *models.Ability) (string, error) {
	response, err := cc.instance.send.Abilities.PostAPIV2Abilities(
		abilities.NewPostAPIV2AbilitiesParams().WithBody(ability),
		authenticateCaldera,
	)
	if err != nil {
		return "", err
	}
	return response.GetPayload().AbilityID, nil
}

func (cc calderaConnection) DeleteAbility(abilityId string) error {
	_, err := cc.instance.send.Abilities.DeleteAPIV2AbilitiesAbilityID(
		abilities.NewDeleteAPIV2AbilitiesAbilityIDParams().WithAbilityID(abilityId),
		authenticateCaldera,
	)
	return err
}

func (cc calderaConnection) CreateOperation(
	agentGroupId string,
	adversaryId string,
) (string, error) {
	name := "SOARCA operation"
	response, err := cc.instance.send.Operationsops.PostAPIV2Operations(
		operationsops.NewPostAPIV2OperationsParams().WithBody(&models.Operation{
			Adversary: &models.Adversary{
				AdversaryID: adversaryId,
				AtomicOrdering: []string{},
			},
			Group:      agentGroupId,
			Autonomous: 1,
			AutoClose:  true,
			Name: &name,
		}),
		authenticateCaldera,
	)
	if err != nil {
		return "", err
	}
	return response.GetPayload().ID, nil
}

func (cc calderaConnection) CreateAdversary(abilityId string) (string, error) {
	response, err := cc.instance.send.Adversaries.PostAPIV2Adversaries(
		adversaries.NewPostAPIV2AdversariesParams().WithBody(&models.Adversary{
			Name: "SOARCA adversary",
			AtomicOrdering: []string{abilityId},
		}),
		authenticateCaldera,
	)
	if err != nil {
		return "", err
	}
	return response.GetPayload().AdversaryID, nil
}

func (cc calderaConnection) IsOperationFinished(operationId string) (bool, error) {
	response, err := cc.instance.send.Operationsops.GetAPIV2OperationsID(
		operationsops.NewGetAPIV2OperationsIDParams().WithID(operationId),
		authenticateCaldera,
	)
	if err != nil {
		return false, err
	}
	if response.GetPayload().State == "finished" {
		return true, nil
	}
	return false, nil
}

func (cc calderaConnection) RequestFacts(operationId string) ([]*models.PartialLink, error) {
	response, err := cc.instance.send.Operationsops.GetAPIV2OperationsIDLinks(
		operationsops.NewGetAPIV2OperationsIDLinksParams().WithID(operationId),
		authenticateCaldera,
	)
	if err != nil {
		return nil, err
	}
	return response.GetPayload(), nil
}
