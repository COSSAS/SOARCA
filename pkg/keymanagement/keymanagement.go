package keymanagement

import (
	"fmt"
	"os"
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
func parsePrivateKey(filename string, passphrase string) (ssh.Signer, error) {
	log.Tracef("Opening private key from path %s", filename)
	file_buffer, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	if passphrase == "" {
		return ssh.ParsePrivateKey(file_buffer)
	} else {
		return ssh.ParsePrivateKeyWithPassphrase(file_buffer, []byte(passphrase))
	}
}
func parsePublicKey(filename string) (ssh.PublicKey, error) {
	log.Tracef("Opening public key from path %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	file_buffer := make([]byte, 2048)
	if _, err = file.Read(file_buffer); err != nil {
		return nil, err
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey(file_buffer)
	return key, err

}

func (management *KeyManagement) Insert(public string, private string, passphrase string, name string) error {
	if _, err := management.GetKeyPair(name); err == nil {
		return fmt.Errorf("Key with name already exists: %s (error: %s)", name, err)
	}
	return management.insertInternal(public, private, passphrase, name)
}

func (management *KeyManagement) Update(public string, private string, passphrase string, name string) error {
	if _, err := management.GetKeyPair(name); err != nil {
		return fmt.Errorf("No such key exists: %s (error: %s)", name, err)
	}
	return management.insertInternal(public, private, passphrase, name)
}

func (management *KeyManagement) insertInternal(public string, private string, passphrase string, name string) error {
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
	management.database.Create(name, keypair)
	return nil
}

func (management *KeyManagement) ListAllNames() []string {
	keys := make([]string, len(management.cached_keys))
	index := 0
	for key := range management.cached_keys {
		keys[index] = key
		index++
	}
	return keys
}

func (management *KeyManagement) Revoke(keyname string) {
	management.database.Delete(keyname)
	delete(management.cached_keys, keyname)
}
