package keymanagementrepository

import (
	"errors"
	"soarca/internal/database"
	"soarca/pkg/models/keymanagement"
)

type IKeyManagementRepository interface {
	GetKeyNames() ([]string, error)
	Create(name string, keypair keymanagement.KeyPair) error
	Read(id string) (keymanagement.KeyPair, error)
	Update(id string, keypair keymanagement.KeyPair) error
	Delete(id string) error
}

type KeyManagementRepository struct {
	db      database.Database
	options database.FindOptions
}

type KeyPairEntry struct {
	name    string
	keypair keymanagement.KeyPair
}

func SetupKeyManagementRepository(db database.Database, options database.FindOptions) *KeyManagementRepository {
	return &KeyManagementRepository{db: db, options: options}
}

func (keymanagementRepo *KeyManagementRepository) GetKeyNames() ([]string, error) {
	keys, err := keymanagementRepo.db.Find(nil)
	if err != nil {
		return nil, err
	}
	ret := []string{}
	for _, key := range keys {
		ret = append(ret, key.(KeyPairEntry).name)
	}
	return ret, nil
}

func (keymanagementRepo *KeyManagementRepository) Create(name string, keypair keymanagement.KeyPair) error {
	return keymanagementRepo.db.Create(KeyPairEntry{name, keypair})
}

func (keymanagementRepo *KeyManagementRepository) Read(id string) (keymanagement.KeyPair, error) {
	returnedObject, err := keymanagementRepo.db.Read(id)
	if err != nil {
		return keymanagement.KeyPair{}, err
	}

	keypair, ok := returnedObject.(keymanagement.KeyPair)

	if !ok {
		err = errors.New("could not cast lookup object to keypair type")
		return keymanagement.KeyPair{}, err
	}

	return keypair, nil
}

func (keymanagementRepo *KeyManagementRepository) Update(id string, keypair keymanagement.KeyPair) error {
	return keymanagementRepo.db.Update(id, keypair)
}

func (keymanagementRepo *KeyManagementRepository) Delete(id string) error {
	return keymanagementRepo.db.Delete(id)
}
