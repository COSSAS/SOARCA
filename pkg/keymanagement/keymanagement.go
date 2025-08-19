package keymanagement

import (
	"fmt"
	"reflect"
	keymanagementrepository "soarca/internal/database/keymanagement"
	"soarca/internal/logger"
	keys "soarca/pkg/models/keymanagement"

	"golang.org/x/crypto/ssh"
)

var component = reflect.TypeOf(KeyManagement{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

type KeyManagement struct {
	database    keymanagementrepository.IKeyManagementRepository
	cached_keys map[string]keys.KeyPair
}

func InitKeyManagement(database keymanagementrepository.IKeyManagementRepository) *KeyManagement {
	return &KeyManagement{database: database, cached_keys: make(map[string]keys.KeyPair)}
}

func (management *KeyManagement) GetKeyPair(name string) (*keys.KeyPair, error) {
	log.Trace("Getting keypair named", name)
	keypair, ok := management.cached_keys[name]
	if !ok {
		keypair, err := management.database.Read(name)
		if err != nil {
			return nil, err
		}
		return &keypair, nil
	}
	return &keypair, nil
}
func (management *KeyManagement) GetPrivate(name string) (ssh.Signer, error) {
	keypair, err := management.GetKeyPair(name)
	if err != nil {
		return nil, err
	}
	return keypair.Private, nil
}

func (management *KeyManagement) Insert(public string, private string, passphrase string, name string) error {
	if _, err := management.GetKeyPair(name); err == nil {
		return fmt.Errorf("key with name already exists: %s (error: %s)", name, err)
	}
	return management.insertInternal(public, private, passphrase, name)
}

func (management *KeyManagement) Update(public string, private string, passphrase string, name string) error {
	if _, err := management.GetKeyPair(name); err != nil {
		return fmt.Errorf("no such key exists: %s (error: %s)", name, err)
	}
	return management.insertInternal(public, private, passphrase, name)
}

func (management *KeyManagement) insertInternal(public string, private string, passphrase string, name string) error {
	log.Trace("Inserting keypair named", name)
	public_key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(public))
	if err != nil {
		return fmt.Errorf("parsing public key: %s", err)
	}
	var private_key ssh.Signer
	if passphrase == "" {
		private_key, err = ssh.ParsePrivateKey([]byte(private))
	} else {
		private_key, err = ssh.ParsePrivateKeyWithPassphrase([]byte(private), []byte(passphrase))
	}
	if err != nil {
		return fmt.Errorf("parsing private key: %s", err)
	}
	keypair := keys.KeyPair{Public: public_key, Private: private_key}
	management.cached_keys[name] = keypair
	return management.database.Create(name, keypair)
}

func (management *KeyManagement) ListAllNames() []string {
	log.Trace("Listing all keys")
	keys := make([]string, len(management.cached_keys))
	index := 0
	for key := range management.cached_keys {
		keys[index] = key
		index++
	}
	return keys
}

func (management *KeyManagement) Revoke(keyname string) error {
	log.Trace("Deleting keypair named", keyname)
	err := management.database.Delete(keyname)
	delete(management.cached_keys, keyname)
	return err
}
