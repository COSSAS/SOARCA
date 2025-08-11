package ssh

import (
	"fmt"
	"os"
	"path"
	"slices"
	"soarca/pkg/utils"
	"strings"

	"golang.org/x/crypto/ssh"
)

type KeyPair struct {
	Public  ssh.PublicKey
	Private ssh.Signer
}
type KeyManagement struct {
	underlying_dir string
	cached_keys    map[string]KeyPair
}

var keyManagement *KeyManagement

func InitKeyManagement(underlying_dir string) error {
	keyManagement = &KeyManagement{
		underlying_dir: underlying_dir,
	}
	return keyManagement.Refresh()
}

func (management *KeyManagement) GetKeyPair(name string) (*KeyPair, error) {
	keypair, ok := management.cached_keys[name]
	if !ok {
		return nil, fmt.Errorf("Could not find keypair for %s", name)
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
func parsePrivateKey(filename string) (ssh.Signer, error) {
	log.Tracef("Opening private key from path %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	passphrase := utils.GetEnv("SSH_KMS_PASSPHRASE", "")
	file_buffer := make([]byte, 2048)
	if _, err := file.Read(file_buffer); err != nil {
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
	if _, err := file.Read(file_buffer); err != nil {
		return nil, err
	}
	return ssh.ParsePublicKey(file_buffer)

}
func (management *KeyManagement) Refresh() error {
	management.cached_keys = make(map[string]KeyPair)
	dir, err := os.Open(management.underlying_dir)
	if err != nil {
		return err
	}
	filenames, err := dir.Readdirnames(0)
	if err != nil {
		return err
	}
	for _, filename := range filenames {
		if strings.HasSuffix(filename, ".pub") {
			private_filename := strings.TrimSuffix(filename, ".pub")
			if !slices.Contains(filenames, private_filename) {
				return fmt.Errorf("found public key %s without corresponding private key (%s)", filename, private_filename)
			}
			private, err := parsePrivateKey(private_filename)
			if err != nil {
				return err
			}
			public, err := parsePublicKey(filename)
			if err != nil {
				return err
			}
			management.cached_keys[private_filename] = KeyPair{Public: public, Private: private}
		}
	}
	return nil
}

func (management *KeyManagement) Insert(public string, private string, name string) error {
	for key, _ := range management.cached_keys {
		if key == name {
			return fmt.Errorf("Inserting key with name %s: already present!", name)
		}
	}
	public_filename := path.Clean(name)
	if strings.HasPrefix(public_filename, "..") {
		return fmt.Errorf("Cannot insert key in parent of Key Management directory")
	}
	public_filename = path.Join(management.underlying_dir, public_filename)
	private_filename := public_filename + ".pub"
	public_file, err := os.Open(public_filename)
	if err != nil {
		return err
	}
	private_file, err := os.Open(private_filename)
	if err != nil {
		return err
	}
	n_read, err := public_file.WriteString(public)
	if err != nil {
		return err
	}
	if n_read < len(public) {
		return fmt.Errorf("Write error: did not write whole public key file")
	}
	if _, err := private_file.WriteString(private); err != nil {
		return err
	}
	if n_read < len(private) {
		return fmt.Errorf("Write error: did not write whole private key file")
	}
	public_file.Close()
	private_file.Close()
	return management.insertInternal(public, private, name)
}

func (management *KeyManagement) insertInternal(public string, private string, name string) error {
	public_key, err := ssh.ParsePublicKey([]byte(public))
	if err != nil {
		return err
	}
	passphrase := utils.GetEnv("SSH_KMS_PASSPHRASE", "")
	var private_key ssh.Signer
	if passphrase == "" {
		private_key, err = ssh.ParsePrivateKey([]byte(private))
	} else {
		private_key, err = ssh.ParsePrivateKeyWithPassphrase([]byte(private), []byte(passphrase))
	}
	management.cached_keys[name] = KeyPair{Public: public_key, Private: private_key}
	return nil
}
