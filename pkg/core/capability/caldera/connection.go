// This file bundles the abstract connection request methods used by the main business logic of the Caldera capability.
package caldera

import (
	"soarca/pkg/core/capability/caldera/api/client/abilities"
	"soarca/pkg/core/capability/caldera/api/client/adversaries"
	"soarca/pkg/core/capability/caldera/api/client/operationsops"
	"soarca/pkg/core/capability/caldera/api/models"
)

// ICalderaConnectionFactory is a factory struct that defines a Create method to construct a specific type of ICalderaConnection.
type ICalderaConnectionFactory interface {
	Create() (ICalderaConnection, error)
}

// calderaConnectionFactory is the default ICalderaConnectionFactory and builds a calderaConnection.
type calderaConnectionFactory struct{}

// Create builds a calderaConnection.
// It first tries to get the Caldera server instance, and returns a calderaConnection struct connected to the instance.
//
// If it fails to retrieve the Caldera instance, it returns nil with the error.
func (c calderaConnectionFactory) Create() (ICalderaConnection, error) {
	var instance, err = GetCalderaInstance()
	if err != nil {
		return nil, err
	}
	return &calderaConnection{instance}, nil
}

// ICalderaConnection is the inferface that defines the higher level abstract connection requests.
//
// These methods are used for the main business logic of the Caldera capability.
// Each method creates a request to the Caldera instance and returns either the parsed response or an error if something went wrong.
type ICalderaConnection interface {
	CreateAbility(ability *models.Ability) (string, error)
	DeleteAbility(abilityId string) error
	CreateOperation(agentGroupId string, adversaryId string) (string, error)
	IsOperationFinished(operationId string) (bool, error)
	RequestFacts(operationId string) ([]*models.PartialLink, error)
	CreateAdversary(abilityId string) (string, error)
}

// calderaConnection is the default Caldera connection struct built by the calderaConnectionFactory.
// It contains a pointer to the Caldera instance it is connected to.
//
// This connection is used for the actual connection requests to a Caldera instance.
type calderaConnection struct {
	instance *calderaInstance
}

// CalderaFacts is a map between the fact name as string to the value, also as a string.
//
// Plans are to be able to have different types of values, like integers or arrays.
type CalderaFacts map[string]string

// CreateAbility initiates a request to the Caldera instance to create a Caldera Ability.
//
// It returns the id of the created ability, or an error if it fails.
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

// DeleteAbility initiates a request to the Caldera instance to delete a Caldera Ability.
//
// It returns an error if it fails.
func (cc calderaConnection) DeleteAbility(abilityId string) error {
	_, err := cc.instance.send.Abilities.DeleteAPIV2AbilitiesAbilityID(
		abilities.NewDeleteAPIV2AbilitiesAbilityIDParams().WithAbilityID(abilityId),
		authenticateCaldera,
	)
	return err
}

// CreateOperation initiates a request to the Caldera instance to create a Caldera Operation.
//
// It returns the id of the created operation, or an error if it fails.
func (cc calderaConnection) CreateOperation(
	agentGroupId string,
	adversaryId string,
) (string, error) {
	name := "SOARCA operation"
	response, err := cc.instance.send.Operationsops.PostAPIV2Operations(
		operationsops.NewPostAPIV2OperationsParams().WithBody(&models.Operation{
			Adversary: &models.Adversary{
				AdversaryID:    adversaryId,
				AtomicOrdering: []string{},
			},
			Group:      agentGroupId,
			Autonomous: 1,
			AutoClose:  true,
			Name:       &name,
		}),
		authenticateCaldera,
	)
	if err != nil {
		return "", err
	}
	return response.GetPayload().ID, nil
}

// CreateAdversary initiates a request to the Caldera instance to create a Caldera Adversary.
//
// It returns the id of the created adversary, or an error if it fails.
func (cc calderaConnection) CreateAdversary(abilityId string) (string, error) {
	response, err := cc.instance.send.Adversaries.PostAPIV2Adversaries(
		adversaries.NewPostAPIV2AdversariesParams().WithBody(&models.Adversary{
			Name:           "SOARCA adversary",
			AtomicOrdering: []string{abilityId},
		}),
		authenticateCaldera,
	)
	if err != nil {
		return "", err
	}
	return response.GetPayload().AdversaryID, nil
}

// IsOperationFinished initiates a request to the Caldera instance to check if a given Caldera Operation is finished.
//
// It returns true if the operation is finished, false if it is not, or an error if it fails.
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

// RequestFacts initiates a request to the Caldera instance to get the facts of a given Caldera Operation.
//
// It returns a list of Caldera links, which contain the Caldera facts, or an error if it fails.
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
