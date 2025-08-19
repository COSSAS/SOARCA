package memory

import (
	"fmt"
	"soarca/pkg/models/keymanagement"
)

type InMemoryKeyManagementDatabase struct {
	keys map[string]keymanagement.KeyPair
}

func NewKeyManagementDatabase() *InMemoryKeyManagementDatabase {
	return &InMemoryKeyManagementDatabase{keys: make(map[string]keymanagement.KeyPair)}
}

func (database *InMemoryKeyManagementDatabase) GetKeyNames() ([]string, error) {
	ret := []string{}
	for key := range database.keys {
		ret = append(ret, key)
	}
	return ret, nil
}
func (database *InMemoryKeyManagementDatabase) Create(name string, keypair keymanagement.KeyPair) error {
	database.keys[name] = keypair
	return nil
}
func (database *InMemoryKeyManagementDatabase) Read(id string) (keymanagement.KeyPair, error) {
	keypair, ok := database.keys[id]
	if !ok {
		return keymanagement.KeyPair{}, fmt.Errorf("could not find key named %s", id)
	}
	return keypair, nil
}
func (database *InMemoryKeyManagementDatabase) Update(id string, keypair keymanagement.KeyPair) error {
	database.keys[id] = keypair
	return nil
}
func (database *InMemoryKeyManagementDatabase) Delete(id string) error {
	delete(database.keys, id)
	return nil
}
