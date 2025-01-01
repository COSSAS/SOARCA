package mock_caldera

import (
	"errors"
	"soarca/pkg/core/capability/caldera"
	"soarca/pkg/core/capability/caldera/api/models"
)

type MockCalderaConnectionFactory struct{}
type MockCalderaConnection struct{}

func (f MockCalderaConnectionFactory) Create() (caldera.ICalderaConnection, error) {
	return &MockCalderaConnection{}, nil
}

func (m MockCalderaConnection) CreateAbility(ability *models.Ability) (string, error) {
	return "abilityID", nil
}

func (m MockCalderaConnection) CreateAdversary(abilityID string) (string, error) {
	return "adversaryID", nil
}

func (m MockCalderaConnection) CreateOperation(
	agentGroupId string,
	abilityId string,
) (string, error) {
	return "operationID", nil
}

func (m MockCalderaConnection) DeleteAbility(abilityId string) error {
	return nil
}

func (m MockCalderaConnection) IsOperationFinished(operationId string) (bool, error) {
	return true, nil
}

func (m MockCalderaConnection) RequestFacts(operationId string) ([]*models.PartialLink, error) {
	return make([]*models.PartialLink, 0), nil
}

type MockBadCalderaConnectionFactory struct{}
type MockBadCalderaConnection struct{}

func (f MockBadCalderaConnectionFactory) Create() (caldera.ICalderaConnection, error) {
	return &MockBadCalderaConnection{}, nil
}

func (m MockBadCalderaConnection) CreateAbility(ability *models.Ability) (string, error) {
	return "", errors.New("Error Creating Ability")
}

func (m MockBadCalderaConnection) CreateAdversary(abilityId string) (string, error) {
	return "", errors.New("Error Creating Adversary")
}

func (m MockBadCalderaConnection) CreateOperation(
	agentGroupId string,
	abilityId string,
) (string, error) {
	return "", errors.New("Error Creating Operation")
}

func (m MockBadCalderaConnection) DeleteAbility(abilityId string) error {
	return errors.New("Error Deleting Ability")
}

func (m MockBadCalderaConnection) IsOperationFinished(operationId string) (bool, error) {
	return false, errors.New("Error Fetching Finished")
}

func (m MockBadCalderaConnection) RequestFacts(operationId string) ([]*models.PartialLink, error) {
	return make([]*models.PartialLink, 0), errors.New("Error Requesting Facts")
}
